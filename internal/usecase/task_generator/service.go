package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskgeneratordomain "example.com/taskservice/internal/domain/task_generator"
	taskgeneratorrepository "example.com/taskservice/internal/repository/postgres/task_generator"
	usecasetask "example.com/taskservice/internal/usecase/task"
	"github.com/jackc/pgx/v5"
)

type TaskGeneratorService struct {
	taskService usecasetask.TaskUsecase
	repo        TaskGeneratorRepository
	runParams   RunTaskGeneratorParams
	now         func() time.Time
}

func NewTaskGeneratorService(repo TaskGeneratorRepository, taskService usecasetask.TaskUsecase, input RunTaskGeneratorParams) *TaskGeneratorService {
	return &TaskGeneratorService{
		taskService: taskService,
		repo:        repo,
		runParams:   input,
		now:         func() time.Time { return time.Now().UTC() },
	}
}

func (s *TaskGeneratorService) Run(ctx context.Context) chan struct{} {
	cancel := make(chan struct{})
	ticker := time.NewTicker(time.Duration(s.runParams.EveryNMinutes) * time.Minute)
	workerChan := make(chan int64, 100)

	go func() {
		defer close(cancel)
		defer close(workerChan)
		for {
			select {
			case <-ticker.C:
				now := time.Now()
				endDate := time.Date(now.Year(), now.Month(), now.Day()+int(s.runParams.DaysAhead), 23, 59, 59, 0, now.Location())

				taskGeneratorProcessList := taskgeneratorrepository.TaskGeneratorProcessList{EndDate: endDate, Now: s.now}
				taskGeneratorIDs, err := s.ProcessList(ctx, taskGeneratorProcessList, s.runParams.ReprocessingCooldown)
				if err != nil || len(taskGeneratorIDs) == 0 {
					continue
				}
				for _, id := range taskGeneratorIDs {
					select {
					case workerChan <- id:
					case <-cancel:
						return
					case <-ctx.Done():
						return
					}
				}
			case <-cancel:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	for range s.runParams.NumWorkers {
		go func() {
			for {
				select {
				case id, ok := <-workerChan:
					if !ok {
						return
					}
					s.Generate(ctx, id)
				case <-cancel:
					return
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	return cancel
}

func (s *TaskGeneratorService) Generate(ctx context.Context, id int64) (*taskdomain.Task, error) {
	taskGenerator, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	tx, err := s.repo.NewTx(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	taskInput := usecasetask.CreateTaskInput{
		DoctorID:    taskGenerator.DoctorID,
		Title:       taskGenerator.Title,
		Description: taskGenerator.Description,
		Status:      taskGenerator.Status,
		DueDate:     &taskGenerator.NextDueDate,
		GeneratorID: &taskGenerator.ID,
	}

	var patchInput PatchTaskGeneratorInput
	patchInput.NextDueDate = s.calculateNextDueDate(taskGenerator)

	created, err := s.taskService.CreateWithTx(ctx, taskInput, tx)
	if err != nil {
		return nil, err
	}
	_, err = s.PatchWithTx(ctx, id, patchInput, tx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *TaskGeneratorService) Create(ctx context.Context, input CreateTaskGeneratorInput) (*taskgeneratordomain.TaskGenerator, error) {
	normalized, err := validateCreateTaskGeneratorInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskgeneratordomain.TaskGenerator{
		DoctorID:    normalized.DoctorID,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		EveryNDays:  normalized.EveryNDays,
		EveryIthDay: normalized.EveryIthDay,
		Parity:      normalized.Parity,
		NextDueDate: normalized.NextDueDate,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *TaskGeneratorService) GetByID(ctx context.Context, id int64) (*taskgeneratordomain.TaskGenerator, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *TaskGeneratorService) Update(ctx context.Context, id int64, input UpdateTaskGeneratorInput) (*taskgeneratordomain.TaskGenerator, error) {
	return s.update(ctx, id, input, nil)
}

func (s *TaskGeneratorService) UpdateWithTx(ctx context.Context, id int64, input UpdateTaskGeneratorInput, tx pgx.Tx) (*taskgeneratordomain.TaskGenerator, error) {
	return s.update(ctx, id, input, tx)
}

func (s *TaskGeneratorService) update(ctx context.Context, id int64, input UpdateTaskGeneratorInput, tx pgx.Tx) (*taskgeneratordomain.TaskGenerator, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateTaskGeneratorInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskgeneratordomain.TaskGenerator{
		ID:          id,
		DoctorID:    normalized.DoctorID,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		EveryNDays:  normalized.EveryNDays,
		EveryIthDay: normalized.EveryIthDay,
		Parity:      normalized.Parity,
		NextDueDate: normalized.NextDueDate,
		UpdatedAt:   s.now(),
	}

	updated, err := s.repo.Update(ctx, model, tx)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *TaskGeneratorService) Patch(ctx context.Context, id int64, input PatchTaskGeneratorInput) (*taskgeneratordomain.TaskGenerator, error) {
	return s.patch(ctx, id, input, nil)
}

func (s *TaskGeneratorService) PatchWithTx(ctx context.Context, id int64, input PatchTaskGeneratorInput, tx pgx.Tx) (*taskgeneratordomain.TaskGenerator, error) {
	return s.patch(ctx, id, input, tx)
}

func (s *TaskGeneratorService) patch(ctx context.Context, id int64, input PatchTaskGeneratorInput, tx pgx.Tx) (*taskgeneratordomain.TaskGenerator, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	if input.DoctorID == nil && input.Title == nil && input.Description == nil && input.Status == nil && input.EveryNDays == nil && input.EveryIthDay == nil && input.Parity == nil && input.NextDueDate == nil {
		return nil, fmt.Errorf("%w: empty patch input", ErrInvalidInput)
	}

	taskGenerator, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	var updateInput UpdateTaskGeneratorInput
	if input.DoctorID != nil {
		updateInput.DoctorID = input.DoctorID
	} else {
		updateInput.DoctorID = taskGenerator.DoctorID
	}
	if input.Title != nil {
		updateInput.Title = *input.Title
	} else {
		updateInput.Title = taskGenerator.Title
	}
	if input.Description != nil {
		updateInput.Description = *input.Description
	} else {
		updateInput.Description = taskGenerator.Description
	}
	if input.Status != nil {
		updateInput.Status = *input.Status
	} else {
		updateInput.Status = taskGenerator.Status
	}
	if input.EveryNDays != nil {
		updateInput.EveryNDays = input.EveryNDays
	} else {
		updateInput.EveryNDays = taskGenerator.EveryNDays
	}
	if input.EveryIthDay != nil {
		updateInput.EveryIthDay = input.EveryIthDay
	} else {
		updateInput.EveryIthDay = taskGenerator.EveryIthDay
	}
	if input.Parity != nil {
		updateInput.Parity = *input.Parity
	} else {
		updateInput.Parity = taskGenerator.Parity
	}
	if input.NextDueDate != nil {
		updateInput.NextDueDate = *input.NextDueDate
	} else {
		updateInput.NextDueDate = taskGenerator.NextDueDate
	}

	updated, err := s.UpdateWithTx(ctx, id, updateInput, tx)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *TaskGeneratorService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *TaskGeneratorService) List(ctx context.Context, taskGeneratorList taskgeneratorrepository.TaskGeneratorList) ([]taskgeneratordomain.TaskGenerator, error) {
	return s.repo.List(ctx, taskGeneratorList)
}

func (s *TaskGeneratorService) ProcessList(ctx context.Context, taskGeneratorProcessList taskgeneratorrepository.TaskGeneratorProcessList, cooldown int64) ([]int64, error) {
	return s.repo.ProcessList(ctx, taskGeneratorProcessList, cooldown)
}

func validateCreateTaskGeneratorInput(input CreateTaskGeneratorInput) (CreateTaskGeneratorInput, error) {
	if input.DoctorID != nil && *input.DoctorID <= 0 {
		return CreateTaskGeneratorInput{}, fmt.Errorf("%w: doctor id must be positive", ErrInvalidInput)
	}

	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateTaskGeneratorInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return CreateTaskGeneratorInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if (input.EveryIthDay != nil && input.EveryNDays != nil) || (input.EveryIthDay != nil && input.Parity) || (input.Parity && input.EveryNDays != nil) {
		return CreateTaskGeneratorInput{}, fmt.Errorf("%w: only one type of periodicity should be specified", ErrInvalidInput)
	}

	if input.EveryIthDay == nil && input.EveryNDays == nil && !input.Parity {
		return CreateTaskGeneratorInput{}, fmt.Errorf("%w: one type of periodicity should be specified", ErrInvalidInput)
	}

	if input.EveryIthDay != nil && (*input.EveryIthDay <= 0 || *input.EveryIthDay > 31) {
		return CreateTaskGeneratorInput{}, fmt.Errorf("%w: invalid month's day", ErrInvalidInput)
	}

	if input.EveryNDays != nil && (*input.EveryNDays <= 0 || *input.EveryNDays > 365) {
		return CreateTaskGeneratorInput{}, fmt.Errorf("%w: invalid periodicity interval", ErrInvalidInput)
	}

	return input, nil
}

func validateUpdateTaskGeneratorInput(input UpdateTaskGeneratorInput) (UpdateTaskGeneratorInput, error) {
	if input.DoctorID != nil && *input.DoctorID <= 0 {
		return UpdateTaskGeneratorInput{}, fmt.Errorf("%w: doctor id must be positive", ErrInvalidInput)
	}

	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateTaskGeneratorInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateTaskGeneratorInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if (input.EveryIthDay != nil && input.EveryNDays != nil) || (input.EveryIthDay != nil && input.Parity) || (input.Parity && input.EveryNDays != nil) {
		return UpdateTaskGeneratorInput{}, fmt.Errorf("%w: only one type of periodicity should be specified", ErrInvalidInput)
	}

	if input.EveryIthDay == nil && input.EveryNDays == nil && !input.Parity {
		return UpdateTaskGeneratorInput{}, fmt.Errorf("%w: one type of periodicity should be specified", ErrInvalidInput)
	}

	if input.EveryIthDay != nil && (*input.EveryIthDay <= 0 || *input.EveryIthDay > 31) {
		return UpdateTaskGeneratorInput{}, fmt.Errorf("%w: invalid month's day", ErrInvalidInput)
	}

	if input.EveryNDays != nil && (*input.EveryNDays <= 0 || *input.EveryNDays > 365) {
		return UpdateTaskGeneratorInput{}, fmt.Errorf("%w: invalid periodicity interval", ErrInvalidInput)
	}

	return input, nil
}

func (s *TaskGeneratorService) calculateNextDueDate(taskGenerator *taskgeneratordomain.TaskGenerator) *time.Time {
	if taskGenerator.EveryNDays != nil {
		nextDueDate := taskGenerator.NextDueDate.AddDate(0, 0, int(*taskGenerator.EveryNDays))
		return &nextDueDate
	} else if taskGenerator.EveryIthDay != nil {
		year, month, _ := taskGenerator.NextDueDate.Date()
		hour, minute, second := taskGenerator.NextDueDate.Hour(), taskGenerator.NextDueDate.Minute(), taskGenerator.NextDueDate.Second()
		month++

		lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, s.now().Location()).Day()
		for *taskGenerator.EveryIthDay > int64(lastDay) {
			month++
			lastDay = time.Date(year, month+1, 0, 0, 0, 0, 0, s.now().Location()).Day()
		}

		nextDueDate := time.Date(year, month, int(*taskGenerator.EveryIthDay), hour, minute, second, 0, s.now().Location())
		return &nextDueDate
	} else if taskGenerator.Parity {
		nextDueDate := taskGenerator.NextDueDate.AddDate(0, 0, 1)
		for taskGenerator.NextDueDate.Day()%2 != nextDueDate.Day()%2 {
			nextDueDate = nextDueDate.AddDate(0, 0, 1)
		}
		return &nextDueDate
	} else {
		return nil
	}
}
