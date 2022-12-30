package db

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/apex/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB
var mx sync.Mutex

// NamedExec is just a Proxy to sqlx.NamedExec with a Mutex Lock
func NamedExec(query string, arg interface{}) (sql.Result, error) {
	mx.Lock()
	defer mx.Unlock()
	return DB.NamedExec(query, arg)
}

// Get is just a Proxy to sqlx.Get with a Mutex Lock
func Get(dest interface{}, query string, args ...interface{}) error {
	mx.Lock()
	defer mx.Unlock()
	return DB.Get(dest, query, args...)
}

func Exec(query string, args ...any) (sql.Result, error) {
	mx.Lock()
	defer mx.Unlock()
	return DB.Exec(query, args...)
}

func CreateDB(host, name, user, pass string) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, pass, host, name)

	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		log.Fatal("unable to connect DB")
	}

	if err := db.Ping(); err != nil {
		log.Fatal(fmt.Sprintf("%s", err))
	}

	DB = db

	fmt.Println("Database connection OK.")
}
