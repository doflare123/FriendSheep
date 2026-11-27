package pushdevices

import (
	"context"
	"crypto/hmac"
	"database/sql"
	"errors"
	"math"

	"github.com/jackc/pgx/v5/pgconn"
)

const installationColumns = `
	installation_id,
	state,
	version::text,
	user_id,
	COALESCE(session_id, ''),
	COALESCE(binding_id::text, ''),
	COALESCE(platform, ''),
	COALESCE(provider, ''),
	COALESCE(provider_project, ''),
	COALESCE(environment, ''),
	COALESCE(provider_token_ciphertext, ''::bytea),
	COALESCE(provider_token_nonce, ''::bytea),
	COALESCE(provider_token_key_id, ''),
	COALESCE(token_fingerprint, ''::bytea),
	COALESCE(last_rejected_token_fingerprint, ''::bytea),
	COALESCE(last_operation_id::text, ''),
	COALESCE(last_operation_fingerprint, ''::bytea)`

type installationRecord struct {
	InstallationID     string
	State              string
	Version            string
	UserID             sql.NullInt64
	SessionID          string
	BindingID          string
	Platform           string
	Provider           string
	ProviderProject    string
	Environment        string
	TokenCiphertext    []byte
	TokenNonce         []byte
	TokenKeyID         string
	TokenFingerprint   []byte
	RejectedTokenPrint []byte
	LastOperationID    string
	LastOperationPrint []byte
}

func (r *Registry) Register(ctx context.Context, input RegisterInput) (MutationResult, error) {
	if err := r.validateRegister(input); err != nil {
		return MutationResult{}, err
	}
	operationFingerprint, err := r.fingerprintRegister(input)
	if err != nil {
		return MutationResult{}, err
	}
	tokenFingerprint := r.tokenFingerprint(input.ProviderToken)
	ciphertext, nonce, err := r.encryptProviderToken(input)
	if err != nil {
		return MutationResult{}, err
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return MutationResult{}, wrapStoreError("не удалось начать регистрацию push-устройства", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := lockInstallation(ctx, tx, input.InstallationID); err != nil {
		return MutationResult{}, err
	}
	record, found, err := loadInstallationForUpdate(ctx, tx, input.InstallationID)
	if err != nil {
		return MutationResult{}, err
	}
	if found && record.LastOperationID == input.OperationID {
		if !hmac.Equal(record.LastOperationPrint, operationFingerprint) {
			return MutationResult{}, domainError(ErrorCodePushOperationConflict, nil)
		}
		return mutationResult(record), commitReplay(tx)
	}
	if !expectedVersionMatches(found, record.Version, input.ExpectedVersion) {
		return MutationResult{}, versionConflict(currentVersion(found, record.Version))
	}
	if found && record.Platform != "" && (record.Platform != input.Platform ||
		record.Provider != input.Provider || record.ProviderProject != input.ProviderProject ||
		record.Environment != input.Environment) {
		return MutationResult{}, domainError(ErrorCodePushInstallationConflict, nil)
	}
	if found && len(record.RejectedTokenPrint) > 0 && hmac.Equal(record.RejectedTokenPrint, tokenFingerprint) {
		return MutationResult{}, domainError(ErrorCodePushTokenInvalidated, nil)
	}

	version, err := newUUIDv4()
	if err != nil {
		return MutationResult{}, wrapStoreError("не удалось создать версию привязки", err)
	}
	bindingID := record.BindingID
	if !found || record.State == StateDetached || !record.UserID.Valid || uint64(record.UserID.Int64) != input.UserID || record.SessionID != input.SessionID {
		bindingID, err = newUUIDv4()
		if err != nil {
			return MutationResult{}, wrapStoreError("не удалось создать поколение привязки", err)
		}
	}

	if found {
		err = updateRegistration(ctx, tx, input, version, bindingID, ciphertext, nonce,
			r.encryptionKeyID, tokenFingerprint, operationFingerprint)
	} else {
		err = insertRegistration(ctx, tx, input, version, bindingID, ciphertext, nonce,
			r.encryptionKeyID, tokenFingerprint, operationFingerprint)
	}
	if err != nil {
		if isProviderTokenConflict(err) {
			return MutationResult{}, domainError(ErrorCodePushTokenConflict, nil)
		}
		return MutationResult{}, wrapStoreError("не удалось сохранить регистрацию push-устройства", err)
	}
	if err := tx.Commit(); err != nil {
		return MutationResult{}, wrapStoreError("не удалось зафиксировать регистрацию push-устройства", err)
	}
	return MutationResult{SchemaVersion: SchemaVersion, InstallationID: input.InstallationID, Version: version, State: StateActive}, nil
}

func (r *Registry) Detach(ctx context.Context, input DetachInput) (MutationResult, error) {
	if err := validateDetach(input); err != nil {
		return MutationResult{}, err
	}
	operationFingerprint, err := r.fingerprintDetach(input)
	if err != nil {
		return MutationResult{}, err
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return MutationResult{}, wrapStoreError("не удалось начать отвязку push-устройства", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := lockInstallation(ctx, tx, input.InstallationID); err != nil {
		return MutationResult{}, err
	}
	record, found, err := loadInstallationForUpdate(ctx, tx, input.InstallationID)
	if err != nil {
		return MutationResult{}, err
	}
	if found && record.LastOperationID == input.OperationID {
		if !hmac.Equal(record.LastOperationPrint, operationFingerprint) {
			return MutationResult{}, domainError(ErrorCodePushOperationConflict, nil)
		}
		return mutationResult(record), commitReplay(tx)
	}
	if !expectedVersionMatches(found, record.Version, input.ExpectedVersion) {
		return MutationResult{}, versionConflict(currentVersion(found, record.Version))
	}
	if found && (record.State == StateActive || record.State == StateInvalidAddress) &&
		(!record.UserID.Valid || uint64(record.UserID.Int64) != input.UserID || record.SessionID != input.SessionID) {
		return MutationResult{}, domainError(ErrorCodePushDeviceNotFound, nil)
	}

	version, err := newUUIDv4()
	if err != nil {
		return MutationResult{}, wrapStoreError("не удалось создать версию отвязки", err)
	}
	if found {
		_, err = tx.ExecContext(ctx, `
			UPDATE notify_service.push_device_installations
			SET state = $2,
			    version = $3::uuid,
			    user_id = NULL,
			    session_id = NULL,
			    binding_id = NULL,
			    provider_token_ciphertext = NULL,
			    provider_token_nonce = NULL,
			    provider_token_key_id = NULL,
			    token_fingerprint = NULL,
			    last_operation_id = $4::uuid,
			    last_operation_fingerprint = $5,
			    updated_at = NOW()
			WHERE installation_id = $1`,
			input.InstallationID, StateDetached, version, input.OperationID, operationFingerprint)
	} else {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO notify_service.push_device_installations (
				installation_id, state, version, last_operation_id, last_operation_fingerprint
			) VALUES ($1, $2, $3::uuid, $4::uuid, $5)`,
			input.InstallationID, StateDetached, version, input.OperationID, operationFingerprint)
	}
	if err != nil {
		return MutationResult{}, wrapStoreError("не удалось сохранить отвязку push-устройства", err)
	}
	if err := tx.Commit(); err != nil {
		return MutationResult{}, wrapStoreError("не удалось зафиксировать отвязку push-устройства", err)
	}
	return MutationResult{SchemaVersion: SchemaVersion, InstallationID: input.InstallationID, Version: version, State: StateDetached}, nil
}

func (r *Registry) ListActiveBindings(ctx context.Context, userID uint64) ([]Binding, error) {
	if userID == 0 || userID > math.MaxInt64 {
		return nil, domainError(ErrorCodeInvalidRequest, errors.New("некорректный userId"))
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT installation_id, binding_id::text, user_id, session_id, version::text
		FROM notify_service.push_device_installations
		WHERE state = 'active' AND user_id = $1
		ORDER BY installation_id`, int64(userID))
	if err != nil {
		return nil, wrapStoreError("не удалось прочитать действующие push-привязки", err)
	}
	defer rows.Close()

	bindings := make([]Binding, 0)
	for rows.Next() {
		var binding Binding
		var storedUserID int64
		if err := rows.Scan(&binding.InstallationID, &binding.BindingID, &storedUserID, &binding.SessionID, &binding.Version); err != nil {
			return nil, wrapStoreError("не удалось разобрать действующую push-привязку", err)
		}
		binding.UserID = uint64(storedUserID)
		bindings = append(bindings, binding)
	}
	if err := rows.Err(); err != nil {
		return nil, wrapStoreError("не удалось завершить чтение push-привязок", err)
	}
	return bindings, nil
}

// ListBindingsForReconciliation возвращает страницу привязок, для которых
// требуется сверка с источником сессий. Курсор является installationId последней
// уже обработанной строки.
func (r *Registry) ListBindingsForReconciliation(ctx context.Context, afterInstallationID string, limit int) ([]Binding, error) {
	if (afterInstallationID != "" && !installationIDPattern.MatchString(afterInstallationID)) || limit < 1 || limit > MaxReconciliationPageSize {
		return nil, domainError(ErrorCodeInvalidRequest, errors.New("некорректная страница сверки push-привязок"))
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT installation_id, binding_id::text, user_id, session_id, version::text
		FROM notify_service.push_device_installations
		WHERE state IN ($1, $2)
		  AND installation_id > $3
		ORDER BY installation_id
		LIMIT $4`, StateActive, StateInvalidAddress, afterInstallationID, limit)
	if err != nil {
		return nil, wrapStoreError("не удалось прочитать push-привязки для сверки", err)
	}
	defer rows.Close()

	bindings := make([]Binding, 0, limit)
	for rows.Next() {
		var binding Binding
		var storedUserID int64
		if err := rows.Scan(&binding.InstallationID, &binding.BindingID, &storedUserID, &binding.SessionID, &binding.Version); err != nil {
			return nil, wrapStoreError("не удалось разобрать push-привязку для сверки", err)
		}
		binding.UserID = uint64(storedUserID)
		bindings = append(bindings, binding)
	}
	if err := rows.Err(); err != nil {
		return nil, wrapStoreError("не удалось завершить чтение push-привязок для сверки", err)
	}
	return bindings, nil
}

// DetachIfCurrent отвязывает только всё ещё совпадающие поколение, пару
// пользователь/сессия и версию. Отпечаток последнего отклонённого адреса
// сохраняется, чтобы прежний адрес нельзя было снова активировать.
func (r *Registry) DetachIfCurrent(ctx context.Context, binding Binding) (bool, error) {
	if !installationIDPattern.MatchString(binding.InstallationID) || !validateUUIDv4(binding.BindingID) ||
		binding.UserID == 0 || binding.UserID > math.MaxInt64 || !validateSessionID(binding.SessionID) ||
		!validateUUIDv4(binding.Version) {
		return false, domainError(ErrorCodeInvalidRequest, errors.New("некорректная условная отвязка push-привязки"))
	}
	newVersion, err := newUUIDv4()
	if err != nil {
		return false, wrapStoreError("не удалось создать версию условной отвязки", err)
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE notify_service.push_device_installations
		SET state = $7,
		    version = $6::uuid,
		    user_id = NULL,
		    session_id = NULL,
		    binding_id = NULL,
		    provider_token_ciphertext = NULL,
		    provider_token_nonce = NULL,
		    provider_token_key_id = NULL,
		    token_fingerprint = NULL,
		    last_operation_id = NULL,
		    last_operation_fingerprint = NULL,
		    updated_at = NOW()
		WHERE installation_id = $1
		  AND binding_id = $2::uuid
		  AND user_id = $3
		  AND session_id = $4
		  AND version = $5::uuid
		  AND state IN ($8, $9)`,
		binding.InstallationID, binding.BindingID, int64(binding.UserID), binding.SessionID,
		binding.Version, newVersion, StateDetached, StateActive, StateInvalidAddress)
	if err != nil {
		return false, wrapStoreError("не удалось условно отвязать push-привязку", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, wrapStoreError("не удалось проверить условную отвязку push-привязки", err)
	}
	return rowsAffected == 1, nil
}

func (r *Registry) ResolveActiveAddress(ctx context.Context, installationID, bindingID string, userID uint64) (ResolvedAddress, bool, error) {
	if !installationIDPattern.MatchString(installationID) || !validateUUIDv4(bindingID) || userID == 0 || userID > math.MaxInt64 {
		return ResolvedAddress{}, false, domainError(ErrorCodeInvalidRequest, nil)
	}
	row := r.db.QueryRowContext(ctx, `SELECT `+installationColumns+`
		FROM notify_service.push_device_installations
		WHERE installation_id = $1 AND binding_id = $2::uuid AND user_id = $3 AND state = $4`,
		installationID, bindingID, int64(userID), StateActive)
	record, found, err := scanInstallation(row)
	if err != nil || !found {
		return ResolvedAddress{}, false, err
	}
	token, err := r.decryptProviderToken(record)
	if err != nil {
		return ResolvedAddress{}, false, wrapStoreError("не удалось разрешить адрес push-устройства", err)
	}
	return ResolvedAddress{
		Binding: Binding{InstallationID: record.InstallationID, BindingID: record.BindingID,
			UserID: uint64(record.UserID.Int64), SessionID: record.SessionID, Version: record.Version},
		Platform: record.Platform, Provider: record.Provider, ProviderProject: record.ProviderProject,
		Environment: record.Environment, ProviderToken: token,
	}, true, nil
}

func (r *Registry) InvalidateAddress(ctx context.Context, installationID, bindingID string, userID uint64, version string) (bool, error) {
	if !installationIDPattern.MatchString(installationID) || !validateUUIDv4(bindingID) ||
		!validateUUIDv4(version) || userID == 0 || userID > math.MaxInt64 {
		return false, domainError(ErrorCodeInvalidRequest, nil)
	}
	newVersion, err := newUUIDv4()
	if err != nil {
		return false, wrapStoreError("не удалось создать версию инвалидизации адреса", err)
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE notify_service.push_device_installations
		SET state = $6,
		    version = $5::uuid,
		    last_rejected_token_fingerprint = token_fingerprint,
		    provider_token_ciphertext = NULL,
		    provider_token_nonce = NULL,
		    provider_token_key_id = NULL,
		    token_fingerprint = NULL,
		    last_operation_id = NULL,
		    last_operation_fingerprint = NULL,
		    updated_at = NOW()
		WHERE installation_id = $1
		  AND binding_id = $2::uuid
		  AND user_id = $3
		  AND version = $4::uuid
		  AND state = $7`,
		installationID, bindingID, int64(userID), version, newVersion, StateInvalidAddress, StateActive)
	if err != nil {
		return false, wrapStoreError("не удалось инвалидировать адрес push-устройства", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, wrapStoreError("не удалось проверить инвалидизацию адреса push-устройства", err)
	}
	return rowsAffected == 1, nil
}

func (r *Registry) validateRegister(input RegisterInput) error {
	if err := verifyInstallationOwnership(input.InstallationID, input.InstallationSecret); err != nil {
		return err
	}
	if input.SchemaVersion != SchemaVersion {
		return domainError(ErrorCodeUnsupportedSchemaVersion, nil)
	}
	if input.UserID == 0 || input.UserID > math.MaxInt64 || !validateSessionID(input.SessionID) ||
		!validateUUIDv4(input.OperationID) || !validateExpectedVersion(input.ExpectedVersion) {
		return domainError(ErrorCodeInvalidRequest, nil)
	}
	if !input.NotificationsEnabled {
		return domainError(ErrorCodePushOptInRequired, nil)
	}
	if input.Platform != "android" && input.Platform != "ios" {
		return domainError(ErrorCodeInvalidRequest, errors.New("некорректная платформа"))
	}
	if input.Provider != ProviderFCM || !scopePattern.MatchString(input.ProviderProject) ||
		!scopePattern.MatchString(input.Environment) {
		return domainError(ErrorCodeInvalidProviderScope, nil)
	}
	if _, allowed := r.allowedScopes[ProviderScope{Project: input.ProviderProject, Environment: input.Environment}]; !allowed {
		return domainError(ErrorCodeInvalidProviderScope, nil)
	}
	if !validateProviderToken(input.ProviderToken) {
		return domainError(ErrorCodeInvalidRequest, errors.New("некорректный адрес провайдера"))
	}
	return nil
}

func validateDetach(input DetachInput) error {
	if err := verifyInstallationOwnership(input.InstallationID, input.InstallationSecret); err != nil {
		return err
	}
	if input.SchemaVersion != SchemaVersion {
		return domainError(ErrorCodeUnsupportedSchemaVersion, nil)
	}
	if input.UserID == 0 || input.UserID > math.MaxInt64 || !validateSessionID(input.SessionID) ||
		!validateUUIDv4(input.OperationID) || !validateExpectedVersion(input.ExpectedVersion) {
		return domainError(ErrorCodeInvalidRequest, nil)
	}
	return nil
}

func lockInstallation(ctx context.Context, tx *sql.Tx, installationID string) error {
	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1, 0))", installationID); err != nil {
		return wrapStoreError("не удалось заблокировать push-устройство", err)
	}
	return nil
}

func loadInstallationForUpdate(ctx context.Context, tx *sql.Tx, installationID string) (installationRecord, bool, error) {
	row := tx.QueryRowContext(ctx, `SELECT `+installationColumns+`
		FROM notify_service.push_device_installations
		WHERE installation_id = $1
		FOR UPDATE`, installationID)
	return scanInstallation(row)
}

func scanInstallation(row interface{ Scan(...any) error }) (installationRecord, bool, error) {
	var record installationRecord
	err := row.Scan(
		&record.InstallationID, &record.State, &record.Version, &record.UserID,
		&record.SessionID, &record.BindingID, &record.Platform, &record.Provider,
		&record.ProviderProject, &record.Environment, &record.TokenCiphertext,
		&record.TokenNonce, &record.TokenKeyID, &record.TokenFingerprint,
		&record.RejectedTokenPrint,
		&record.LastOperationID, &record.LastOperationPrint,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return installationRecord{}, false, nil
	}
	if err != nil {
		return installationRecord{}, false, wrapStoreError("не удалось прочитать push-устройство", err)
	}
	return record, true, nil
}

func insertRegistration(ctx context.Context, tx *sql.Tx, input RegisterInput, version, bindingID string, ciphertext, nonce []byte, keyID string, tokenFingerprint, operationFingerprint []byte) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO notify_service.push_device_installations (
			installation_id, state, version, user_id, session_id, binding_id,
			platform, provider, provider_project, environment,
			provider_token_ciphertext, provider_token_nonce, provider_token_key_id, token_fingerprint,
			last_operation_id, last_operation_fingerprint
		) VALUES ($1, $2, $3::uuid, $4, $5, $6::uuid, $7, $8, $9, $10, $11, $12, $13, $14, $15::uuid, $16)`,
		input.InstallationID, StateActive, version, int64(input.UserID), input.SessionID, bindingID,
		input.Platform, input.Provider, input.ProviderProject, input.Environment,
		ciphertext, nonce, keyID, tokenFingerprint, input.OperationID, operationFingerprint)
	return err
}

func updateRegistration(ctx context.Context, tx *sql.Tx, input RegisterInput, version, bindingID string, ciphertext, nonce []byte, keyID string, tokenFingerprint, operationFingerprint []byte) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE notify_service.push_device_installations
		SET state = $2,
		    version = $3::uuid,
		    user_id = $4,
		    session_id = $5,
		    binding_id = $6::uuid,
		    platform = $7,
		    provider = $8,
		    provider_project = $9,
		    environment = $10,
		    provider_token_ciphertext = $11,
		    provider_token_nonce = $12,
		    provider_token_key_id = $13,
		    token_fingerprint = $14,
		    last_operation_id = $15::uuid,
		    last_operation_fingerprint = $16,
		    updated_at = NOW()
		WHERE installation_id = $1`,
		input.InstallationID, StateActive, version, int64(input.UserID), input.SessionID, bindingID,
		input.Platform, input.Provider, input.ProviderProject, input.Environment,
		ciphertext, nonce, keyID, tokenFingerprint, input.OperationID, operationFingerprint)
	return err
}

func expectedVersionMatches(found bool, version string, expected *string) bool {
	if !found {
		return expected == nil
	}
	return expected != nil && *expected == version
}

func currentVersion(found bool, version string) *string {
	if !found {
		return nil
	}
	value := version
	return &value
}

func mutationResult(record installationRecord) MutationResult {
	return MutationResult{SchemaVersion: SchemaVersion, InstallationID: record.InstallationID, Version: record.Version, State: record.State}
}

func commitReplay(tx *sql.Tx) error {
	if err := tx.Commit(); err != nil {
		return wrapStoreError("не удалось завершить повтор операции push-устройства", err)
	}
	return nil
}

func isProviderTokenConflict(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505" &&
		postgresError.ConstraintName == "push_device_installations_unique_active_token"
}
