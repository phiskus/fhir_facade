package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	"strings"
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

func (s *PatientStore) Update(ctx context.Context, id string, p *fhir.Patient) (*fhir.Patient, error) {
	// Check it exists and isn't deleted
	existing, err := s.Read(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil // caller returns 404
	}

	// Increment version
	currentVersion := 1
	if existing.Meta != nil {
		fmt.Sscanf(existing.Meta.VersionID, "%d", &currentVersion)
	}
	newVersion := currentVersion + 1

	p.ResourceType = "Patient"
	p.ID = id
	p.Meta = &fhir.Meta{
		VersionID:   fmt.Sprintf("%d", newVersion),
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
	}

	familyName, givenName := extractName(p)
	phone, email := extractTelecom(p)
	mrn := extractMRN(p)

	resource, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("marshal patient: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE patients SET
			version_id   = $2,
			last_updated = now(),
			family_name  = $3,
			given_name   = $4,
			birth_date   = $5,
			gender       = $6,
			active       = $7,
			mrn          = $8,
			phone        = $9,
			email        = $10,
			resource     = $11
		WHERE id = $1 AND deleted_at IS NULL`,
		id, newVersion, familyName, givenName,
		nullString(p.BirthDate), nullString(p.Gender),
		p.Active, nullString(mrn), nullString(phone), nullString(email), resource,
	)
	if err != nil {
		return nil, fmt.Errorf("update patient: %w", err)
	}

	return p, nil
}

func (s *PatientStore) Delete(ctx context.Context, id string) (bool, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE patients SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		return false, fmt.Errorf("delete patient: %w", err)
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil // false means it didn't exist
}

// --- Search ---

func (s *PatientStore) Search(ctx context.Context, params fhir.SearchParams) ([]*fhir.Patient, error) {
	query := `SELECT resource FROM patients WHERE deleted_at IS NULL`
	args := []interface{}{}
	i := 1

	if params.ID != "" {
		query += fmt.Sprintf(" AND id = $%d", i)
		args = append(args, params.ID)
		i++
	}
	if params.Name != "" {
		query += fmt.Sprintf(" AND (lower(family_name) LIKE $%d OR lower(given_name) LIKE $%d)", i, i)
		args = append(args, "%"+strings.ToLower(params.Name)+"%")
		i++
	}
	if params.Family != "" {
		query += fmt.Sprintf(" AND lower(family_name) LIKE $%d", i)
		args = append(args, strings.ToLower(params.Family)+"%")
		i++
	}
	if params.Given != "" {
		query += fmt.Sprintf(" AND lower(given_name) LIKE $%d", i)
		args = append(args, strings.ToLower(params.Given)+"%")
		i++
	}
	if params.Gender != "" {
		query += fmt.Sprintf(" AND gender = $%d", i)
		args = append(args, params.Gender)
		i++
	}
	if params.Active != "" {
		query += fmt.Sprintf(" AND active = $%d", i)
		args = append(args, params.Active == "true")
		i++
	}
	if params.Identifier != "" {
		query += fmt.Sprintf(" AND mrn = $%d", i)
		args = append(args, params.Identifier)
		i++
	}
	if params.Birthdate != "" {
		op := map[string]string{
			"eq": "=", "lt": "<", "gt": ">", "ge": ">=", "le": "<=",
		}[params.BirthdatePrefix]
		if op == "" {
			op = "="
		}
		query += fmt.Sprintf(" AND birth_date %s $%d", op, i)
		args = append(args, params.Birthdate)
		i++
	}

	query += " ORDER BY last_updated DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search patients: %w", err)
	}
	defer rows.Close()

	var patients []*fhir.Patient
	for rows.Next() {
		var resource []byte
		if err := rows.Scan(&resource); err != nil {
			return nil, fmt.Errorf("scan patient: %w", err)
		}
		var p fhir.Patient
		if err := json.Unmarshal(resource, &p); err != nil {
			return nil, fmt.Errorf("unmarshal patient: %w", err)
		}
		patients = append(patients, &p)
	}
	return patients, nil
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
