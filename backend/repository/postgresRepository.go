package repository

import (
	"database/sql"
	"fmt"
	"friendship/config"
	"friendship/logger"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostgresRepository interface {
	Model(value interface{}) *gorm.DB
	Select(query interface{}, args ...interface{}) *gorm.DB
	Find(out interface{}, where ...interface{}) *gorm.DB
	Exec(sql string, values ...interface{}) *gorm.DB
	First(out interface{}, where ...interface{}) *gorm.DB
	Raw(sql string, values ...interface{}) *gorm.DB
	Create(value interface{}) *gorm.DB
	Save(value interface{}) *gorm.DB
	Updates(value interface{}) *gorm.DB
	Delete(value interface{}) *gorm.DB
	Where(query interface{}, args ...interface{}) *gorm.DB
	Preload(column string, conditions ...interface{}) *gorm.DB
	Scopes(funcs ...func(*gorm.DB) *gorm.DB) *gorm.DB
	ScanRows(rows *sql.Rows, result interface{}) error
	Transaction(fc func(tx PostgresRepository) error) (err error)
	Close() error
	DropTableIfExists(value interface{}) error
	GetSQLDB() (*sql.DB, error)
	Clauses(conds ...clause.Expression) *gorm.DB
	AutoMigrate(value interface{}) error
}

type postRepository struct {
	db *gorm.DB
}

func NewSheepRepository(logger logger.Logger, conf config.Config) PostgresRepository {
	logger.Info("Try database connection")
	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		conf.Postgres.Host, conf.Postgres.Port, conf.Postgres.User,
		conf.Postgres.DbName, conf.Postgres.Password)
	db, err := initPostgres(dsn, logger)
	if err != nil {
		logger.Error("Failure database connection", "error", err)
		os.Exit(config.ErrExitStatus)
	}
	logger.Info("Success database connection")
	return &postRepository{
		db: db,
	}
}

func initPostgres(dsn string, logger logger.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	logger.Info("connected to DB")
	return db, nil
}

func (rep *postRepository) Model(value interface{}) *gorm.DB {
	return rep.db.Model(value)
}

func (rep *postRepository) Select(query interface{}, args ...interface{}) *gorm.DB {
	return rep.db.Select(query, args...)
}

func (rep *postRepository) Find(out interface{}, where ...interface{}) *gorm.DB {
	return rep.db.Find(out, where...)
}

func (rep *postRepository) Exec(sql string, values ...interface{}) *gorm.DB {
	return rep.db.Exec(sql, values...)
}

func (rep *postRepository) First(out interface{}, where ...interface{}) *gorm.DB {
	return rep.db.First(out, where...)
}

func (rep *postRepository) Raw(sql string, values ...interface{}) *gorm.DB {
	return rep.db.Raw(sql, values...)
}

func (rep *postRepository) Create(value interface{}) *gorm.DB {
	return rep.db.Create(value)
}

func (rep *postRepository) Save(value interface{}) *gorm.DB {
	return rep.db.Save(value)
}

func (rep *postRepository) Updates(value interface{}) *gorm.DB {
	return rep.db.Updates(value)
}

func (rep *postRepository) Delete(value interface{}) *gorm.DB {
	return rep.db.Delete(value)
}

func (rep *postRepository) Where(query interface{}, args ...interface{}) *gorm.DB {
	return rep.db.Where(query, args...)
}

func (rep *postRepository) Preload(column string, conditions ...interface{}) *gorm.DB {
	return rep.db.Preload(column, conditions...)
}

func (rep *postRepository) Scopes(funcs ...func(*gorm.DB) *gorm.DB) *gorm.DB {
	return rep.db.Scopes(funcs...)
}

func (rep *postRepository) ScanRows(rows *sql.Rows, result interface{}) error {
	return rep.db.ScanRows(rows, result)
}

func (rep *postRepository) Close() error {
	sqlDB, _ := rep.db.DB()
	return sqlDB.Close()
}

func (rep *postRepository) DropTableIfExists(value interface{}) error {
	return rep.db.Migrator().DropTable(value)
}

func (rep *postRepository) AutoMigrate(value interface{}) error {
	return rep.db.AutoMigrate(value)
}

func (rep *postRepository) GetSQLDB() (*sql.DB, error) {
	return rep.db.DB()
}

func (rep *postRepository) Clauses(conds ...clause.Expression) *gorm.DB {
	return rep.db.Clauses(conds...)
}

func (rep *postRepository) Transaction(fc func(tx PostgresRepository) error) (err error) {
	panicked := true
	tx := rep.db.Begin()
	defer func() {
		if panicked || err != nil {
			tx.Rollback()
		}
	}()

	txrep := &postRepository{}
	txrep.db = tx
	err = fc(txrep)

	if err == nil {
		err = tx.Commit().Error
	}

	panicked = false
	return
}
