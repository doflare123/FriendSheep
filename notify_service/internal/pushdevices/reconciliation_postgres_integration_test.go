package pushdevices

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
)

func TestRegistryPostgresReconciliationPagesAndConditionallyDetaches(t *testing.T) {
	db := requirePushDevicePostgres(t)
	registry := newPushDeviceRegistry(t, db)

	firstID, firstSecret := newPushDeviceInstallation(t, db)
	firstInput := validRegisterInput(firstID, firstSecret, 701, testSessionID(81), testOperationID(81), nil, "reconcile-token-one", "project", "test")
	_, err := registry.Register(context.Background(), firstInput)
	if err != nil {
		t.Fatal(err)
	}
	firstBindings, err := registry.ListActiveBindings(context.Background(), firstInput.UserID)
	if err != nil || len(firstBindings) != 1 {
		t.Fatalf("первая привязка = %#v, %v", firstBindings, err)
	}

	secondID, secondSecret := newPushDeviceInstallation(t, db)
	secondInput := validRegisterInput(secondID, secondSecret, 702, testSessionID(82), testOperationID(82), nil, "reconcile-token-two", "project", "test")
	secondResult, err := registry.Register(context.Background(), secondInput)
	if err != nil {
		t.Fatal(err)
	}
	secondBindings, err := registry.ListActiveBindings(context.Background(), secondInput.UserID)
	if err != nil || len(secondBindings) != 1 {
		t.Fatalf("вторая привязка = %#v, %v", secondBindings, err)
	}
	if applied, err := registry.InvalidateAddress(context.Background(), secondID, secondBindings[0].BindingID, secondInput.UserID, secondResult.Version); err != nil || !applied {
		t.Fatalf("InvalidateAddress() = %t, %v", applied, err)
	}
	var rejectedBefore []byte
	if err := db.QueryRow(`SELECT last_rejected_token_fingerprint
		FROM notify_service.push_device_installations WHERE installation_id = $1`, secondID).Scan(&rejectedBefore); err != nil {
		t.Fatal(err)
	}

	page, err := registry.ListBindingsForReconciliation(context.Background(), "", 1)
	if err != nil || len(page) != 1 {
		t.Fatalf("первая страница = %#v, %v", page, err)
	}
	next, err := registry.ListBindingsForReconciliation(context.Background(), page[0].InstallationID, 1)
	if err != nil || len(next) != 1 || next[0].InstallationID <= page[0].InstallationID {
		t.Fatalf("вторая страница = %#v, %v", next, err)
	}

	stale := firstBindings[0]
	stale.Version = testOperationID(999)
	if applied, err := registry.DetachIfCurrent(context.Background(), stale); err != nil || applied {
		t.Fatalf("устаревшая DetachIfCurrent() = %t, %v", applied, err)
	}
	wrongSession := firstBindings[0]
	wrongSession.SessionID = testSessionID(99)
	if applied, err := registry.DetachIfCurrent(context.Background(), wrongSession); err != nil || applied {
		t.Fatalf("DetachIfCurrent() для другой сессии = %t, %v", applied, err)
	}
	wrongUser := firstBindings[0]
	wrongUser.UserID++
	if applied, err := registry.DetachIfCurrent(context.Background(), wrongUser); err != nil || applied {
		t.Fatalf("DetachIfCurrent() для другого пользователя = %t, %v", applied, err)
	}
	if applied, err := registry.DetachIfCurrent(context.Background(), firstBindings[0]); err != nil || !applied {
		t.Fatalf("актуальная active DetachIfCurrent() = %t, %v", applied, err)
	}
	if applied, err := registry.DetachIfCurrent(context.Background(), secondBindings[0]); err != nil || applied {
		t.Fatalf("DetachIfCurrent() со старой версией invalid_address = %t, %v", applied, err)
	}
	secondBindings[0].Version = currentInstallationVersion(t, db, secondID)
	if applied, err := registry.DetachIfCurrent(context.Background(), secondBindings[0]); err != nil || !applied {
		t.Fatalf("актуальная invalid_address DetachIfCurrent() = %t, %v", applied, err)
	}
	var state string
	var rejectedAfter []byte
	if err := db.QueryRow(`SELECT state, last_rejected_token_fingerprint
		FROM notify_service.push_device_installations WHERE installation_id = $1`, secondID).Scan(&state, &rejectedAfter); err != nil {
		t.Fatal(err)
	}
	if state != StateDetached || !bytes.Equal(rejectedBefore, rejectedAfter) {
		t.Fatalf("после отвязки state=%q, rejected fingerprint preserved=%t", state, bytes.Equal(rejectedBefore, rejectedAfter))
	}
}

func currentInstallationVersion(t *testing.T, db *sql.DB, installationID string) string {
	t.Helper()
	var version string
	if err := db.QueryRow(`SELECT version::text FROM notify_service.push_device_installations WHERE installation_id = $1`, installationID).Scan(&version); err != nil {
		t.Fatal(err)
	}
	return version
}
