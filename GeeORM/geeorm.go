package geeorm

import (
	"database/sql"
	"geeorm/dialect"
	"geeorm/logger"
	"geeorm/session"
)

type Engine struct {
	db      *sql.DB
	dialect dialect.Dialect
}

func NewEngine(driver, source string) (e *Engine, err error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		logger.Error(err)
		return
	}

	if err = db.Ping(); err != nil {
		logger.Error(err)
		return
	}

	dial, ok := dialect.GetDialect(driver)
	if !ok {
		logger.Errorf("dialect %s not found", driver)
		return
	}
	e = &Engine{db: db, dialect: dial}
	logger.Info("Connect database success")
	return
}

func (e *Engine) Close() {
	if err := e.db.Close(); err != nil {
		logger.Error("Failed to close database")
	}
	logger.Info("Close database success")
}

func (e *Engine) NewSession() *session.Session {
	return session.New(e.db, e.dialect)
}
