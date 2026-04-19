package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskrepository "example.com/taskservice/internal/repository/postgres/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

type TaskHandler struct {
	usecase taskusecase.TaskUsecase
}

func NewTaskHandler(usecase taskusecase.TaskUsecase) *TaskHandler {
	return &TaskHandler{usecase: usecase}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req taskMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	created, err := h.usecase.Create(r.Context(), taskusecase.CreateTaskInput{
		DoctorID:    req.DoctorID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		DueDate:     req.DueDate,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, NewTaskDTO(created))
}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := getTaskIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	task, err := h.usecase.GetByID(r.Context(), id)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, NewTaskDTO(task))
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getTaskIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req taskMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	updated, err := h.usecase.Update(r.Context(), id, taskusecase.UpdateTaskInput{
		DoctorID:    req.DoctorID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		DueDate:     req.DueDate,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, NewTaskDTO(updated))
}

func (h *TaskHandler) Patch(w http.ResponseWriter, r *http.Request) {
	id, err := getTaskIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req taskPatchMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	updated, err := h.usecase.Patch(r.Context(), id, taskusecase.PatchTaskInput{
		DoctorID:    req.DoctorID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		DueDate:     req.DueDate,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, NewTaskDTO(updated))
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getTaskIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.Delete(r.Context(), id); err != nil {
		writeUsecaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	taskList, err := getTaskListFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	tasks, err := h.usecase.List(r.Context(), *taskList)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	response := make([]taskDTO, 0, len(tasks))
	for i := range tasks {
		response = append(response, NewTaskDTO(&tasks[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func getTaskIDFromRequest(r *http.Request) (int64, error) {
	rawID := mux.Vars(r)["id"]
	if rawID == "" {
		return 0, errors.New("missing task id")
	}

	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		return 0, errors.New("invalid task id")
	}

	if id <= 0 {
		return 0, errors.New("invalid task id")
	}

	return id, nil
}

func getTaskListFromRequest(r *http.Request) (*taskrepository.TaskList, error) {
	var taskList taskrepository.TaskList

	rawDoctorID := r.URL.Query().Get("doctor_id")
	if rawDoctorID != "" {
		doctorID, err := strconv.ParseInt(rawDoctorID, 10, 64)
		if err != nil {
			return nil, errors.New("invalid doctor id")
		}

		if doctorID <= 0 {
			return nil, errors.New("invalid doctor id")
		}

		taskList.DoctorID = &doctorID
	}

	rawStatus := r.URL.Query().Get("status")
	if rawStatus != "" {
		status := taskdomain.Status(rawStatus)
		taskList.Status = &status
	}

	rawStartDate := r.URL.Query().Get("start_date")
	if rawStartDate != "" {
		startDate, err := time.Parse(time.RFC3339, rawStartDate)
		if err != nil {
			return nil, errors.New("invalid start date")
		}

		taskList.StartDate = &startDate
	}

	rawEndDate := r.URL.Query().Get("end_date")
	if rawEndDate != "" {
		endDate, err := time.Parse(time.RFC3339, rawEndDate)
		if err != nil {
			return nil, errors.New("invalid end date")
		}

		if taskList.StartDate != nil && endDate.Before(*taskList.StartDate) {
			return nil, errors.New("invalid time bounds")
		}

		taskList.EndDate = &endDate
	}

	rawPage := r.URL.Query().Get("page")
	if rawPage != "" {
		page, err := strconv.ParseInt(rawPage, 10, 64)
		if err != nil {
			return nil, err
		}
		taskList.Page = &page
	}

	rawGeneratorID := r.URL.Query().Get("generator_id")
	if rawGeneratorID != "" {
		generatorID, err := strconv.ParseInt(rawGeneratorID, 10, 64)
		if err != nil {
			return nil, errors.New("invalid generator id")
		}

		if generatorID <= 0 {
			return nil, errors.New("invalid generator id")
		}

		taskList.GeneratorID = &generatorID
	}

	return &taskList, nil
}

func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	return nil
}

func writeUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, taskdomain.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, taskusecase.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{
		"error": err.Error(),
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(payload)
}
