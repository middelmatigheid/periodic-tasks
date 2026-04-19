package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskgeneratordomain "example.com/taskservice/internal/domain/task_generator"
)

type taskGeneratorMutationDTO struct {
	DoctorID    *int64            `json:"doctor_id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	EveryNDays  *int64            `json:"every_n_days"`
	EveryIthDay *int64            `json:"every_ith_day"`
	Parity      bool              `json:"parity"`
	NextDueDate time.Time         `json:"next_due_date"`
}

type taskGeneratorPatchMutationDTO struct {
	DoctorID    *int64             `json:"doctor_id"`
	Title       *string            `json:"title"`
	Description *string            `json:"description"`
	Status      *taskdomain.Status `json:"status"`
	EveryNDays  *int64             `json:"every_n_days"`
	EveryIthDay *int64             `json:"every_ith_day"`
	Parity      *bool              `json:"parity"`
	NextDueDate *time.Time         `json:"next_due_date"`
}

type taskGeneratorDTO struct {
	ID          int64             `json:"id"`
	DoctorID    *int64            `json:"doctor_id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	EveryNDays  *int64            `json:"every_n_days"`
	EveryIthDay *int64            `json:"every_ith_day"`
	Parity      bool              `json:"parity"`
	NextDueDate time.Time         `json:"next_due_date"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	ProcessedAt time.Time         `json:"processed_at"`
}

func newTaskGeneratorDTO(taskGenerator *taskgeneratordomain.TaskGenerator) taskGeneratorDTO {
	return taskGeneratorDTO{
		ID:          taskGenerator.ID,
		DoctorID:    taskGenerator.DoctorID,
		Title:       taskGenerator.Title,
		Description: taskGenerator.Description,
		Status:      taskGenerator.Status,
		EveryNDays:  taskGenerator.EveryNDays,
		EveryIthDay: taskGenerator.EveryIthDay,
		Parity:      taskGenerator.Parity,
		NextDueDate: taskGenerator.NextDueDate,
		CreatedAt:   taskGenerator.CreatedAt,
		UpdatedAt:   taskGenerator.UpdatedAt,
		ProcessedAt: taskGenerator.ProcessedAt,
	}
}
