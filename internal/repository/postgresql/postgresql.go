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

func (rep *PgRepository) CreateUser(ctx context.Context, login string, hashedPass string) (*models.UserPublic, error) {
	err := createUser(ctx, login, hashedPass, rep.pool)

	if err != nil {
		return nil, pgerr.ClassifyUserErr(err)
	}

	user, err := getUserByName(ctx, login, rep.pool)

	if err != nil {
		return nil, err
	}

	return &models.UserPublic{
		ID:  user.ID,
		Log: user.Log,
	}, nil
}

func (rep *PgRepository) GetUserByName(ctx context.Context, login string) (*models.User, error) {
	return getUserByName(ctx, login, rep.pool)
}

func (rep *PgRepository) CreateOrder(ctx context.Context, number string, userID int, status models.OrderStatus) error {
	return createOrder(ctx, number, userID, status, rep.pool)
}

func (rep *PgRepository) UpdateOrder(ctx context.Context, number string, status models.OrderStatus, accrual *float64) error {
	return updateOrder(ctx, number, status, accrual, rep.pool)
}

func (rep *PgRepository) UpdateOrderAndBalance(ctx context.Context, userID int, number string, status models.OrderStatus, accrual *float64) error {
	txOptions := pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	}

	tx, err := rep.pool.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}

	err = updateOrder(ctx, number, status, accrual, tx)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	err = updateUserBalance(ctx, userID, *accrual, tx)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	tx.Commit(ctx)
	return nil
}

func (rep *PgRepository) GetUserBalance(ctx context.Context, userID int) (*models.Balance, error) {
	sum, err := getWithdrawsSum(ctx, userID, rep.pool)
	if err != nil {
		return nil, err
	}

	user, err := getUserByID(ctx, userID, rep.pool)
	if err != nil {
		return nil, err
	}

	return &models.Balance{
		Current:   user.Balance,
		Withdrawn: sum,
	}, nil
}

func (rep *PgRepository) GetWithdraws(ctx context.Context, userID int) ([]models.Withdraw, error) {
	return getWithdraws(ctx, userID, rep.pool)
}

func (rep *PgRepository) MakeWithdraw(ctx context.Context, userID int, withdraw float64, orderNum string) error {
	txOptions := pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	}

	tx, err := rep.pool.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}

	err = updateUserBalance(ctx, userID, -withdraw, tx)
	if err != nil {
		tx.Rollback(ctx)
		return pgerr.ClassifyUserErr(err)
	}

	err = addWithdraw(ctx, userID, withdraw, orderNum, tx)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	tx.Commit(ctx)
	return nil
}

func (rep *PgRepository) GetOrderByNum(ctx context.Context, number string) (*models.Order, error) {
	order, err := getOrderByNumber(ctx, number, rep.pool)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNoOrderFound
		}
		return nil, err
	}

	return order, nil
}

func (rep *PgRepository) GetUserOrders(ctx context.Context, userID int) ([]models.Order, error) {
	return getUserOrders(ctx, userID, rep.pool)
}

func (rep *PgRepository) GetProcessingOrders(ctx context.Context) ([]models.Order, error) {
	return getProcessingOrders(ctx, rep.pool)
}

func (rep *PgRepository) Close() {
	rep.pool.Close()
}

func (rep *PgRepository) Ping(ctx context.Context) error {
	return rep.pool.Ping(ctx)
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

func getUserByName(ctx context.Context, username string, DBCon DBConnection) (*models.User, error) {
	user := models.User{}
	err := DBCon.QueryRow(ctx,
		`SELECT id, username, password_hash, balance FROM users WHERE username = @username`,
		pgx.NamedArgs{"username": username},
	).Scan(&user.ID, &user.Log, &user.Pass, &user.Balance)

	return &user, err
}

func getUserByID(ctx context.Context, id int, DBCon DBConnection) (*models.User, error) {
	user := models.User{}
	err := DBCon.QueryRow(ctx,
		`SELECT id, username, password_hash, balance FROM users WHERE id = @id`,
		pgx.NamedArgs{"id": id},
	).Scan(&user.ID, &user.Log, &user.Pass, &user.Balance)

	return &user, err
}

func updateUserBalance(ctx context.Context, userID int, changeBalance float64, DBCon DBConnection) error {
	_, err := DBCon.Exec(ctx,
		`UPDATE users SET balance = balance + @change_balance WHERE id = @userID`,
		pgx.NamedArgs{
			"change_balance": changeBalance,
			"userID":         userID,
		},
	)

	return err
}

func createUser(ctx context.Context, login string, hashedPass string, DBCon DBConnection) error {
	_, err := DBCon.Exec(ctx,
		`INSERT INTO users (username, password_hash) VALUES (@login, @hashed_pass)`,
		pgx.NamedArgs{
			"login":       login,
			"hashed_pass": hashedPass,
		},
	)

	return err
}

func getOrderByNumber(ctx context.Context, number string, DBCon DBConnection) (*models.Order, error) {
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

func getUserOrders(ctx context.Context, userID int, DBCon DBConnection) ([]models.Order, error) {
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

func getProcessingOrders(ctx context.Context, DBCon DBConnection) ([]models.Order, error) {
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

func createOrder(ctx context.Context, number string, userID int, status models.OrderStatus, DBCon DBConnection) error {
	_, err := DBCon.Exec(ctx,
		`INSERT INTO orders (number, status, user_id) VALUES (@number, @status, @userID)`,
		pgx.NamedArgs{
			"number": number,
			"status": status,
			"userID": userID,
		},
	)

	return err
}

func updateOrder(ctx context.Context, number string, status models.OrderStatus, accrual *float64, DBCon DBConnection) error {
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

func addWithdraw(ctx context.Context, userID int, withdraw float64, orderNum string, DBCon DBConnection) error {
	_, err := DBCon.Exec(ctx,
		`INSERT INTO withdraws (withdraw, order_num, user_id) VALUES (@withdraw, @order_num, @userID)`,
		pgx.NamedArgs{
			"withdraw":  withdraw,
			"order_num": orderNum,
			"userID":    userID,
		},
	)

	return err
}

func getWithdraws(ctx context.Context, userID int, DBCon DBConnection) ([]models.Withdraw, error) {
	var withdraws []models.Withdraw

	rows, err := DBCon.Query(ctx,
		`SELECT withdraw, order_num, processed_at FROM withdraws WHERE user_id = @userID`,
		pgx.NamedArgs{"userID": userID},
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		withdraw := models.Withdraw{}

		err := rows.Scan(&withdraw.Sum, &withdraw.Order, &withdraw.Processed)
		if err != nil {
			return nil, err
		}

		withdraws = append(withdraws, withdraw)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return withdraws, nil
}

func getWithdrawsSum(ctx context.Context, userID int, DBCon DBConnection) (float64, error) {
	var sum sql.NullFloat64

	err := DBCon.QueryRow(ctx,
		`SELECT SUM(withdraw) FROM withdraws WHERE user_id = @userID`,
		pgx.NamedArgs{"userID": userID},
	).Scan(&sum)

	if err != nil {
		return 0, err
	}

	if sum.Valid {
		return sum.Float64, nil
	}

	return 0, nil
}
