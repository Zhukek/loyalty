package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	models "github.com/Zhukek/loyalty/internal/model"
	"github.com/Zhukek/loyalty/internal/repository/postgresql/pgerr"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBConnection interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type PgRepository struct {
	pool *pgxpool.Pool
}

func (rep *PgRepository) CreateUser(login string, hashed_pass string, ctx context.Context) (*models.User, error) {
	err := createUser(login, hashed_pass, rep.pool, ctx)

	if err != nil {
		return nil, pgerr.ClassifyErr(err)
	}

	user, err := getUserByName(login, rep.pool, ctx)

	if err != nil {
		return nil, pgerr.ClassifyErr(err)
	}

	return user, nil
}

func (rep *PgRepository) Close() {
	rep.pool.Close()
}

func (r *PgRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func NewPGRepository(DBURI string) (*PgRepository, error) {
	config, err := pgxpool.ParseConfig(DBURI)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Minute * 30
	config.MaxConnIdleTime = time.Minute * 15
	config.HealthCheckPeriod = time.Minute * 1
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	if err := migration(DBURI); err != nil {
		pool.Close()
		return nil, err
	}

	return &PgRepository{pool: pool}, nil
}

func migration(DBURI string) error {
	db, err := sql.Open("postgres", DBURI)
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	migration, err := migrate.NewWithDatabaseInstance("file://migrations",
		"postgres", driver)
	if err != nil {
		return err
	}

	err = migration.Up()
	if err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
		fmt.Println("migration: no change")
	}

	return nil
}

func getUserByName(username string, DBCon DBConnection, ctx context.Context) (*models.User, error) {
	user := models.User{}
	err := DBCon.QueryRow(ctx,
		`SELECT id, username FROM users WHERE username = @username`,
		pgx.NamedArgs{"username": username},
	).Scan(&user.Id, &user.Log)

	return &user, err
}

func getUserByID(id int, DBCon DBConnection, ctx context.Context) (*models.User, error) {
	user := models.User{}
	err := DBCon.QueryRow(ctx,
		`SELECT id, username FROM users WHERE id = @id`,
		pgx.NamedArgs{"id": id},
	).Scan(&user.Id, &user.Log)

	return &user, err
}

func createUser(login string, hashed_pass string, DBCon DBConnection, ctx context.Context) error {
	_, err := DBCon.Exec(ctx,
		`INSERT INTO users (username, password_hash) VALUES (@login, @hashed_pass)`,
		pgx.NamedArgs{
			"login":       login,
			"hashed_pass": hashed_pass,
		},
	)

	return err
}
