package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Zhukek/loyalty/internal/errs"
	models "github.com/Zhukek/loyalty/internal/models"
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

func (rep *PgRepository) CreateUser(login string, hashed_pass string, ctx context.Context) (*models.UserPublic, error) {
	err := createUser(login, hashed_pass, rep.pool, ctx)

	if err != nil {
		return nil, pgerr.ClassifyUserErr(err)
	}

	user, err := getUserByName(login, rep.pool, ctx)

	if err != nil {
		return nil, err
	}

	return &models.UserPublic{
		Id:  user.Id,
		Log: user.Log,
	}, nil
}

func (rep *PgRepository) GetUserByName(login string, ctx context.Context) (*models.User, error) {
	return getUserByName(login, rep.pool, ctx)
}

func (rep *PgRepository) CreateOrder(number string, userID int, status models.OrderStatus, ctx context.Context) error {
	return createOrder(number, userID, status, rep.pool, ctx)
}

func (rep *PgRepository) UpdateOrder(number string, status models.OrderStatus, accrual *float64, ctx context.Context) error {
	return updateOrder(number, status, accrual, rep.pool, ctx)
}

func (rep *PgRepository) UpdateOrderAndBalance(user_id int, number string, status models.OrderStatus, accrual *float64, ctx context.Context) error {
	txOptions := pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	}

	tx, err := rep.pool.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}

	err = updateOrder(number, status, accrual, tx, ctx)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	err = updateUserBalance(user_id, *accrual, tx, ctx)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	tx.Commit(ctx)
	return nil
}

func (rep *PgRepository) MakeWithdraw(user_id int, withdraw float64, orderNum string, ctx context.Context) error {
	txOptions := pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	}

	tx, err := rep.pool.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}

	err = updateUserBalance(user_id, -withdraw, tx, ctx)
	if err != nil {
		tx.Rollback(ctx)
		return pgerr.ClassifyUserErr(err)
	}

	err = addWithdraw(user_id, withdraw, orderNum, tx, ctx)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	tx.Commit(ctx)
	return nil
}

func (rep *PgRepository) GetOrderByNum(number string, ctx context.Context) (*models.Order, error) {
	order, err := getOrderByNumber(number, rep.pool, ctx)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNoOrderFound
		}
		return nil, err
	}

	return order, nil
}

func (rep *PgRepository) GetUserOrders(userID int, ctx context.Context) ([]models.Order, error) {
	return getUserOrders(userID, rep.pool, ctx)
}

func (rep *PgRepository) GetProcessingOrders(ctx context.Context) ([]models.Order, error) {
	return getProcessingOrders(rep.pool, ctx)
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
		`SELECT id, username, password_hash, balance FROM users WHERE username = @username`,
		pgx.NamedArgs{"username": username},
	).Scan(&user.Id, &user.Log, &user.Pass, &user.Balance)

	return &user, err
}

func updateUserBalance(user_id int, changeBalance float64, DBCon DBConnection, ctx context.Context) error {
	_, err := DBCon.Exec(ctx,
		`UPDATE users SET balance = balance + @change_balance WHERE id = @user_id`,
		pgx.NamedArgs{
			"change_balance": changeBalance,
			"user_id":        user_id,
		},
	)

	return err
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

func getOrderByNumber(number string, DBCon DBConnection, ctx context.Context) (*models.Order, error) {
	order := models.Order{}
	var accrual sql.NullFloat64

	err := DBCon.QueryRow(ctx,
		`SELECT number, status, accrual, uploaded_at, user_id FROM orders WHERE number = @number`,
		pgx.NamedArgs{"number": number},
	).Scan(&order.Number, &order.Status, &accrual, &order.Uploaded, &order.UserID)

	if accrual.Valid {
		order.Accrual = accrual.Float64
	}

	return &order, err
}

func getUserOrders(userID int, DBCon DBConnection, ctx context.Context) ([]models.Order, error) {
	var orders []models.Order

	rows, err := DBCon.Query(ctx,
		`SELECT number, status, accrual, uploaded_at FROM orders WHERE user_id = @userID`,
		pgx.NamedArgs{"userID": userID},
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		order := models.Order{}
		var accrual sql.NullFloat64

		err = rows.Scan(&order.Number, &order.Status, &accrual, &order.Uploaded)
		if err != nil {
			return nil, err
		}

		if accrual.Valid {
			order.Accrual = accrual.Float64
		}

		orders = append(orders, order)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func getProcessingOrders(DBCon DBConnection, ctx context.Context) ([]models.Order, error) {
	var orders []models.Order

	rows, err := DBCon.Query(ctx,
		`SELECT number, status, user_id FROM orders WHERE status = 'NEW' OR status = 'PROCESSING'`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		order := models.Order{}
		err = rows.Scan(&order.Number, &order.Status, &order.UserID)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func createOrder(number string, userID int, status models.OrderStatus, DBCon DBConnection, ctx context.Context) error {
	_, err := DBCon.Exec(ctx,
		`INSERT INTO orders (number, status, user_id) VALUES (@number, @status, @user_id)`,
		pgx.NamedArgs{
			"number":  number,
			"status":  status,
			"user_id": userID,
		},
	)

	return err
}

func updateOrder(number string, status models.OrderStatus, accrual *float64, DBCon DBConnection, ctx context.Context) error {
	query := `UPDATE orders SET status = @status`
	args := pgx.NamedArgs{
		"status": status,
		"number": number,
	}
	if accrual != nil {
		query += `, accrual = @accrual`
		args["accrual"] = *accrual
	}
	query += ` WHERE number = @number`
	_, err := DBCon.Exec(ctx, query, args)

	return err
}

func addWithdraw(user_id int, withdraw float64, order_num string, DBCon DBConnection, ctx context.Context) error {
	_, err := DBCon.Exec(ctx,
		`INSERT INTO withdraws (withdraw, order_num, user_id) VALUES (@withdraw, @order_num, @user_id)`,
		pgx.NamedArgs{
			"withdraw":  withdraw,
			"order_num": order_num,
			"user_id":   user_id,
		},
	)

	return err
}
