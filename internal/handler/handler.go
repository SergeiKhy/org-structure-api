package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/SergeiKhy/org-structure-api/internal/model"
	"github.com/SergeiKhy/org-structure-api/internal/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(s *service.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *Handler) WriteError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}

func (h *Handler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	var req model.CreateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}

	dept, err := h.service.CreateDepartment(req)
	if err != nil {
		h.WriteError(w, http.StatusBadGateway, err.Error())
		return
	}

	h.writeJSON(w, http.StatusCreated, dept)
}

func (h *Handler) UpdateDepartment(w http.ResponseWriter, r *http.Request) {
	// Извлекаем ID из URL (/departments/{id})
	idStr := strings.TrimPrefix(r.URL.Path, "/departments/")
	idStr = strings.Split(idStr, "/")[0]

	// В main.go путь зарегистрирован как /departments/{id}, но net/http не поддерживает паттерны {id}
	// Поэтому в main.go будет использовать /departments/ и передовать ID другим способом

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req model.UpdateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}

	dept, err := h.service.UpdateDepartment(id, req)
	if err != nil {
		if err == service.ErrNotFound {
			h.WriteError(w, http.StatusNotFound, err.Error())
		} else if err == service.ErrCycleDetected || err == service.ErrSelfParent {
			h.WriteError(w, http.StatusConflict, err.Error())
		} else {
			h.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	h.writeJSON(w, http.StatusOK, dept)
}

func (h *Handler) DeleteDepartment(w http.ResponseWriter, r *http.Request) {
	// Парсинг ID аналогично Update
	idStr := strings.TrimPrefix(r.URL.Path, "/departments/")
	idStr = strings.Split(idStr, "/")[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "cascade"
	}

	var reassignToID *int
	if mode == "reassign" {
		val := r.URL.Query().Get("reassign_to_department_id")
		if val == "" {
			h.WriteError(w, http.StatusBadRequest, "reassign_to_department_id required")
			return
		}
		idVal, err := strconv.Atoi(val)
		if err != nil {
			h.WriteError(w, http.StatusBadRequest, "invalid reassign id")
			return
		}
		reassignToID = &idVal
	}

	if err := h.service.DeleteDepartment(id, mode, reassignToID); err != nil {
		if err == service.ErrNotFound {
			h.WriteError(w, http.StatusNotFound, err.Error())
		} else {
			h.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetDepartment(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/departments/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	depth := 1
	if d := r.URL.Query().Get("depth"); d != "" {
		if val, err := strconv.Atoi(d); err == nil {
			depth = val
		}
	}

	includeEmployees := true
	if ie := r.URL.Query().Get("include_employees"); ie == "false" {
		includeEmployees = false
	}

	dept, err := h.service.GetDepartmentTree(id, depth, includeEmployees)
	if err != nil {
		h.WriteError(w, http.StatusNotFound, "not found")
		return
	}

	h.writeJSON(w, http.StatusOK, dept)
}

func (h *Handler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	// Путь: /departments/{id}/employees/
	// Парсинг ID департмента из пути
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	// Ожидаем: departments, {id}, employees
	if len(parts) < 3 {
		h.WriteError(w, http.StatusBadRequest, "invalid path")
		return
	}
	deptID, err := strconv.Atoi(parts[1])
	if err != nil {
		h.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req model.CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}

	emp, err := h.service.CreateEmployee(deptID, req)
	if err != nil {
		if err == service.ErrNotFound {
			h.WriteError(w, http.StatusNotFound, err.Error())
		} else {
			h.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	h.writeJSON(w, http.StatusCreated, emp)
}
