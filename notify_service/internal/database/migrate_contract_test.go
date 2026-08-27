package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"strings"
	"testing"

	"notify_service/migrations"
)

type migrationExec struct {
	query string
	args  []driver.NamedValue
}

type migrationDriverState struct {
	execs           []migrationExec
	queries         []string
	applied         bool
	appliedName     string
	appliedChecksum string
	committed       bool
	rolledBack      bool
}

type migrationConnector struct {
	state *migrationDriverState
}

func (connector *migrationConnector) Connect(context.Context) (driver.Conn, error) {
	return &migrationConn{state: connector.state}, nil
}

func (connector *migrationConnector) Driver() driver.Driver {
	return migrationDriver{state: connector.state}
}

type migrationDriver struct {
	state *migrationDriverState
}

func (databaseDriver migrationDriver) Open(string) (driver.Conn, error) {
	return &migrationConn{state: databaseDriver.state}, nil
}

type migrationConn struct {
	state *migrationDriverState
}

func (*migrationConn) Prepare(string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (*migrationConn) Close() error                        { return nil }
func (conn *migrationConn) Begin() (driver.Tx, error)      { return &migrationTx{state: conn.state}, nil }
func (conn *migrationConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return &migrationTx{state: conn.state}, nil
}

func (conn *migrationConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	conn.state.execs = append(conn.state.execs, migrationExec{query: query, args: append([]driver.NamedValue(nil), args...)})
	return driver.RowsAffected(1), nil
}

func (conn *migrationConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	conn.state.queries = append(conn.state.queries, query)
	rows := &migrationRows{}
	if conn.state.applied {
		rows.values = [][]driver.Value{{conn.state.appliedName, conn.state.appliedChecksum}}
	}
	return rows, nil
}

type migrationTx struct {
	state *migrationDriverState
}

func (tx *migrationTx) Commit() error {
	tx.state.committed = true
	return nil
}

func (tx *migrationTx) Rollback() error {
	tx.state.rolledBack = true
	return nil
}

type migrationRows struct {
	values [][]driver.Value
	index  int
}

func (*migrationRows) Columns() []string { return []string{"name", "checksum"} }
func (*migrationRows) Close() error      { return nil }
func (rows *migrationRows) Next(destination []driver.Value) error {
	if rows.index >= len(rows.values) {
		return io.EOF
	}
	copy(destination, rows.values[rows.index])
	rows.index++
	return nil
}

func TestMigrateStoresChecksumInHistory(t *testing.T) {
	t.Parallel()

	state := &migrationDriverState{}
	db := sql.OpenDB(&migrationConnector{state: state})
	defer db.Close()

	if err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("Migrate() завершился ошибкой: %v", err)
	}
	if !state.committed {
		t.Fatal("транзакция миграций не была зафиксирована")
	}

	var historyHasChecksum bool
	insertedChecksums := make(map[int64]string)
	for _, execution := range state.execs {
		lowerQuery := strings.ToLower(execution.query)
		if strings.Contains(lowerQuery, "schema_migrations") && strings.Contains(lowerQuery, "checksum text not null") {
			historyHasChecksum = true
		}
		if strings.Contains(lowerQuery, "insert into notify_service.schema_migrations") {
			if len(execution.args) != 3 {
				t.Fatalf("запись истории содержит %d аргументов, ожидалось 3", len(execution.args))
			}
			version, _ := execution.args[0].Value.(int64)
			checksum, _ := execution.args[2].Value.(string)
			insertedChecksums[version] = checksum
		}
	}
	if !historyHasChecksum {
		t.Fatal("таблица истории миграций не требует обязательную контрольную сумму")
	}
	for version, name := range map[int64]string{
		1: "000001_initial_schema.sql",
		2: "000002_event_lifecycle_scheduler.sql",
	} {
		body, err := migrations.Files.ReadFile(name)
		if err != nil {
			t.Fatalf("не удалось прочитать встроенную миграцию: %v", err)
		}
		wantChecksum := fmt.Sprintf("%x", sha256.Sum256(body))
		if insertedChecksums[version] != wantChecksum {
			t.Errorf("контрольная сумма версии %d = %q, ожидалась %q", version, insertedChecksums[version], wantChecksum)
		}
	}
	if len(state.queries) == 0 || !strings.Contains(strings.ToLower(state.queries[0]), "select name, checksum") {
		t.Fatal("исполнитель миграций не проверяет имя и контрольную сумму ранее применённой версии")
	}
}

func TestMigrateRejectsChangedAppliedMigration(t *testing.T) {
	t.Parallel()

	state := &migrationDriverState{
		applied:         true,
		appliedName:     "000001_initial_schema.sql",
		appliedChecksum: "izmenennaya-istoriya",
	}
	db := sql.OpenDB(&migrationConnector{state: state})
	defer db.Close()

	err := Migrate(context.Background(), db)
	if err == nil {
		t.Fatal("Migrate() не отклонил изменённую применённую миграцию")
	}
	if !strings.Contains(err.Error(), "не совпадает со встроенной историей") {
		t.Fatalf("Migrate() вернул нестабильную ошибку нарушения истории: %v", err)
	}
	if state.committed {
		t.Fatal("транзакция с изменённой историей миграций была зафиксирована")
	}
	if !state.rolledBack {
		t.Fatal("транзакция с изменённой историей миграций не была отменена")
	}
	for _, execution := range state.execs {
		if strings.Contains(strings.ToLower(execution.query), "insert into notify_service.schema_migrations") {
			t.Fatal("изменённая применённая миграция была повторно записана")
		}
	}
}
