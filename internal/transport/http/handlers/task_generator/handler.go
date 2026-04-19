package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskgeneratordomain "example.com/taskservice/internal/domain/task_generator"
	taskgeneratorrepository "example.com/taskservice/internal/repository/postgres/task_generator"
	taskhandler "example.com/taskservice/internal/transport/http/handlers/task"
	taskgeneratorusecase "example.com/taskservice/internal/usecase/task_generator"
)

type TaskGeneratorHandler struct {
	usecase taskgeneratorusecase.TaskGeneratorUsecase
}

func NewTaskGeneratorHandler(usecase taskgeneratorusecase.TaskGeneratorUsecase) *TaskGeneratorHandler {
	return &TaskGeneratorHandler{usecase: usecase}
}

func (h *TaskGeneratorHandler) Generate(w http.ResponseWriter, r *http.Request) {
	id, err := getTaskGeneratorIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	generated, err := h.usecase.Generate(r.Context(), id)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, taskhandler.NewTaskDTO(generated))
}

func (h *TaskGeneratorHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req taskGeneratorMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	created, err := h.usecase.Create(r.Context(), taskgeneratorusecase.CreateTaskGeneratorInput{
		DoctorID:    req.DoctorID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		EveryNDays:  req.EveryNDays,
		EveryIthDay: req.EveryIthDay,
		Parity:      req.Parity,
		NextDueDate: req.NextDueDate,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newTaskGeneratorDTO(created))
}

func (h *TaskGeneratorHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := getTaskGeneratorIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	task, err := h.usecase.GetByID(r.Context(), id)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTaskGeneratorDTO(task))
}

func (h *TaskGeneratorHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getTaskGeneratorIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req taskGeneratorMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	updated, err := h.usecase.Update(r.Context(), id, taskgeneratorusecase.UpdateTaskGeneratorInput{
		DoctorID:    req.DoctorID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		EveryNDays:  req.EveryNDays,
		EveryIthDay: req.EveryIthDay,
		Parity:      req.Parity,
		NextDueDate: req.NextDueDate,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTaskGeneratorDTO(updated))
}

func (h *TaskGeneratorHandler) Patch(w http.ResponseWriter, r *http.Request) {
	id, err := getTaskGeneratorIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req taskGeneratorPatchMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	updated, err := h.usecase.Patch(r.Context(), id, taskgeneratorusecase.PatchTaskGeneratorInput{
		DoctorID:    req.DoctorID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		EveryNDays:  req.EveryNDays,
		EveryIthDay: req.EveryIthDay,
		Parity:      req.Parity,
		NextDueDate: req.NextDueDate,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTaskGeneratorDTO(updated))
}

func (h *TaskGeneratorHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getTaskGeneratorIDFromRequest(r)
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

func (h *TaskGeneratorHandler) List(w http.ResponseWriter, r *http.Request) {
	taskGeneratorList, err := getTaskGeneratorListFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	taskGenerators, err := h.usecase.List(r.Context(), *taskGeneratorList)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	response := make([]taskGeneratorDTO, 0, len(taskGenerators))
	for i := range taskGenerators {
		response = append(response, newTaskGeneratorDTO(&taskGenerators[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func getTaskGeneratorIDFromRequest(r *http.Request) (int64, error) {
	rawID := mux.Vars(r)["id"]
	if rawID == "" {
		return 0, errors.New("missing task generator id")
	}

	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		return 0, errors.New("invalid task generator id")
	}

	if id <= 0 {
		return 0, errors.New("invalid task generator id")
	}

	return id, nil
}

func getTaskGeneratorListFromRequest(r *http.Request) (*taskgeneratorrepository.TaskGeneratorList, error) {
	var taskGeneratorList taskgeneratorrepository.TaskGeneratorList

	rawDoctorID := r.URL.Query().Get("doctor_id")
	if rawDoctorID != "" {
		doctorID, err := strconv.ParseInt(rawDoctorID, 10, 64)
		if err != nil {
			return nil, errors.New("invalid doctor id")
		}

		if doctorID <= 0 {
			return nil, errors.New("invalid doctor id")
		}

		taskGeneratorList.DoctorID = &doctorID
	}

	rawStatus := r.URL.Query().Get("status")
	if rawStatus != "" {
		status := taskdomain.Status(rawStatus)
		taskGeneratorList.Status = &status
	}

	rawStartDate := r.URL.Query().Get("start_date")
	if rawStartDate != "" {
		startDate, err := time.Parse(time.RFC3339, rawStartDate)
		if err != nil {
			return nil, errors.New("invalid start date")
		}

		taskGeneratorList.StartDate = &startDate
	}

	rawEndDate := r.URL.Query().Get("end_date")
	if rawEndDate != "" {
		endDate, err := time.Parse(time.RFC3339, rawEndDate)
		if err != nil {
			return nil, errors.New("invalid end date")
		}

		if taskGeneratorList.StartDate != nil && endDate.Before(*taskGeneratorList.StartDate) {
			return nil, errors.New("invalid time bounds")
		}

		taskGeneratorList.EndDate = &endDate
	}

	rawPage := r.URL.Query().Get("page")
	if rawPage != "" {
		page, err := strconv.ParseInt(rawPage, 10, 64)
		if err != nil {
			return nil, err
		}
		taskGeneratorList.Page = &page
	}

	return &taskGeneratorList, nil
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
	case errors.Is(err, taskgeneratordomain.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, taskgeneratorusecase.ErrInvalidInput):
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
