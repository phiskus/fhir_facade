package fhir

import (
	"net/url"
	"strings"
)

type SearchParams struct {
	Name      string
	Family    string
	Given     string
	Gender    string
	Birthdate string
	BirthdatePrefix string
	Identifier string
	Active    string
	ID        string
}

func ParseSearchParams(q url.Values) SearchParams {
	p := SearchParams{
		Name:       q.Get("name"),
		Family:     q.Get("family"),
		Given:      q.Get("given"),
		Gender:     q.Get("gender"),
		Identifier: q.Get("identifier"),
		Active:     q.Get("active"),
		ID:         q.Get("_id"),
	}

	// Parse birthdate prefix (eq, lt, gt, ge, le)
	if bd := q.Get("birthdate"); bd != "" {
		prefixes := []string{"eq", "lt", "gt", "ge", "le"}
		for _, prefix := range prefixes {
			if strings.HasPrefix(bd, prefix) {
				p.BirthdatePrefix = prefix
				p.Birthdate = strings.TrimPrefix(bd, prefix)
				break
			}
		}
		if p.Birthdate == "" {
			p.BirthdatePrefix = "eq"
			p.Birthdate = bd
		}
	}

	return p
}
