package bugsby

import "time"

// BugsbyResponse represents the standard response from Bugsby API
type BugsbyResponse struct {
	Bugs  []BugsbyBug `json:"bugs"`
	Count int         `json:"count,omitempty"`
	Total int         `json:"total,omitempty"`
}

// BugsbyBug represents a bug from the Bugsby API
// This structure should match the actual Bugsby API response
type BugsbyBug struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	BugType     string    `json:"bug_type"`
	CVE         string    `json:"cve,omitempty"`
	AssignedTo  string    `json:"assigned_to,omitempty"`
	Manager     string    `json:"manager,omitempty"`
	Release     string    `json:"release"`
	Component   string    `json:"component"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// Additional fields that might be in Bugsby response
	// Add more fields as you discover them from the actual API
	URL         string   `json:"url,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// BugsbyQuery represents query parameters for Bugsby API
type BugsbyQuery struct {
	Query  string // Bugsby query string (e.g., "release=='wifi-ooty' AND status=='resolved'")
	Limit  int    // Maximum number of results to return
	Offset int    // Offset for pagination
}

// BugFilters represents common filter options for querying bugs
type BugFilters struct {
	Release    string   // Filter by release name
	Status     string   // Filter by status
	Severity   []string // Filter by severity levels
	BugType    string   // Filter by bug type
	Component  string   // Filter by component
	AssignedTo string   // Filter by assigned user
	Manager    string   // Filter by manager
}

// BuildQuery constructs a Bugsby query string from filters
func (f *BugFilters) BuildQuery() string {
	var conditions []string

	if f.Release != "" {
		conditions = append(conditions, `release=="`+f.Release+`"`)
	}
	if f.Status != "" {
		conditions = append(conditions, `status=="`+f.Status+`"`)
	}
	if f.BugType != "" {
		conditions = append(conditions, `bug_type=="`+f.BugType+`"`)
	}
	if f.Component != "" {
		conditions = append(conditions, `component=="`+f.Component+`"`)
	}
	if f.AssignedTo != "" {
		conditions = append(conditions, `assigned_to=="`+f.AssignedTo+`"`)
	}
	if f.Manager != "" {
		conditions = append(conditions, `manager=="`+f.Manager+`"`)
	}
	if len(f.Severity) > 0 {
		severityList := ""
		for i, sev := range f.Severity {
			if i > 0 {
				severityList += ","
			}
			severityList += `"` + sev + `"`
		}
		conditions = append(conditions, `severity in [`+severityList+`]`)
	}

	if len(conditions) == 0 {
		return ""
	}

	query := conditions[0]
	for i := 1; i < len(conditions); i++ {
		query += " AND " + conditions[i]
	}

	return query
}

// HTTPMethod represents HTTP request methods
type HTTPMethod string

const (
	MethodGet    HTTPMethod = "GET"
	MethodPost   HTTPMethod = "POST"
	MethodPut    HTTPMethod = "PUT"
	MethodPatch  HTTPMethod = "PATCH"
	MethodDelete HTTPMethod = "DELETE"
)

// RequestOptions represents options for making HTTP requests
type RequestOptions struct {
	Method  HTTPMethod
	Endpoint string
	Params  map[string]string
	Body    interface{}
	Headers map[string]string
}

