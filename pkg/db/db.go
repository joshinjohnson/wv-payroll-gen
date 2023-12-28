package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

const (
	UsernameField  = "user"
	PasswordField  = "password"
	HostNameField  = "hostname"
	PortField      = "port"
	DbNameField    = "db-name"
	WarehouseField = "warehouse-name"
	SchemaField    = "schema"
	SSLModeField   = "ssl-mode"
)

type DbWrapper struct {
	ctx    context.Context
	DB     *sql.DB
	Tx     *sql.Tx
	logger logrus.Logger
}

func NewDbWrapper(ctx context.Context, config map[string]string) (*DbWrapper, error) {
	db, err := newPostgresDb(config)
	if err != nil {
		return nil, err
	}

	// TODO: Get from config
	db.SetConnMaxIdleTime(time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	return &DbWrapper{
		DB:     db,
		ctx:    ctx,
		logger: logrus.Logger{},
	}, nil
}

func newPostgresDb(config map[string]string) (*sql.DB, error) {
	var user, pass, host, port, dbName, sslmode string
	var ok bool

	if user, ok = config[UsernameField]; !ok {
		return nil, errors.New("db-username not found")
	}
	if pass, ok = config[PasswordField]; !ok {
		return nil, errors.New("db-password not found")
	}
	if host, ok = config[HostNameField]; !ok {
		return nil, errors.New("db-host not found")
	}
	if port, ok = config[PortField]; !ok {
		return nil, errors.New("db-port not found")
	}
	if dbName, ok = config[DbNameField]; !ok {
		return nil, errors.New("db-name not found")
	}
	if sslmode, ok = config[SSLModeField]; !ok {
		return nil, errors.New("sslm-mode not found")
	}

	_, err := strconv.ParseUint(port, 10, 32)
	if err != nil {
		return nil, errors.New("invalid port found")
	}

	db, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, pass, host, port, dbName, sslmode))
	if err != nil {
		return nil, err
	}

	return db, nil
}
