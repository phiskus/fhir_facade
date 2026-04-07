package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"fhir_facade/fhir"

	"github.com/google/uuid"
)

type PatientStore struct {
	db *sql.DB
}

func NewPatientStore(db *sql.DB) *PatientStore {
	return &PatientStore{db: db}
}

func (s *PatientStore) Create(ctx context.Context, p *fhir.Patient) (*fhir.Patient, error) {
	// Assign server-generated ID and meta
	p.ResourceType = "Patient"
	p.ID = uuid.New().String()
	p.Meta = &fhir.Meta{
		VersionID:   "1",
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
	}

	// Extract indexed fields
	familyName, givenName := extractName(p)
	phone, email := extractTelecom(p)
	mrn := extractMRN(p)

	// Marshal back to JSON with injected ID and meta
	resource, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("marshal patient: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO patients
			(id, family_name, given_name, birth_date, gender, active, mrn, phone, email, resource)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		p.ID, familyName, givenName, nullString(p.BirthDate), nullString(p.Gender),
		p.Active, nullString(mrn), nullString(phone), nullString(email), resource,
	)
	if err != nil {
		return nil, fmt.Errorf("insert patient: %w", err)
	}

	return p, nil
}

func (s *PatientStore) Read(ctx context.Context, id string) (*fhir.Patient, error) {
	var resource []byte
	err := s.db.QueryRowContext(ctx,
		`SELECT resource FROM patients WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&resource)
	if err == sql.ErrNoRows {
		return nil, nil // caller checks for nil → 404
	}
	if err != nil {
		return nil, fmt.Errorf("query patient: %w", err)
	}

	var p fhir.Patient
	if err := json.Unmarshal(resource, &p); err != nil {
		return nil, fmt.Errorf("unmarshal patient: %w", err)
	}
	return &p, nil
}

// --- helpers ---

func extractName(p *fhir.Patient) (family, given string) {
	if len(p.Name) > 0 {
		family = p.Name[0].Family
		if len(p.Name[0].Given) > 0 {
			given = p.Name[0].Given[0]
		}
	}
	return
}

func extractTelecom(p *fhir.Patient) (phone, email string) {
	for _, t := range p.Telecom {
		if t.System == "phone" && phone == "" {
			phone = t.Value
		}
		if t.System == "email" && email == "" {
			email = t.Value
		}
	}
	return
}

func extractMRN(p *fhir.Patient) string {
	for _, id := range p.Identifier {
		if id.Type != nil {
			for _, c := range id.Type.Coding {
				if c.Code == "MR" {
					return id.Value
				}
			}
		}
	}
	return ""
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
