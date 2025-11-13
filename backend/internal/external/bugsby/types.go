package bugsby

import "time"

// BugsbyResponse represents the standard response from Bugsby API
type BugsbyResponse struct {
	Bugs     []BugsbyBug    `json:"bugs"`
	Count    int            `json:"count,omitempty"`
	Total    int            `json:"total,omitempty"`
	Metadata BugsbyMetadata `json:"metadata,omitempty"`
}

// BugsbyMetadata represents pagination metadata from Bugsby API
type BugsbyMetadata struct {
	HasNext bool `json:"hasNext"`
	Links   struct {
		Next string `json:"next"`
	} `json:"links"`
	Cursor int `json:"cursor"`
}

// BugsbyBug represents a bug from the Bugsby API
// This structure matches the actual Bugsby v3 API response
type BugsbyBug struct {
	ID                  int        `json:"id"`
	Alias               *string    `json:"alias"`
	ReportedBy          string     `json:"reportedBy"`
	ReportedTime        time.Time  `json:"reportedTime"`
	LastUpdateTime      time.Time  `json:"lastUpdateTime"`
	LastOpenedTime      time.Time  `json:"lastOpenedTime"`
	LastClosedTime      *time.Time `json:"lastClosedTime"`
	LastDiffed          time.Time  `json:"lastDiffed"`
	Package             string     `json:"package"`
	IssueType           string     `json:"issueType"`
	Product             string     `json:"product"`
	Component           string     `json:"component"`
	Deadline            *time.Time `json:"deadline"`
	Version             string     `json:"version"`
	ScheduleKey         *string    `json:"scheduleKey"`
	Priority            string     `json:"priority"`
	Severity            string     `json:"severity"`
	Title               string     `json:"title"`
	Assignee            string     `json:"assignee"`
	Status              string     `json:"status"`
	Resolution          string     `json:"resolution"`
	FixList             []string   `json:"fixList"`
	FixListGerrit       []string   `json:"fixListGerrit"`
	MultiRepoFixList    []string   `json:"multiRepoFixList"`
	ReviewList          []string   `json:"reviewList"`
	FixListReviewboard  *string    `json:"fixListReviewboard"`
	TargetMilestone     string     `json:"targetMilestone"`
	ReleaseNote         *string    `json:"releaseNote"`
	ReleaseNoteApproval *bool      `json:"releaseNoteApproval"`
	Description         string     `json:"description"`
	EstimatedTime       float64    `json:"estimatedTime"`
	RemainingTime       float64    `json:"remainingTime"`
	Blocks              []int      `json:"blocks"`
	DependsOn           []int      `json:"dependsOn"`
	Supersedes          []int      `json:"supersedes"`
	SupersededBys       []int      `json:"supersededBys"`
	DuplicateOf         *int       `json:"duplicateOf"`
	DuplicatedBys       []int      `json:"duplicatedBys"`
	VersionsFixed       []string   `json:"versionsFixed"`
	VersionsIntroduced  []string   `json:"versionsIntroduced"`
	AffectedCategories  []int      `json:"affectedCategories"`
	AffectedPlatforms   []string   `json:"affectedPlatforms"`
	Watchers            []string   `json:"watchers"`
	ChainHead           *int       `json:"chainHead"`
	Chain               []int      `json:"chain"`
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
	Method   HTTPMethod
	Endpoint string
	Params   map[string]string
	Body     interface{}
	Headers  map[string]string
}
