package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskrepository "example.com/taskservice/internal/repository/postgres/task"
	"github.com/jackc/pgx/v5"
)

type TaskService struct {
	repo TaskRepository
	now  func() time.Time
}

func NewTaskService(repo TaskRepository) *TaskService {
	return &TaskService{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *TaskService) Create(ctx context.Context, input CreateTaskInput) (*taskdomain.Task, error) {
	return s.create(ctx, input, nil)
}

func (s *TaskService) CreateWithTx(ctx context.Context, input CreateTaskInput, tx pgx.Tx) (*taskdomain.Task, error) {
	return s.create(ctx, input, tx)
}

func (s *TaskService) create(ctx context.Context, input CreateTaskInput, tx pgx.Tx) (*taskdomain.Task, error) {
	normalized, err := validateCreateTaskInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		DoctorID:    normalized.DoctorID,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		DueDate:     normalized.DueDate,
		GeneratorID: normalized.GeneratorID,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model, tx)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *TaskService) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *TaskService) Update(ctx context.Context, id int64, input UpdateTaskInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateTaskInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:          id,
		DoctorID:    normalized.DoctorID,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		DueDate:     normalized.DueDate,
		UpdatedAt:   s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *TaskService) Patch(ctx context.Context, id int64, input PatchTaskInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	if input.DoctorID == nil && input.Title == nil && input.Description == nil && input.Status == nil && input.DueDate == nil {
		return nil, fmt.Errorf("%w: empty patch input", ErrInvalidInput)
	}

	task, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var updateTaskInput UpdateTaskInput
	if input.DoctorID != nil {
		updateTaskInput.DoctorID = input.DoctorID
	} else {
		updateTaskInput.DoctorID = task.DoctorID
	}
	if input.Title != nil {
		updateTaskInput.Title = *input.Title
	} else {
		updateTaskInput.Title = task.Title
	}
	if input.Description != nil {
		updateTaskInput.Description = *input.Description
	} else {
		updateTaskInput.Description = task.Description
	}
	if input.Status != nil {
		updateTaskInput.Status = *input.Status
	} else {
		updateTaskInput.Status = task.Status
	}
	if input.DueDate != nil {
		updateTaskInput.DueDate = input.DueDate
	} else {
		updateTaskInput.DueDate = task.DueDate
	}

	updated, err := s.Update(ctx, id, updateTaskInput)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *TaskService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *TaskService) List(ctx context.Context, taskList taskrepository.TaskList) ([]taskdomain.Task, error) {
	return s.repo.List(ctx, taskList)
}

func validateCreateTaskInput(input CreateTaskInput) (CreateTaskInput, error) {
	if input.DoctorID != nil && *input.DoctorID <= 0 {
		return CreateTaskInput{}, fmt.Errorf("%w: doctor id must be positive", ErrInvalidInput)
	}

	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateTaskInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateTaskInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if input.GeneratorID != nil && *input.GeneratorID <= 0 {
		return CreateTaskInput{}, fmt.Errorf("%w: generator id must be positive", ErrInvalidInput)
	}

	return input, nil
}

func validateUpdateTaskInput(input UpdateTaskInput) (UpdateTaskInput, error) {
	if input.DoctorID != nil && *input.DoctorID <= 0 {
		return UpdateTaskInput{}, fmt.Errorf("%w: doctor id must be positive", ErrInvalidInput)
	}

	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateTaskInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateTaskInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}
