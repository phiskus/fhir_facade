package handler

import (
	"context"
	"encoding/json"
	"fhir_facade/fhir"
	"fmt"
	"net/http"
	"strings"

	"fhir_facade/store"
)

type PatientHandler struct {
	store *store.PatientStore
	base  string // e.g. "http://localhost:8080"
}

func NewPatientHandler(s *store.PatientStore, base string) *PatientHandler {
	return &PatientHandler{store: s, base: base}
}

// POST /Patient
func (h *PatientHandler) Create(w http.ResponseWriter, r *http.Request) {
	var p fhir.Patient
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeOperationOutcome(w, http.StatusBadRequest, "invalid", "invalid JSON: "+err.Error())
		return
	}
	if p.ResourceType != "Patient" {
		writeOperationOutcome(w, http.StatusBadRequest, "invalid", "resourceType must be Patient")
		return
	}

	created, err := h.store.Create(context.Background(), &p)
	if err != nil {
		writeOperationOutcome(w, http.StatusInternalServerError, "processing", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/fhir+json")
	w.Header().Set("Location", fmt.Sprintf("%s/Patient/%s/_history/%s", h.base, created.ID, created.Meta.VersionID))
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// GET /Patient/{id}
func (h *PatientHandler) Read(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.PathValue("id"), "")
	if id == "" {
		writeOperationOutcome(w, http.StatusBadRequest, "invalid", "missing patient id")
		return
	}

	p, err := h.store.Read(context.Background(), id)
	if err != nil {
		writeOperationOutcome(w, http.StatusInternalServerError, "processing", err.Error())
		return
	}
	if p == nil {
		writeOperationOutcome(w, http.StatusNotFound, "not-found", "Patient/"+id+" not found")
		return
	}

	w.Header().Set("Content-Type", "application/fhir+json")
	w.Header().Set("ETag", fmt.Sprintf(`W/"%s"`, p.Meta.VersionID))
	json.NewEncoder(w).Encode(p)
}

// PUT /Patient/{id}
func (h *PatientHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var p fhir.Patient
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeOperationOutcome(w, http.StatusBadRequest, "invalid", "invalid JSON: "+err.Error())
		return
	}
	if p.ResourceType != "Patient" {
		writeOperationOutcome(w, http.StatusBadRequest, "invalid", "resourceType must be Patient")
		return
	}

	updated, err := h.store.Update(context.Background(), id, &p)
	if err != nil {
		writeOperationOutcome(w, http.StatusInternalServerError, "processing", err.Error())
		return
	}
	if updated == nil {
		writeOperationOutcome(w, http.StatusNotFound, "not-found", "Patient/"+id+" not found")
		return
	}

	w.Header().Set("Content-Type", "application/fhir+json")
	w.Header().Set("ETag", fmt.Sprintf(`W/"%s"`, updated.Meta.VersionID))
	json.NewEncoder(w).Encode(updated)
}

// DELETE /Patient/{id}
func (h *PatientHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	found, err := h.store.Delete(context.Background(), id)
	if err != nil {
		writeOperationOutcome(w, http.StatusInternalServerError, "processing", err.Error())
		return
	}
	if !found {
		writeOperationOutcome(w, http.StatusNotFound, "not-found", "Patient/"+id+" not found")
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204
}
// GET /Patient?name=...
func (h *PatientHandler) Search(w http.ResponseWriter, r *http.Request) {
	params := fhir.ParseSearchParams(r.URL.Query())

	patients, err := h.store.Search(context.Background(), params)
	if err != nil {
		writeOperationOutcome(w, http.StatusInternalServerError, "processing", err.Error())
		return
	}

	bundle := fhir.Bundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Total:        len(patients),
	}
	for _, p := range patients {
		bundle.Entry = append(bundle.Entry, fhir.BundleEntry{
			FullURL:  fmt.Sprintf("%s/Patient/%s", h.base, p.ID),
			Resource: p,
		})
	}

	w.Header().Set("Content-Type", "application/fhir+json")
	json.NewEncoder(w).Encode(bundle)
}
