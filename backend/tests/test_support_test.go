package tests

import (
	"context"
	"database/sql"

	"friendship/logger"
	"friendship/repository"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ repository.PostgresRepository = (*testPostgresRepository)(nil)
var _ logger.Logger = (*testLogger)(nil)

type testPostgresRepository struct {
	db *gorm.DB
}

func (r *testPostgresRepository) Model(value interface{}) *gorm.DB {
	return r.db.Model(value)
}

func (r *testPostgresRepository) Select(query interface{}, args ...interface{}) *gorm.DB {
	return r.db.Select(query, args...)
}

func (r *testPostgresRepository) Find(out interface{}, where ...interface{}) *gorm.DB {
	return r.db.Find(out, where...)
}

func (r *testPostgresRepository) Exec(sql string, values ...interface{}) *gorm.DB {
	return r.db.Exec(sql, values...)
}

func (r *testPostgresRepository) First(out interface{}, where ...interface{}) *gorm.DB {
	return r.db.First(out, where...)
}

func (r *testPostgresRepository) Raw(sql string, values ...interface{}) *gorm.DB {
	return r.db.Raw(sql, values...)
}

func (r *testPostgresRepository) Create(value interface{}) *gorm.DB {
	return r.db.Create(value)
}

func (r *testPostgresRepository) Save(value interface{}) *gorm.DB {
	return r.db.Save(value)
}

func (r *testPostgresRepository) Updates(value interface{}) *gorm.DB {
	return r.db.Updates(value)
}

func (r *testPostgresRepository) Delete(value interface{}) *gorm.DB {
	return r.db.Delete(value)
}

func (r *testPostgresRepository) Where(query interface{}, args ...interface{}) *gorm.DB {
	return r.db.Where(query, args...)
}

func (r *testPostgresRepository) Preload(column string, conditions ...interface{}) *gorm.DB {
	return r.db.Preload(column, conditions...)
}

func (r *testPostgresRepository) Scopes(funcs ...func(*gorm.DB) *gorm.DB) *gorm.DB {
	return r.db.Scopes(funcs...)
}

func (r *testPostgresRepository) ScanRows(rows *sql.Rows, result interface{}) error {
	return r.db.ScanRows(rows, result)
}

func (r *testPostgresRepository) Transaction(fc func(tx repository.PostgresRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fc(&testPostgresRepository{db: tx})
	})
}

func (r *testPostgresRepository) TransactionWithContext(ctx context.Context, fc func(tx repository.PostgresRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fc(&testPostgresRepository{db: tx})
	})
}

func (r *testPostgresRepository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (r *testPostgresRepository) DropTableIfExists(value interface{}) error {
	return r.db.Migrator().DropTable(value)
}

func (r *testPostgresRepository) GetSQLDB() (*sql.DB, error) {
	return r.db.DB()
}

func (r *testPostgresRepository) Clauses(conds ...clause.Expression) *gorm.DB {
	return r.db.Clauses(conds...)
}

func (r *testPostgresRepository) AutoMigrate(value interface{}) error {
	return r.db.AutoMigrate(value)
}

func (r *testPostgresRepository) Order(value interface{}) *gorm.DB {
	return r.db.Order(value)
}

func (r *testPostgresRepository) Limit(limit int) *gorm.DB {
	return r.db.Limit(limit)
}

func (r *testPostgresRepository) Count(count *int64) *gorm.DB {
	return r.db.Count(count)
}

func (r *testPostgresRepository) Association(column string) *gorm.Association {
	return r.db.Association(column)
}

type testLogger struct{}

func (l *testLogger) Info(msg string, fields ...interface{})  {}
func (l *testLogger) Error(msg string, fields ...interface{}) {}
func (l *testLogger) Debug(msg string, fields ...interface{}) {}
func (l *testLogger) Warn(msg string, fields ...interface{})  {}
func (l *testLogger) Fatal(msg string, fields ...interface{}) {}
func (l *testLogger) Panic(msg string, fields ...interface{}) {}
