package router

import "regexp"

type (
	Rule struct {
		Host       string  `json:"host" jsonschema:"omitempty"`
		HostRegexp string  `json:"hostRegexp" jsonschema:"omitempty,format=regexp"`
		Paths      []*Path `json:"paths" jsonschema:"omitempty"`
	}

	// Path is second level entry of router.
	Path struct {
		Path           string    `json:"path,omitempty" jsonschema:"omitempty,pattern=^/"`
		Backend        string    `json:"backend" jsonschema:"required"`
		Methods        []string  `json:"methods,omitempty" jsonschema:"omitempty,uniqueItems=true,format=httpmethod-array"`
		Headers        []*Header `json:"headers" jsonschema:"omitempty"`
		Queries        []*Query  `json:"queries,omitempty" jsonschema:"omitempty"`
		MatchAllHeader bool      `json:"matchAllHeader" jsonschema:"omitempty"`
	}

	// Header is the third level entry of router. A header entry is always under a specific path entry, that is to mean
	// the headers entry will only be checked after a path entry matched. However, the headers entry has a higher priority
	// than the path entry itself.
	Header struct {
		headerRE *regexp.Regexp
		Key      string   `json:"key" jsonschema:"required"`
		Regexp   string   `json:"regexp,omitempty" jsonschema:"omitempty,format=regexp"`
		Values   []string `json:"values,omitempty" jsonschema:"omitempty,uniqueItems=true"`
	}

	// Query is the third level entry
	Query struct {
		re     *regexp.Regexp
		Key    string   `json:"key" jsonschema:"required"`
		Regexp string   `json:"regexp,omitempty" jsonschema:"omitempty,format=regexp"`
		Values []string `json:"values,omitempty" jsonschema:"omitempty,uniqueItems=true"`
	}
)


func (h *Header) initHeaderRoute() {
	h.headerRE = regexp.MustCompile(h.Regexp)
}

func (q *Query) initQueryRoute() {
	if q.Regexp != "" {
		q.re = regexp.MustCompile(q.Regexp)
	}
}
