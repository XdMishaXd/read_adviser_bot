package postgresql

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"read_adviser_bot/internal/config"
	"read_adviser_bot/internal/storage"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepo struct {
	pool *pgxpool.Pool
}

// Connect создает подключение к базе данных и возвращает репозиторий.
func Connect(ctx context.Context, cfg *config.Config, log *slog.Logger) (*PostgresRepo, error) {
	const op = "storage.postgresql.Connect"

	dsn := dsn(cfg)
	log.Info("dsn string (quoted)", slog.String("raw", fmt.Sprintf("%q", dsn)))

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse config: %w", op, err)
	}

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = time.Minute * 30

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create pool: %w", op, err)
	}

	// if err := pool.Ping(ctx); err != nil {
	// 	pool.Close()
	// 	return nil, fmt.Errorf("%s: failed to ping database: %w", op, err) // TODO: в проде убрать
	// }

	log.Info("successfully connected to database")

	return &PostgresRepo{pool: pool}, nil
}

// New создает репозиторий с уже готовым пулом (для тестов или специальных случаев).
func New(pool *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{
		pool: pool,
	}
}

// SaveLink сохраняет ссылку в базу данных.
func (r *PostgresRepo) SaveLink(ctx context.Context, link string, username string) (int64, error) {
	const op = "storage.postgresql.SaveLink"
	var id int64

	err := r.pool.QueryRow(
		ctx,
		`INSERT INTO links (link, username) VALUES ($1, $2) RETURNING id`,
		link,
		username,
	).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return -1, storage.ErrAlreadyExists
		}

		return -1, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

// RandomLink берёт из базы данных одну случайную ссылку и удаляет её.
func (r *PostgresRepo) RandomLink(ctx context.Context, username string) (*storage.Link, error) {
	const op = "storage.postgresql.RandomLink"
	var l storage.Link

	err := r.pool.QueryRow(
		ctx,
		`WITH deleted AS (
			SELECT id, link, username FROM links 
			WHERE username = $1 
			ORDER BY RANDOM() 
			LIMIT 1
		)
		DELETE FROM links
		USING deleted
		WHERE links.id = deleted.id
		RETURNING deleted.id, deleted.link, deleted.username`,
		username,
	).Scan(&l.ID, &l.Link, &l.Username)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrEmptyTable
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &l, nil
}

// IsExists проверяет, существует ли ссылка в базе данных.
func (r *PostgresRepo) IsExists(ctx context.Context, link string) (int64, error) {
	const op = "storage.postgresql.IsExists"
	var id int64

	err := r.pool.QueryRow(
		ctx,
		`SELECT id FROM links WHERE link = $1 LIMIT 1`,
		link,
	).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return -1, storage.ErrEmptyTable
		}

		return -1, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

// Close закрывает соединение с базой данных.
func (r *PostgresRepo) Close() {
	r.pool.Close()
}

// dsn формирует конфигурацию базы данных.
func dsn(cfg *config.Config) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s database=%s sslmode=%s",
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.DBName,
		cfg.Postgres.SSLMode,
	)
}
