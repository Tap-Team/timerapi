package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/docker/go-connections/nat"
	"github.com/golang-migrate/migrate/v4"
	mpq "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/testcontainers/testcontainers-go"
	ctnrpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	dbname   = "amiddoc_go"
	user     = "postgres"
	password = "postgres"
	port     = "5432"
)

const DEFAULT_MIGRATION_PATH = "/home/amidman/.go/src/github.com/Tap-Team/timer/migrations"

// return postgres, close function and err
// can set migration source package,
// them should be in file://<package> format
func NewContainer(ctx context.Context, migrationFolder string) (*Postgres, func(context.Context) error, error) {
	container, err := ctnrpostgres.RunContainer(
		ctx,
		testcontainers.WithImage("docker.io/postgres:alpine3.17"),
		ctnrpostgres.WithDatabase(dbname),
		ctnrpostgres.WithUsername(user),
		ctnrpostgres.WithPassword(password),
		testcontainers.WithWaitStrategy(wait.ForListeningPort(nat.Port(port))),
	)
	if err != nil {
		return nil, nil, err
	}
	container.Start(ctx)
	dbUrl, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, nil, err
	}
	p, err := New(dbUrl)
	if err != nil {
		return nil, nil, err
	}
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		return nil, nil, err
	}
	driver, err := mpq.WithInstance(db, &mpq.Config{})
	if err != nil {
		return nil, nil, err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationFolder,
		dbUrl,
		driver,
	)
	if err != nil {
		return nil, nil, errors.Join(err, errors.New("migrate error"))
	}
	err = m.Up()
	if err != nil {
		return nil, nil, errors.Join(err, errors.New("migrate up error"))
	}
	term := func(ctx context.Context) error {
		p.Close()
		return container.Terminate(ctx)
	}
	return p, term, nil
}
