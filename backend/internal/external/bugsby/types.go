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
	ID                  int         `json:"id"`
	Alias               *string     `json:"alias"`
	ReportedBy          string      `json:"reportedBy"`
	ReportedTime        time.Time   `json:"reportedTime"`
	LastUpdateTime      time.Time   `json:"lastUpdateTime"`
	LastOpenedTime      time.Time   `json:"lastOpenedTime"`
	LastClosedTime      *time.Time  `json:"lastClosedTime"`
	LastDiffed          time.Time   `json:"lastDiffed"`
	Package             string      `json:"package"`
	IssueType           string      `json:"issueType"`
	Product             string      `json:"product"`
	Component           string      `json:"component"`
	Deadline            *time.Time  `json:"deadline"`
	Version             string      `json:"version"`
	ScheduleKey         *string     `json:"scheduleKey"`
	Priority            string      `json:"priority"`
	Severity            string      `json:"severity"`
	Title               string      `json:"title"`
	Assignee            string      `json:"assignee"`
	Status              string      `json:"status"`
	Resolution          string      `json:"resolution"`
	FixList             []string    `json:"fixList"`
	FixListGerrit       []string    `json:"fixListGerrit"`
	MultiRepoFixList    []string    `json:"multiRepoFixList"`
	ReviewList          []string    `json:"reviewList"`
	FixListReviewboard  interface{} `json:"fixListReviewboard"` // Can be string or number from Bugsby API
	TargetMilestone     string      `json:"targetMilestone"`
	ReleaseNote         *string     `json:"releaseNote"`
	ReleaseNoteApproval *bool       `json:"releaseNoteApproval"`
	Description         string      `json:"description"`
	EstimatedTime       float64     `json:"estimatedTime"`
	RemainingTime       float64     `json:"remainingTime"`
	Blocks              []int       `json:"blocks"`
	DependsOn           []int       `json:"dependsOn"`
	Supersedes          []int       `json:"supersedes"`
	SupersededBys       []int       `json:"supersededBys"`
	DuplicateOf         *int        `json:"duplicateOf"`
	DuplicatedBys       []int       `json:"duplicatedBys"`
	VersionsFixed       []string    `json:"versionsFixed"`
	VersionsIntroduced  []string    `json:"versionsIntroduced"`
	AffectedCategories  []int       `json:"affectedCategories"`
	AffectedPlatforms   []int       `json:"affectedPlatforms"` // Changed from []string to []int
	Watchers            []string    `json:"watchers"`
	ChainHead           *int        `json:"chainHead"`
	Chain               []int       `json:"chain"`
}

// BugsbyComment represents a comment from the Bugsby API v1
// Note: Comments API uses v1, not v3!
// The actual API response uses 'the_text' and 'epoch_time' fields
type BugsbyComment struct {
	ID        int    `json:"id"`
	BugID     int    `json:"bugId"`
	User      string `json:"user"`
	Text      string `json:"the_text"` // API uses 'the_text', not 'text'
	EpochTime int64  `json:"epoch_time"`
	RealName  string `json:"real_name"`
	IsNoisy   bool   `json:"is_noisy"`
}

// BugsbyCommentsResponse represents the response from Bugsby comments API
type BugsbyCommentsResponse struct {
	Comments []BugsbyComment `json:"comments"`
	Count    int             `json:"count,omitempty"`
	Metadata BugsbyMetadata  `json:"metadata,omitempty"`
}

// ParsedCommitInfo represents extracted commit information from gerrit comment
type ParsedCommitInfo struct {
	CommitHash  string    `json:"commit_hash"`
	GerritURL   string    `json:"gerrit_url"`
	Repository  string    `json:"repository"`
	Branch      string    `json:"branch"`
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	ChangeID    string    `json:"change_id"`
	MergedBy    string    `json:"merged_by"`
	FullText    string    `json:"full_text"`
	CommentID   int       `json:"comment_id"`
	CommentedAt time.Time `json:"commented_at"`
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
	TextQuery  string   // Elasticsearch simple query string for text search (searches alias, title, description, releaseNote, comment, attachment)
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
