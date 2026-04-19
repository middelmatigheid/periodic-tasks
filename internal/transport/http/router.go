package transporthttp

import (
	"net/http"

	"github.com/gorilla/mux"

	swaggerdocs "example.com/taskservice/internal/transport/http/docs"
	taskhandler "example.com/taskservice/internal/transport/http/handlers/task"
	taskgeneratorhandler "example.com/taskservice/internal/transport/http/handlers/task_generator"
)

func NewRouter(taskHandler *taskhandler.TaskHandler, taskGeneratorHandler *taskgeneratorhandler.TaskGeneratorHandler, docsHandler *swaggerdocs.Handler) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/swagger/openapi.json", docsHandler.ServeSpec).Methods(http.MethodGet)
	router.HandleFunc("/swagger/", docsHandler.ServeUI).Methods(http.MethodGet)
	router.HandleFunc("/swagger", docsHandler.RedirectToUI).Methods(http.MethodGet)

	api := router.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/tasks", taskHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/tasks", taskHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.GetByID).Methods(http.MethodGet)
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.Update).Methods(http.MethodPut)
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.Patch).Methods(http.MethodPatch)
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.Delete).Methods(http.MethodDelete)

	api.HandleFunc("/task_generators", taskGeneratorHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/task_generators", taskGeneratorHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/task_generators/{id:[0-9]+}", taskGeneratorHandler.GetByID).Methods(http.MethodGet)
	api.HandleFunc("/task_generators/generate/{id:[0-9]+}", taskGeneratorHandler.Generate).Methods(http.MethodGet)
	api.HandleFunc("/task_generators/{id:[0-9]+}", taskGeneratorHandler.Update).Methods(http.MethodPut)
	api.HandleFunc("/task_generators/{id:[0-9]+}", taskGeneratorHandler.Patch).Methods(http.MethodPatch)
	api.HandleFunc("/task_generators/{id:[0-9]+}", taskGeneratorHandler.Delete).Methods(http.MethodDelete)

	return router
}
