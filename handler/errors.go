package handler

import (
	"encoding/json"
	"fhir_facade/fhir"
	"net/http"
)

func writeOperationOutcome(w http.ResponseWriter, status int, code, diagnostics string) {
	w.Header().Set("Content-Type", "application/fhir+json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(fhir.OperationOutcome{
		ResourceType: "OperationOutcome",
		Issue: []fhir.Issue{
			{Severity: "error", Code: code, Diagnostics: diagnostics},
		},
	})
}
