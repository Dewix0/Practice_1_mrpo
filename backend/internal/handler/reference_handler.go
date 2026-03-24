package handler

import (
	"net/http"

	"shoe-store/internal/repository"
)

type ReferenceHandler struct {
	Repo *repository.ReferenceRepo
}

func NewReferenceHandler(repo *repository.ReferenceRepo) *ReferenceHandler {
	return &ReferenceHandler{Repo: repo}
}

func (h *ReferenceHandler) Categories(w http.ResponseWriter, r *http.Request) {
	data, err := h.Repo.ListCategories()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *ReferenceHandler) Manufacturers(w http.ResponseWriter, r *http.Request) {
	data, err := h.Repo.ListManufacturers()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *ReferenceHandler) Suppliers(w http.ResponseWriter, r *http.Request) {
	data, err := h.Repo.ListSuppliers()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *ReferenceHandler) Units(w http.ResponseWriter, r *http.Request) {
	data, err := h.Repo.ListUnits()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *ReferenceHandler) OrderStatuses(w http.ResponseWriter, r *http.Request) {
	data, err := h.Repo.ListOrderStatuses()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *ReferenceHandler) PickupPoints(w http.ResponseWriter, r *http.Request) {
	data, err := h.Repo.ListPickupPoints()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}
