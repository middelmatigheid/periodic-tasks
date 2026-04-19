package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type TaskList struct {
	DoctorID    *int64
	Status      *taskdomain.Status
	StartDate   *time.Time
	EndDate     *time.Time
	Page        *int64
	GeneratorID *int64
}

type TaskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

func (r *TaskRepository) Create(ctx context.Context, task *taskdomain.Task, tx pgx.Tx) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (doctor_id, title, description, status, due_date, generator_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, doctor_id, title, description, status, due_date, generator_id, created_at, updated_at;
	`

	var row pgx.Row
	if tx != nil {
		row = tx.QueryRow(ctx, query, task.DoctorID, task.Title, task.Description, task.Status, task.DueDate, task.GeneratorID, task.CreatedAt, task.UpdatedAt)
	} else {
		row = r.pool.QueryRow(ctx, query, task.DoctorID, task.Title, task.Description, task.Status, task.DueDate, task.GeneratorID, task.CreatedAt, task.UpdatedAt)
	}
	created, err := ScanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, doctor_id, title, description, status, due_date, generator_id, created_at, updated_at
		FROM tasks
		WHERE id = $1;
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := ScanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return found, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE tasks
		SET doctor_id = $1,
			title = $2,
			description = $3,
			status = $4,
			due_date = $5,
			generator_id = $6,
			updated_at = $7
		WHERE id = $8
		RETURNING id, doctor_id, title, description, status, due_date, generator_id, created_at, updated_at;
	`

	row := r.pool.QueryRow(ctx, query, task.DoctorID, task.Title, task.Description, task.Status, task.DueDate, task.GeneratorID, task.UpdatedAt, task.ID)
	updated, err := ScanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *TaskRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1;`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *TaskRepository) List(ctx context.Context, taskList TaskList) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, doctor_id, title, description, status, due_date, generator_id, created_at, updated_at
		FROM tasks
		WHERE ($1::BIGINT IS NULL OR doctor_id = $1) AND ($2::TEXT IS NULL OR status = $2) AND (due_date IS NULL OR (($3::TIMESTAMP IS NULL OR due_date >= $3) AND ($4::TIMESTAMP IS NULL OR due_date <= $4))) AND ($5::BIGINT IS NULL OR generator_id = $5)
		ORDER BY due_date ASC, id DESC
		LIMIT 10 OFFSET $6;
	`
	page := 0
	if taskList.Page != nil {
		page = int(*taskList.Page)
	}
	rows, err := r.pool.Query(ctx, query, taskList.DoctorID, taskList.Status, taskList.StartDate, taskList.EndDate, taskList.GeneratorID, 10*page)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := ScanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func ScanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task   taskdomain.Task
		status string
	)

	if err := scanner.Scan(
		&task.ID,
		&task.DoctorID,
		&task.Title,
		&task.Description,
		&status,
		&task.DueDate,
		&task.GeneratorID,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)

	return &task, nil
}
