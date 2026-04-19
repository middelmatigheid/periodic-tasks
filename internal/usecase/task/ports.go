package usecase

import (
	"context"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskrepository "example.com/taskservice/internal/repository/postgres/task"
	"github.com/jackc/pgx/v5"
)

type TaskRepository interface {
	Create(ctx context.Context, task *taskdomain.Task, tx pgx.Tx) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, taskList taskrepository.TaskList) ([]taskdomain.Task, error)
}

type TaskUsecase interface {
	Create(ctx context.Context, input CreateTaskInput) (*taskdomain.Task, error)
	CreateWithTx(ctx context.Context, input CreateTaskInput, tx pgx.Tx) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateTaskInput) (*taskdomain.Task, error)
	Patch(ctx context.Context, id int64, input PatchTaskInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, taskList taskrepository.TaskList) ([]taskdomain.Task, error)
}

type CreateTaskInput struct {
	DoctorID    *int64
	Title       string
	Description string
	Status      taskdomain.Status
	DueDate     *time.Time
	GeneratorID *int64
}

type UpdateTaskInput struct {
	DoctorID    *int64
	Title       string
	Description string
	Status      taskdomain.Status
	DueDate     *time.Time
}

type PatchTaskInput struct {
	DoctorID    *int64
	Title       *string
	Description *string
	Status      *taskdomain.Status
	DueDate     *time.Time
}
