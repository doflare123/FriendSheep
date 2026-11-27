package pushdevices

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"notify_service/internal/database"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestRegistryPostgresReplaysRotatesAndRejectsStaleInvalidation(t *testing.T) {
	db := requirePushDevicePostgres(t)
	registry := newPushDeviceRegistry(t, db)
	installationID, secret := newPushDeviceInstallation(t, db)
	input := validRegisterInput(installationID, secret, 7, testSessionID(1), testOperationID(1), nil, "token-one", "project", "test")

	type result struct {
		value MutationResult
		err   error
	}
	start, results := make(chan struct{}), make(chan result, 2)
	for range 2 {
		go func() {
			<-start
			value, err := registry.Register(context.Background(), input)
			results <- result{value: value, err: err}
		}()
	}
	close(start)
	first, second := <-results, <-results
	created, err := first.value, first.err
	if err != nil || created.State != StateActive || created.Version == "" {
		t.Fatalf("Register() = %#v, %v", created, err)
	}
	if second.err != nil || second.value != created {
		t.Fatalf("параллельный повтор Register() = %#v, %v, ожидался %#v/nil", second.value, second.err, created)
	}
	modifiedReplay := input
	modifiedReplay.ProviderToken = "different-token"
	_, err = registry.Register(context.Background(), modifiedReplay)
	assertPushDomainError(t, err, ErrorCodePushOperationConflict)
	var ciphertext, nonce, fingerprint []byte
	if err := db.QueryRow(`SELECT provider_token_ciphertext, provider_token_nonce, token_fingerprint
		FROM notify_service.push_device_installations WHERE installation_id = $1`, installationID).
		Scan(&ciphertext, &nonce, &fingerprint); err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(ciphertext, []byte(input.ProviderToken)) || len(nonce) != 12 || len(fingerprint) != 32 {
		t.Fatal("адрес провайдера должен храниться зашифрованным с nonce и отпечатком")
	}

	bindings, err := registry.ListActiveBindings(context.Background(), input.UserID)
	if err != nil || len(bindings) != 1 {
		t.Fatalf("ListActiveBindings() = %#v, %v", bindings, err)
	}
	firstBinding := bindings[0]
	rotate := input
	rotate.ExpectedVersion = stringPointer(created.Version)
	rotate.OperationID = testOperationID(2)
	rotate.ProviderToken = "token-two"
	rotated, err := registry.Register(context.Background(), rotate)
	if err != nil || rotated.Version == created.Version {
		t.Fatalf("ротация = %#v, %v", rotated, err)
	}
	bindings, err = registry.ListActiveBindings(context.Background(), input.UserID)
	if err != nil || len(bindings) != 1 || bindings[0].BindingID != firstBinding.BindingID || bindings[0].Version != rotated.Version {
		t.Fatalf("привязка после ротации = %#v, %v", bindings, err)
	}
	address, found, err := registry.ResolveActiveAddress(context.Background(), installationID, firstBinding.BindingID, input.UserID)
	if err != nil || !found || address.ProviderToken != "token-two" {
		t.Fatalf("ResolveActiveAddress() = %#v, %t, %v", address, found, err)
	}
	if applied, invalidateErr := registry.InvalidateAddress(context.Background(), installationID, firstBinding.BindingID, input.UserID, created.Version); invalidateErr != nil || applied {
		t.Fatalf("запоздалая инвалидизация = %t, %v", applied, invalidateErr)
	}
	if applied, invalidateErr := registry.InvalidateAddress(context.Background(), installationID, firstBinding.BindingID, input.UserID, rotated.Version); invalidateErr != nil || !applied {
		t.Fatalf("актуальная инвалидизация = %t, %v", applied, invalidateErr)
	}
	_, found, err = registry.ResolveActiveAddress(context.Background(), installationID, firstBinding.BindingID, input.UserID)
	if err != nil || found {
		t.Fatalf("ResolveActiveAddress() после invalidation = found:%t error:%v", found, err)
	}
}

func TestRegistryPostgresPreventsAccountAFromChangingNewBindingB(t *testing.T) {
	db := requirePushDevicePostgres(t)
	registry := newPushDeviceRegistry(t, db)
	installationID, secret := newPushDeviceInstallation(t, db)
	a := validRegisterInput(installationID, secret, 101, testSessionID(11), testOperationID(11), nil, "token-a", "project", "test")
	first, err := registry.Register(context.Background(), a)
	if err != nil {
		t.Fatal(err)
	}
	oldBindings, err := registry.ListActiveBindings(context.Background(), a.UserID)
	if err != nil || len(oldBindings) != 1 {
		t.Fatalf("bindings A = %#v, %v", oldBindings, err)
	}
	b := a
	b.UserID, b.SessionID, b.OperationID = 202, testSessionID(12), testOperationID(12)
	b.ExpectedVersion, b.ProviderToken = stringPointer(first.Version), "token-b"
	second, err := registry.Register(context.Background(), b)
	if err != nil || second.Version == first.Version {
		t.Fatalf("Register B = %#v, %v", second, err)
	}
	newBindings, err := registry.ListActiveBindings(context.Background(), b.UserID)
	if err != nil || len(newBindings) != 1 || newBindings[0].BindingID == oldBindings[0].BindingID {
		t.Fatalf("bindings B = %#v, %v", newBindings, err)
	}
	staleRegister := a
	staleRegister.OperationID, staleRegister.ExpectedVersion = testOperationID(13), stringPointer(first.Version)
	_, err = registry.Register(context.Background(), staleRegister)
	assertPushDomainError(t, err, ErrorCodePushVersionConflict)
	current, exists := CurrentVersion(err)
	if !exists || current == nil || *current != second.Version {
		t.Fatalf("CurrentVersion = %#v, %t", current, exists)
	}
	staleDetach := DetachInput{InstallationID: installationID, InstallationSecret: secret, SchemaVersion: SchemaVersion, UserID: a.UserID, SessionID: a.SessionID, OperationID: testOperationID(14), ExpectedVersion: stringPointer(second.Version)}
	_, err = registry.Detach(context.Background(), staleDetach)
	assertPushDomainError(t, err, ErrorCodePushDeviceNotFound)
	if applied, invalidateErr := registry.InvalidateAddress(context.Background(), installationID, oldBindings[0].BindingID, a.UserID, first.Version); invalidateErr != nil || applied {
		t.Fatalf("запоздалая инвалидизация A = %t, %v", applied, invalidateErr)
	}
	resolved, found, err := registry.ResolveActiveAddress(context.Background(), installationID, newBindings[0].BindingID, b.UserID)
	if err != nil || !found || resolved.ProviderToken != "token-b" || resolved.Version != second.Version {
		t.Fatalf("привязка B была повреждена: %#v, %t, %v", resolved, found, err)
	}
}

func TestRegistryPostgresTokenScopeConcurrentCASAndDetachReplay(t *testing.T) {
	db := requirePushDevicePostgres(t)
	registry := newPushDeviceRegistry(t, db)
	idA, secretA := newPushDeviceInstallation(t, db)
	idB, secretB := newPushDeviceInstallation(t, db)
	idC, secretC := newPushDeviceInstallation(t, db)
	inA := validRegisterInput(idA, secretA, 1, testSessionID(21), testOperationID(21), nil, "shared-token", "project", "test")
	first, err := registry.Register(context.Background(), inA)
	if err != nil {
		t.Fatal(err)
	}
	inB := validRegisterInput(idB, secretB, 2, testSessionID(22), testOperationID(22), nil, "shared-token", "project", "test")
	_, err = registry.Register(context.Background(), inB)
	assertPushDomainError(t, err, ErrorCodePushTokenConflict)
	inC := validRegisterInput(idC, secretC, 3, testSessionID(23), testOperationID(23), nil, "shared-token", "project", "production")
	inC.Platform = "ios"
	if _, err = registry.Register(context.Background(), inC); err != nil {
		t.Fatalf("тот же адрес в другом scope = %v", err)
	}
	idPhone, secretPhone := newPushDeviceInstallation(t, db)
	inPhone := validRegisterInput(idPhone, secretPhone, 1, testSessionID(41), testOperationID(41), nil, "other-phone-token", "project", "test")
	inPhone.Platform = "ios"
	if _, err = registry.Register(context.Background(), inPhone); err != nil {
		t.Fatalf("вторая установка в том же scope = %v", err)
	}
	bindings, err := registry.ListActiveBindings(context.Background(), 1)
	if err != nil || len(bindings) != 2 {
		t.Fatalf("две установки пользователя = %#v, %v", bindings, err)
	}
	installed := map[string]bool{bindings[0].InstallationID: true, bindings[1].InstallationID: true}
	if !installed[idA] || !installed[idPhone] {
		t.Fatalf("список не содержит обе установки пользователя: %#v", bindings)
	}
	type registerOutcome struct {
		result MutationResult
		err    error
	}
	start, results := make(chan struct{}), make(chan registerOutcome, 2)
	for _, operationID := range []string{testOperationID(24), testOperationID(25)} {
		operationID := operationID
		go func() {
			<-start
			update := validRegisterInput(idA, secretA, 1, testSessionID(21), operationID, stringPointer(first.Version), "token-after-cas", "project", "test")
			result, registerErr := registry.Register(context.Background(), update)
			results <- registerOutcome{result: result, err: registerErr}
		}()
	}
	close(start)
	var successes, conflicts int
	var activeVersion string
	for range 2 {
		outcome := <-results
		if outcome.err == nil {
			successes++
			activeVersion = outcome.result.Version
		} else if ErrorCode(outcome.err) == ErrorCodePushVersionConflict {
			conflicts++
		} else {
			t.Fatalf("конкурентный CAS = %v", outcome.err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("CAS success=%d conflict=%d", successes, conflicts)
	}
	detach := DetachInput{InstallationID: idA, InstallationSecret: secretA, SchemaVersion: SchemaVersion, UserID: 1, SessionID: testSessionID(21), OperationID: testOperationID(26), ExpectedVersion: nil}
	_, err = registry.Detach(context.Background(), detach)
	assertPushDomainError(t, err, ErrorCodePushVersionConflict)
	detach.ExpectedVersion = &activeVersion
	detached, err := registry.Detach(context.Background(), detach)
	if err != nil || detached.State != StateDetached || detached.Version == activeVersion {
		t.Fatalf("отвязка = %#v, %v", detached, err)
	}
	replayed, err := registry.Detach(context.Background(), detach)
	if err != nil || replayed != detached {
		t.Fatalf("повтор отвязки = %#v, %v", replayed, err)
	}
	_, err = registry.Register(context.Background(), inA)
	assertPushDomainError(t, err, ErrorCodePushVersionConflict)
	if _, err := registry.Register(context.Background(), inB); err != nil {
		t.Fatalf("повтор регистрации после освобождения адреса = %v", err)
	}
}

func TestRegistryPostgresCancelWhileWaitingForInstallationLock(t *testing.T) {
	db := requirePushDevicePostgres(t)
	registry := newPushDeviceRegistry(t, db)
	id, secret := newPushDeviceInstallation(t, db)
	ctx, timeout := context.WithTimeout(context.Background(), 10*time.Second)
	defer timeout()
	lockTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer lockTx.Rollback()
	if err := lockInstallation(ctx, lockTx, id); err != nil {
		t.Fatal(err)
	}
	var holderPID int32
	if err := lockTx.QueryRowContext(ctx, `SELECT pg_backend_pid()`).Scan(&holderPID); err != nil {
		t.Fatal(err)
	}
	input := validRegisterInput(id, secret, 33, testSessionID(33), testOperationID(33), nil, "locked-token", "project", "test")
	requestCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	result := make(chan error, 1)
	go func() {
		_, err := registry.Register(requestCtx, input)
		result <- err
	}()

	// Отменяем запрос лишь после того, как PostgreSQL подтвердил ожидание блокировки.
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var blocked bool
		if err := db.QueryRowContext(ctx, `SELECT EXISTS (
			SELECT 1 FROM pg_stat_activity WHERE datname = current_database() AND $1 = ANY(pg_blocking_pids(pid))
		)`, holderPID).Scan(&blocked); err != nil {
			t.Fatal(err)
		}
		if blocked {
			break
		}
		select {
		case err := <-result:
			t.Fatalf("Register завершился до ожидания блокировки: %v", err)
		case <-ctx.Done():
			t.Fatalf("PostgreSQL не подтвердил ожидание блокировки: %v", ctx.Err())
		case <-ticker.C:
		}
	}
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("отменённая регистрация = %v", err)
		}
	case <-ctx.Done():
		t.Fatalf("отмена регистрации не освободила запрос: %v", ctx.Err())
	}
	if err := lockTx.Rollback(); err != nil {
		t.Fatal(err)
	}
	var rows int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM notify_service.push_device_installations WHERE installation_id = $1`, id).Scan(&rows); err != nil || rows != 0 {
		t.Fatalf("после отмены количество строк = %d, %v", rows, err)
	}
	if _, err := registry.Register(ctx, input); err != nil {
		t.Fatalf("повтор регистрации после отмены = %v", err)
	}
}

func TestRegistryPostgresRejectsIncompleteReplayRecord(t *testing.T) {
	db := requirePushDevicePostgres(t)
	registry := newPushDeviceRegistry(t, db)
	id, secret := newPushDeviceInstallation(t, db)
	input := validRegisterInput(id, secret, 34, testSessionID(34), testOperationID(34), nil, "pair-token", "project", "test")
	if _, err := registry.Register(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	_, err := db.Exec(`UPDATE notify_service.push_device_installations SET last_operation_fingerprint = NULL WHERE installation_id = $1`, id)
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23514" || pgErr.ConstraintName != "push_device_installations_operation_fields_check" {
		t.Fatalf("неполная запись о повторе операции = %v", err)
	}
}

func TestRegistryPostgresFirstDetachPreventsDelayedCreate(t *testing.T) {
	db := requirePushDevicePostgres(t)
	registry := newPushDeviceRegistry(t, db)
	id, secret := newPushDeviceInstallation(t, db)
	input := DetachInput{InstallationID: id, InstallationSecret: secret, SchemaVersion: SchemaVersion, UserID: 35, SessionID: testSessionID(35), OperationID: testOperationID(35)}
	detached, err := registry.Detach(context.Background(), input)
	if err != nil || detached.State != StateDetached || detached.Version == "" {
		t.Fatalf("первичная отвязка = %#v, %v", detached, err)
	}
	register := validRegisterInput(id, secret, input.UserID, input.SessionID, testOperationID(36), nil, "delayed-token", "project", "test")
	_, err = registry.Register(context.Background(), register)
	assertPushDomainError(t, err, ErrorCodePushVersionConflict)
	register.ExpectedVersion = &detached.Version
	created, err := registry.Register(context.Background(), register)
	if err != nil || created.State != StateActive || created.Version == detached.Version {
		t.Fatalf("включение после первичной отвязки = %#v, %v", created, err)
	}
}

func TestRegistryPostgresRejectsInvalidatedTokenAndPreservesMarkerAcrossDetach(t *testing.T) {
	db := requirePushDevicePostgres(t)
	registry := newPushDeviceRegistry(t, db)
	id, secret := newPushDeviceInstallation(t, db)
	input := validRegisterInput(id, secret, 36, testSessionID(36), testOperationID(37), nil, "rejected-token", "project", "test")
	created, err := registry.Register(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	bindings, err := registry.ListActiveBindings(context.Background(), input.UserID)
	if err != nil || len(bindings) != 1 {
		t.Fatalf("исходная привязка = %#v, %v", bindings, err)
	}
	binding := bindings[0]
	if applied, err := registry.InvalidateAddress(context.Background(), id, binding.BindingID, input.UserID, created.Version); err != nil || !applied {
		t.Fatalf("InvalidateAddress() = %t, %v", applied, err)
	}
	var rejectedFingerprint []byte
	var invalidVersion string
	if err := db.QueryRow(`SELECT last_rejected_token_fingerprint, version::text
		FROM notify_service.push_device_installations WHERE installation_id = $1`, id).Scan(&rejectedFingerprint, &invalidVersion); err != nil {
		t.Fatal(err)
	}
	if len(rejectedFingerprint) != 32 || !bytes.Equal(rejectedFingerprint, registry.tokenFingerprint(input.ProviderToken)) {
		t.Fatal("инвалидация не сохранила отпечаток отклонённого адреса")
	}

	oldAddress := input
	oldAddress.OperationID = testOperationID(38)
	oldAddress.ExpectedVersion = &invalidVersion
	_, err = registry.Register(context.Background(), oldAddress)
	assertPushDomainError(t, err, ErrorCodePushTokenInvalidated)

	newAddress := oldAddress
	newAddress.OperationID = testOperationID(39)
	newAddress.ProviderToken = "replacement-token"
	active, err := registry.Register(context.Background(), newAddress)
	if err != nil || active.State != StateActive {
		t.Fatalf("регистрация нового адреса = %#v, %v", active, err)
	}
	bindings, err = registry.ListActiveBindings(context.Background(), input.UserID)
	if err != nil || len(bindings) != 1 || bindings[0].BindingID != binding.BindingID || bindings[0].Version != active.Version {
		t.Fatalf("привязка после нового адреса = %#v, %v", bindings, err)
	}
	if applied, err := registry.InvalidateAddress(context.Background(), id, binding.BindingID, input.UserID, invalidVersion); err != nil || applied {
		t.Fatalf("запоздалая инвалидизация не должна менять marker: %t, %v", applied, err)
	}

	detach := DetachInput{InstallationID: id, InstallationSecret: secret, SchemaVersion: SchemaVersion, UserID: input.UserID, SessionID: input.SessionID, OperationID: testOperationID(40), ExpectedVersion: &active.Version}
	detached, err := registry.Detach(context.Background(), detach)
	if err != nil {
		t.Fatal(err)
	}
	var detachedFingerprint []byte
	if err := db.QueryRow(`SELECT last_rejected_token_fingerprint FROM notify_service.push_device_installations WHERE installation_id = $1`, id).Scan(&detachedFingerprint); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(detachedFingerprint, rejectedFingerprint) {
		t.Fatal("отвязка не должна очищать marker отклонённого адреса")
	}
	oldAfterDetach := input
	oldAfterDetach.OperationID = testOperationID(41)
	oldAfterDetach.ExpectedVersion = &detached.Version
	_, err = registry.Register(context.Background(), oldAfterDetach)
	assertPushDomainError(t, err, ErrorCodePushTokenInvalidated)
	rebound := oldAfterDetach
	rebound.OperationID = testOperationID(44)
	rebound.ProviderToken = "rebind-replacement-token"
	activeAfterRebind, err := registry.Register(context.Background(), rebound)
	if err != nil || activeAfterRebind.State != StateActive {
		t.Fatalf("перепривязка с новым адресом = %#v, %v", activeAfterRebind, err)
	}
	var reboundFingerprint []byte
	if err := db.QueryRow(`SELECT last_rejected_token_fingerprint FROM notify_service.push_device_installations WHERE installation_id = $1`, id).Scan(&reboundFingerprint); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(reboundFingerprint, rejectedFingerprint) {
		t.Fatal("перепривязка не должна очищать marker отклонённого адреса")
	}
}

func TestRegistryPostgresAllowsLegacyInvalidAddressWithoutRejectedMarker(t *testing.T) {
	db := requirePushDevicePostgres(t)
	registry := newPushDeviceRegistry(t, db)
	id, secret := newPushDeviceInstallation(t, db)
	input := validRegisterInput(id, secret, 37, testSessionID(37), testOperationID(42), nil, "legacy-token", "project", "test")
	created, err := registry.Register(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	bindings, err := registry.ListActiveBindings(context.Background(), input.UserID)
	if err != nil || len(bindings) != 1 {
		t.Fatalf("исходная legacy-привязка = %#v, %v", bindings, err)
	}
	if applied, err := registry.InvalidateAddress(context.Background(), id, bindings[0].BindingID, input.UserID, created.Version); err != nil || !applied {
		t.Fatalf("InvalidateAddress() = %t, %v", applied, err)
	}
	var invalidVersion string
	if _, err := db.Exec(`UPDATE notify_service.push_device_installations
		SET last_rejected_token_fingerprint = NULL WHERE installation_id = $1`, id); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT version::text FROM notify_service.push_device_installations WHERE installation_id = $1`, id).Scan(&invalidVersion); err != nil {
		t.Fatal(err)
	}
	legacyRetry := input
	legacyRetry.OperationID = testOperationID(43)
	legacyRetry.ExpectedVersion = &invalidVersion
	if _, err := registry.Register(context.Background(), legacyRetry); err != nil {
		t.Fatalf("legacy invalid_address без marker должен принять адрес: %v", err)
	}
}

func requirePushDevicePostgres(t *testing.T) *sql.DB {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("NOTIFY_SERVICE_TEST_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("NOTIFY_SERVICE_TEST_POSTGRES_DSN не задан; PostgreSQL-интеграция реестра устройств пропущена")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := database.Open(ctx, dsn, 5*time.Second)
	if err != nil {
		t.Fatalf("не удалось открыть PostgreSQL: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := database.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate() = %v", err)
	}
	return db
}

func newPushDeviceRegistry(t *testing.T, db *sql.DB) *Registry {
	t.Helper()
	registry, err := NewRegistry(db, CryptoConfig{EncryptionKey: bytesOf(0x10), EncryptionKeyID: "test-key", TokenFingerprintKey: bytesOf(0x20), OperationFingerprintKey: bytesOf(0x30), AllowedScopes: []ProviderScope{{Project: "project", Environment: "test"}, {Project: "project", Environment: "production"}}})
	if err != nil {
		t.Fatal(err)
	}
	return registry
}
func bytesOf(start byte) []byte {
	value := make([]byte, 32)
	for index := range value {
		value[index] = start + byte(index)
	}
	return value
}
func newPushDeviceInstallation(t *testing.T, db *sql.DB) (string, string) {
	t.Helper()
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		t.Fatal(err)
	}
	secret := base64.RawURLEncoding.EncodeToString(value)
	hash := sha256.Sum256(append([]byte(installationDomain), value...))
	id := base64.RawURLEncoding.EncodeToString(hash[:])
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := db.ExecContext(ctx, `DELETE FROM notify_service.push_device_installations WHERE installation_id = $1`, id); err != nil {
			t.Errorf("не удалось удалить тестовую установку: %v", err)
		}
	})
	return id, secret
}
func validRegisterInput(id, secret string, user uint64, session, operation string, expected *string, token, project, environment string) RegisterInput {
	return RegisterInput{InstallationID: id, InstallationSecret: secret, SchemaVersion: SchemaVersion, UserID: user, SessionID: session, OperationID: operation, ExpectedVersion: expected, Platform: "android", Provider: ProviderFCM, ProviderProject: project, Environment: environment, ProviderToken: token, NotificationsEnabled: true}
}
func testSessionID(seed byte) string {
	value := make([]byte, 18)
	for index := range value {
		value[index] = seed + byte(index)
	}
	return base64.RawURLEncoding.EncodeToString(value)
}
func testOperationID(seed uint64) string { return fmt.Sprintf("%08x-0000-4000-8000-%012x", seed, seed) }
func stringPointer(value string) *string { return &value }
func assertPushDomainError(t *testing.T, err error, want string) {
	t.Helper()
	if ErrorCode(err) != want {
		t.Fatalf("ErrorCode(%v) = %q, ожидался %q", err, ErrorCode(err), want)
	}
}
