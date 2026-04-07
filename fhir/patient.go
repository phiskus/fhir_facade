package fhir

// Patient represents a FHIR R4 Patient resource.
type Patient struct {
	ResourceType string      `json:"resourceType"`
	ID           string      `json:"id,omitempty"`
	Meta         *Meta       `json:"meta,omitempty"`
	Identifier   []Identifier `json:"identifier,omitempty"`
	Active       *bool       `json:"active,omitempty"`
	Name         []HumanName `json:"name,omitempty"`
	Telecom      []ContactPoint `json:"telecom,omitempty"`
	Gender       string      `json:"gender,omitempty"`
	BirthDate    string      `json:"birthDate,omitempty"`
	Address      []Address   `json:"address,omitempty"`
}

type Meta struct {
	VersionID   string `json:"versionId,omitempty"`
	LastUpdated string `json:"lastUpdated,omitempty"`
}

type Identifier struct {
	System string          `json:"system,omitempty"`
	Value  string          `json:"value,omitempty"`
	Type   *CodeableConcept `json:"type,omitempty"`
}

type CodeableConcept struct {
	Coding []Coding `json:"coding,omitempty"`
	Text   string   `json:"text,omitempty"`
}

type Coding struct {
	System string `json:"system,omitempty"`
	Code   string `json:"code,omitempty"`
	Display string `json:"display,omitempty"`
}

type HumanName struct {
	Use    string   `json:"use,omitempty"`
	Family string   `json:"family,omitempty"`
	Given  []string `json:"given,omitempty"`
}

type ContactPoint struct {
	System string `json:"system,omitempty"` // phone | fax | email | url
	Value  string `json:"value,omitempty"`
	Use    string `json:"use,omitempty"`    // home | work | mobile
}

type Address struct {
	Use        string   `json:"use,omitempty"`
	Line       []string `json:"line,omitempty"`
	City       string   `json:"city,omitempty"`
	State      string   `json:"state,omitempty"`
	PostalCode string   `json:"postalCode,omitempty"`
	Country    string   `json:"country,omitempty"`
}

