package database

import (
	"database/sql"

	"fmt"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

const (
	SQL_CON_LOCAL = iota //local database
	SQL_CON_CUSTOM

	MAX_IDLE_CON = 100
	MAX_OPEN_CON = 100
)

type SCHEMA string

func (s SCHEMA) String() string {
	return string(s)
}

const (
	SCHEMA_PUBLIC SCHEMA = "public"
)

type SqlConfig struct {
	SqlConfigType int
	User          string
	Pass          string
	Host          string
	Port          int
	Schema        SCHEMA //the schema to use
}

type SQLDatabase struct {
	*sql.DB
}

func InitLocalDB() (*SQLDatabase, error) {
	return InitDb(SqlConfig{SQL_CON_LOCAL,
		"postgres",
		"password",
		"localhost",
		5432,
		SCHEMA_PUBLIC})
}

//first param is the sql connection type, SQL_CON_LOCAL...etc.
func InitDb(sqlConfig SqlConfig) (*SQLDatabase, error) {
	flog := log.WithFields(log.Fields{"func": "InitDb", "file": "sqldb.go"})

	var (
		db  *sql.DB
		err error
	)
	if sqlConfig.User == "" || sqlConfig.Pass == "" || sqlConfig.Schema == "" {
		return nil, fmt.Errorf("Error Username[%s] or pass[%s] or schema[%s] is empty.",
			sqlConfig.User, sqlConfig.Pass, sqlConfig.Schema)
	}
	switch {
	case SQL_CON_LOCAL == sqlConfig.SqlConfigType:
		flog.Info("Creating db connection local.")
		fmt.Printf("%s@/%s\n", sqlConfig.User, sqlConfig.Schema)
		connStr := fmt.Sprintf("user=%s password='%s' host=%s port=%d sslmode=disable",
			sqlConfig.User, sqlConfig.Pass, sqlConfig.Host, sqlConfig.Port)
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			return nil, fmt.Errorf("Error connecting to local db: %s", err.Error())
		}
		//_, err = db.Exec(fmt.Sprintf("SET search_path TO %s"), sqlConfig.Schema)
		//if err != nil {
		//	return nil, fmt.Errorf("Error connecting to local db: %s", err.Error())
		//}
	case SQL_CON_CUSTOM == sqlConfig.SqlConfigType:
		flog.Infof("Creating db connection Custom. %s:%d", sqlConfig.Host, sqlConfig.Port)
		fmt.Printf("%s@/%s\n", sqlConfig.User, sqlConfig.Schema)
		connStr := fmt.Sprintf("user=%s password='%s' host=%s port=%d sslmode=disable",
			sqlConfig.User, sqlConfig.Pass, sqlConfig.Host, sqlConfig.Port)
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			return nil, fmt.Errorf("Error connecting to local db: %s", err.Error())
		}
	}
	db.SetMaxIdleConns(MAX_IDLE_CON)
	db.SetMaxOpenConns(MAX_OPEN_CON)

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("Error testing initial connection: %s", err.Error())
	}

	return NewSQLDatabase(db), nil
}

func NewSQLDatabase(db *sql.DB) *SQLDatabase {
	s := new(SQLDatabase)
	s.DB = db
	return s
}
