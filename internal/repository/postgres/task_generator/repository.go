package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskgeneratordomain "example.com/taskservice/internal/domain/task_generator"
)

type TaskGeneratorList struct {
	DoctorID  *int64
	Status    *taskdomain.Status
	StartDate *time.Time
	EndDate   *time.Time
	Page      *int64
}

type TaskGeneratorProcessList struct {
	Now     func() time.Time
	EndDate time.Time
}

type TaskGeneratorRepository struct {
	pool *pgxpool.Pool
}

func NewTaskGeneratorRepository(pool *pgxpool.Pool) *TaskGeneratorRepository {
	return &TaskGeneratorRepository{pool: pool}
}

func (r *TaskGeneratorRepository) NewTx(ctx context.Context) (pgx.Tx, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (r *TaskGeneratorRepository) Create(ctx context.Context, taskGenerator *taskgeneratordomain.TaskGenerator) (*taskgeneratordomain.TaskGenerator, error) {
	const query = `
        INSERT INTO task_generators (doctor_id, title, description, status, every_n_days, every_ith_day, parity, next_due_date, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING id, doctor_id, title, description, status, every_n_days, every_ith_day, parity, next_due_date, created_at, updated_at, processed_at;
    `

	row := r.pool.QueryRow(ctx, query, taskGenerator.DoctorID, taskGenerator.Title, taskGenerator.Description, taskGenerator.Status, taskGenerator.EveryNDays, taskGenerator.EveryIthDay, taskGenerator.Parity, taskGenerator.NextDueDate, taskGenerator.CreatedAt, taskGenerator.UpdatedAt)
	created, err := scanTaskGenerator(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *TaskGeneratorRepository) GetByID(ctx context.Context, id int64) (*taskgeneratordomain.TaskGenerator, error) {
	const query = `
        SELECT id, doctor_id, title, description, status, every_n_days, every_ith_day, parity, next_due_date, created_at, updated_at, processed_at
        FROM task_generators
        WHERE id = $1;
    `

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTaskGenerator(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskgeneratordomain.ErrNotFound
		}

		return nil, err
	}

	return found, nil
}

func (r *TaskGeneratorRepository) Update(ctx context.Context, taskGenerator *taskgeneratordomain.TaskGenerator, tx pgx.Tx) (*taskgeneratordomain.TaskGenerator, error) {
	const query = `
        UPDATE task_generators
        SET doctor_id = $1,
            title = $2,
            description = $3,
            status = $4,
            every_n_days = $5,
            every_ith_day = $6,
            parity = $7,
            next_due_date = $8,
            updated_at = $9
        WHERE id = $10
        RETURNING id, doctor_id, title, description, status, every_n_days, every_ith_day, parity, next_due_date, created_at, updated_at, processed_at;
    `

	var row pgx.Row
	if tx != nil {
		row = tx.QueryRow(ctx, query, taskGenerator.DoctorID, taskGenerator.Title, taskGenerator.Description, taskGenerator.Status, taskGenerator.EveryNDays, taskGenerator.EveryIthDay, taskGenerator.Parity, taskGenerator.NextDueDate, taskGenerator.UpdatedAt, taskGenerator.ID)
	} else {
		row = r.pool.QueryRow(ctx, query, taskGenerator.DoctorID, taskGenerator.Title, taskGenerator.Description, taskGenerator.Status, taskGenerator.EveryNDays, taskGenerator.EveryIthDay, taskGenerator.Parity, taskGenerator.NextDueDate, taskGenerator.UpdatedAt, taskGenerator.ID)
	}
	updated, err := scanTaskGenerator(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskgeneratordomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *TaskGeneratorRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM task_generators WHERE id = $1;`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskgeneratordomain.ErrNotFound
	}

	return nil
}

func (r *TaskGeneratorRepository) List(ctx context.Context, taskGeneratorList TaskGeneratorList) ([]taskgeneratordomain.TaskGenerator, error) {
	const query = `
        SELECT id, doctor_id, title, description, status, every_n_days, every_ith_day, parity, next_due_date, created_at, updated_at, processed_at
        FROM task_generators
        WHERE ($1::BIGINT IS NULL OR doctor_id = $1) AND ($2::TEXT IS NULL OR status = $2) AND ($3::TIMESTAMP IS NULL OR next_due_date >= $3) AND ($4::TIMESTAMP IS NULL OR next_due_date <= $4)
        ORDER BY next_due_date ASC, id DESC
        LIMIT 10 OFFSET $5;
    `

	page := 0
	if taskGeneratorList.Page != nil {
		page = int(*taskGeneratorList.Page)
	}
	rows, err := r.pool.Query(ctx, query, taskGeneratorList.DoctorID, taskGeneratorList.Status, taskGeneratorList.StartDate, taskGeneratorList.EndDate, 10*page)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	taskGenerators := make([]taskgeneratordomain.TaskGenerator, 0)
	for rows.Next() {
		taskGenerator, err := scanTaskGenerator(rows)
		if err != nil {
			return nil, err
		}

		taskGenerators = append(taskGenerators, *taskGenerator)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return taskGenerators, nil
}

func (r *TaskGeneratorRepository) ProcessList(ctx context.Context, taskGeneratorProcessList TaskGeneratorProcessList, cooldown int64) ([]int64, error) {
	const query = `
        UPDATE task_generators SET processed_at = $3
        WHERE id IN (
            SELECT id 
            FROM task_generators 
            WHERE next_due_date <= $1 AND processed_at <= $2
            ORDER BY next_due_date ASC
            LIMIT 10
        ) RETURNING id;
    `
	now := taskGeneratorProcessList.Now()
	rows, err := r.pool.Query(ctx, query, taskGeneratorProcessList.EndDate, now.Add(-time.Minute*time.Duration(cooldown)), now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	taskGeneratorIDs := make([]int64, 0)
	for rows.Next() {
		var taskGeneratorID int64
		err := rows.Scan(&taskGeneratorID)
		if err != nil {
			return nil, err
		}
		taskGeneratorIDs = append(taskGeneratorIDs, taskGeneratorID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return taskGeneratorIDs, nil
}

type taskGeneratorScanner interface {
	Scan(dest ...any) error
}

func scanTaskGenerator(scanner taskGeneratorScanner) (*taskgeneratordomain.TaskGenerator, error) {
	var (
		taskGenerator taskgeneratordomain.TaskGenerator
		status        string
	)

	if err := scanner.Scan(
		&taskGenerator.ID,
		&taskGenerator.DoctorID,
		&taskGenerator.Title,
		&taskGenerator.Description,
		&status,
		&taskGenerator.EveryNDays,
		&taskGenerator.EveryIthDay,
		&taskGenerator.Parity,
		&taskGenerator.NextDueDate,
		&taskGenerator.CreatedAt,
		&taskGenerator.UpdatedAt,
		&taskGenerator.ProcessedAt,
	); err != nil {
		return nil, err
	}

	taskGenerator.Status = taskdomain.Status(status)

	return &taskGenerator, nil
}
