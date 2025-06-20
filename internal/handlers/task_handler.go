package handlers

import (
	"encoding/json"
	"errors"
	"http_api/internal/models"
	"http_api/internal/services"
	"http_api/internal/storage"
	"net/http"
	"strconv"
	"strings"
)

type TaskHandler struct {
	service *services.TaskService
}

func NewTaskHandler(service *services.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func (h *TaskHandler) HandleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createTask(w, r)
	case http.MethodGet:
		h.listTasks(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *TaskHandler) HandleTaskByID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}
	id := parts[2]

	switch r.Method {
	case http.MethodGet:
		h.getTask(w, r, id)
	case http.MethodPut:
		h.updateTask(w, r, id)
	case http.MethodDelete:
		h.deleteTask(w, r, id)
	case http.MethodPost:
		if strings.HasSuffix(r.URL.Path, "/cancel") {
			h.cancelTask(w, r, id)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *TaskHandler) createTask(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	task, err := h.service.CreateTask(r.Context(), request.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) getTask(w http.ResponseWriter, r *http.Request, id string) {
	task, err := h.service.GetTask(r.Context(), id)
	if err != nil {
		if errors.Is(err, storage.ErrTaskNotFound) {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	respondWithJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) listTasks(w http.ResponseWriter, r *http.Request) {
	// Поддержка простой пагинации
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	tasks, err := h.service.ListTasks(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Применяем пагинацию (в реальном приложении это делалось бы в хранилище)
	start := (page - 1) * pageSize
	if start > len(tasks.Tasks) {
		start = len(tasks.Tasks)
	}
	end := start + pageSize
	if end > len(tasks.Tasks) {
		end = len(tasks.Tasks)
	}

	paginatedTasks := &models.TaskList{
		Tasks: tasks.Tasks[start:end],
		Total: tasks.Total,
	}

	respondWithJSON(w, http.StatusOK, paginatedTasks)
}

func (h *TaskHandler) updateTask(w http.ResponseWriter, r *http.Request, id string) {
	var update models.TaskUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	task, err := h.service.UpdateTask(r.Context(), id, update)
	if err != nil {
		if errors.Is(err, storage.ErrTaskNotFound) {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	respondWithJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) cancelTask(w http.ResponseWriter, r *http.Request, id string) {
	task, err := h.service.CancelTask(r.Context(), id)
	if err != nil {
		if errors.Is(err, storage.ErrTaskNotFound) {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else if errors.Is(err, storage.ErrInvalidState) {
			http.Error(w, "Task cannot be cancelled in its current state", http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	respondWithJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) deleteTask(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.service.DeleteTask(r.Context(), id); err != nil {
		if errors.Is(err, storage.ErrTaskNotFound) {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
