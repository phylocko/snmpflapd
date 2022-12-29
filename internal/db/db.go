package db

import (
	"database/sql"
	"fmt"
	"snmpflapd/internal/logger"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB
var mx sync.Mutex

// NamedExec is just a Proxy to sqlx.NamedExec with a Mutex Lock
func NamedExec(query string, args map[string]interface{}) (sql.Result, error) {
	mx.Lock()
	defer mx.Unlock()
	return DB.NamedExec(query, args)
}

// Get is just a Proxy to sqlx.Get with a Mutex Lock
func Get(dest interface{}, query string, args ...interface{}) error {
	mx.Lock()
	defer mx.Unlock()
	return DB.Get(dest, query, args...)
}

func Exec(query string, args ...interface{}) (sql.Result, error) {
	mx.Lock()
	defer mx.Unlock()
	return DB.Exec(query, args)
}

func CreateDB(host, name, user, pass string) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, pass, host, name)

	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		logger.L.Fatal("unable to connect DB")
	}

	if err := db.Ping(); err != nil {
		logger.L.Fatal("unable to ping DB")
	}

	DB = db

	logger.L.Println("database connection OK")
}
