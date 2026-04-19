package usecase

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskgeneratordomain "example.com/taskservice/internal/domain/task_generator"
	taskgeneratorrepository "example.com/taskservice/internal/repository/postgres/task_generator"
)

type TaskGeneratorRepository interface {
	NewTx(ctx context.Context) (pgx.Tx, error)
	Create(ctx context.Context, taskGenerator *taskgeneratordomain.TaskGenerator) (*taskgeneratordomain.TaskGenerator, error)
	GetByID(ctx context.Context, id int64) (*taskgeneratordomain.TaskGenerator, error)
	Update(ctx context.Context, taskGenerator *taskgeneratordomain.TaskGenerator, tx pgx.Tx) (*taskgeneratordomain.TaskGenerator, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, taskGeneratorList taskgeneratorrepository.TaskGeneratorList) ([]taskgeneratordomain.TaskGenerator, error)
	ProcessList(ctx context.Context, taskGeneratorProcessList taskgeneratorrepository.TaskGeneratorProcessList, reprocessingCooldown int64) ([]int64, error)
}

type TaskGeneratorUsecase interface {
	Generate(ctx context.Context, id int64) (*taskdomain.Task, error)
	Create(ctx context.Context, input CreateTaskGeneratorInput) (*taskgeneratordomain.TaskGenerator, error)
	GetByID(ctx context.Context, id int64) (*taskgeneratordomain.TaskGenerator, error)
	Update(ctx context.Context, id int64, input UpdateTaskGeneratorInput) (*taskgeneratordomain.TaskGenerator, error)
	UpdateWithTx(ctx context.Context, id int64, input UpdateTaskGeneratorInput, tx pgx.Tx) (*taskgeneratordomain.TaskGenerator, error)
	Patch(ctx context.Context, id int64, input PatchTaskGeneratorInput) (*taskgeneratordomain.TaskGenerator, error)
	PatchWithTx(ctx context.Context, id int64, input PatchTaskGeneratorInput, tx pgx.Tx) (*taskgeneratordomain.TaskGenerator, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, taskGeneratorList taskgeneratorrepository.TaskGeneratorList) ([]taskgeneratordomain.TaskGenerator, error)
	ProcessList(ctx context.Context, taskGeneratorProcessList taskgeneratorrepository.TaskGeneratorProcessList, reprocessingCooldown int64) ([]int64, error)
}

type CreateTaskGeneratorInput struct {
	DoctorID    *int64
	Title       string
	Description string
	Status      taskdomain.Status
	EveryNDays  *int64
	EveryIthDay *int64
	Parity      bool
	NextDueDate time.Time
}

type UpdateTaskGeneratorInput struct {
	DoctorID    *int64
	Title       string
	Description string
	Status      taskdomain.Status
	EveryNDays  *int64
	EveryIthDay *int64
	Parity      bool
	NextDueDate time.Time
}

type PatchTaskGeneratorInput struct {
	DoctorID    *int64
	Title       *string
	Description *string
	Status      *taskdomain.Status
	EveryNDays  *int64
	EveryIthDay *int64
	Parity      *bool
	NextDueDate *time.Time
}

type RunTaskGeneratorParams struct {
	ReprocessingCooldown int64
	EveryNMinutes        int64
	DaysAhead            int64
	NumWorkers           int64
}
