package domain

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type TaskGenerator struct {
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
