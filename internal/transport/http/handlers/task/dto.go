package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	DoctorID    *int64            `json:"doctor_id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	DueDate     *time.Time        `json:"due_date"`
	GeneratorID *int64            `json:"generator_id"`
}

type taskPatchMutationDTO struct {
	DoctorID    *int64             `json:"doctor_id"`
	Title       *string            `json:"title"`
	Description *string            `json:"description"`
	Status      *taskdomain.Status `json:"status"`
	DueDate     *time.Time         `json:"due_date"`
	GeneratorID *int64             `json:"generator_id"`
}

type taskDTO struct {
	ID          int64             `json:"id"`
	DoctorID    *int64            `json:"doctor_id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	DueDate     *time.Time        `json:"due_date"`
	GeneratorID *int64            `json:"generator_id"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func NewTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		DoctorID:    task.DoctorID,
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		DueDate:     task.DueDate,
		GeneratorID: task.GeneratorID,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}
