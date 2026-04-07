package fhir

// Bundle represents a FHIR R4 Bundle resource (used for search results).
type Bundle struct {
	ResourceType string        `json:"resourceType"`
	Type         string        `json:"type"` // "searchset"
	Total        int           `json:"total"`
	Entry        []BundleEntry `json:"entry,omitempty"`
}

type BundleEntry struct {
	FullURL  string   `json:"fullUrl,omitempty"`
	Resource *Patient `json:"resource,omitempty"`
}

// OperationOutcome is returned for errors (404, 400, 410, etc.)
type OperationOutcome struct {
	ResourceType string  `json:"resourceType"`
	Issue        []Issue `json:"issue"`
}

type Issue struct {
	Severity    string `json:"severity"`    // fatal | error | warning | information
	Code        string `json:"code"`        // not-found | invalid | gone | processing
	Diagnostics string `json:"diagnostics"` // human-readable message
}
