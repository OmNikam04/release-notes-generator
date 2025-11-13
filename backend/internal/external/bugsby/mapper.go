package bugsby

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/omnikam04/release-notes-generator/internal/models"
)

// MapBugsbyBugToModel converts a BugsbyBug to our internal Bug model
func MapBugsbyBugToModel(bugsbyBug *BugsbyBug, userEmailToIDMap map[string]uuid.UUID) *models.Bug {
	if bugsbyBug == nil {
		return nil
	}

	now := time.Now()
	bug := &models.Bug{
		BugsbyID:     strconv.Itoa(bugsbyBug.ID),
		BugsbyURL:    fmt.Sprintf("https://bugs-service.infra.corp.arista.io/v3/bugs/%d", bugsbyBug.ID),
		Title:        bugsbyBug.Title,
		Severity:     bugsbyBug.Severity,
		Priority:     bugsbyBug.Priority,
		BugType:      bugsbyBug.IssueType, // Map IssueType to BugType
		Release:      bugsbyBug.Version,   // Map Version to Release
		Component:    bugsbyBug.Component,
		Status:       "pending", // Our internal status, not Bugsby's status
		SyncStatus:   "synced",
		LastSyncedAt: &now,
	}

	// Set description (nullable)
	if bugsbyBug.Description != "" {
		bug.Description = &bugsbyBug.Description
	}

	// Note: Bugsby v3 API doesn't have CVE field directly
	// You may need to extract it from description or other fields if needed

	// Map Assignee email to user ID
	if bugsbyBug.Assignee != "" && userEmailToIDMap != nil {
		if userID, ok := userEmailToIDMap[bugsbyBug.Assignee]; ok {
			bug.AssignedTo = &userID
		}
	}

	// Note: Bugsby v3 API doesn't have Manager field
	// You may need to determine manager from other fields or leave it nil

	return bug
}

// MapBugsbyBugsToModels converts a slice of BugsbyBug to our internal Bug models
func MapBugsbyBugsToModels(bugsbyBugs []BugsbyBug, userEmailToIDMap map[string]uuid.UUID) []*models.Bug {
	bugs := make([]*models.Bug, 0, len(bugsbyBugs))

	for i := range bugsbyBugs {
		bug := MapBugsbyBugToModel(&bugsbyBugs[i], userEmailToIDMap)
		if bug != nil {
			bugs = append(bugs, bug)
		}
	}

	return bugs
}

// ExtractUniqueEmails extracts all unique email addresses from Bugsby bugs
// This is useful for creating user records before mapping
func ExtractUniqueEmails(bugsbyBugs []BugsbyBug) []string {
	emailSet := make(map[string]bool)

	for _, bug := range bugsbyBugs {
		if bug.Assignee != "" {
			emailSet[bug.Assignee] = true
		}
		if bug.ReportedBy != "" {
			emailSet[bug.ReportedBy] = true
		}
		// Add watchers
		for _, watcher := range bug.Watchers {
			if watcher != "" {
				emailSet[watcher] = true
			}
		}
	}

	emails := make([]string, 0, len(emailSet))
	for email := range emailSet {
		emails = append(emails, email)
	}

	return emails
}

// MergeBugData merges Bugsby bug data into an existing Bug model
// This is useful for updating existing bugs without losing our internal data
func MergeBugData(existingBug *models.Bug, bugsbyBug *BugsbyBug, userEmailToIDMap map[string]uuid.UUID) {
	if existingBug == nil || bugsbyBug == nil {
		return
	}

	now := time.Now()

	// Update fields from Bugsby
	existingBug.Title = bugsbyBug.Title
	existingBug.Severity = bugsbyBug.Severity
	existingBug.Priority = bugsbyBug.Priority
	existingBug.BugType = bugsbyBug.IssueType // Map IssueType to BugType
	existingBug.Release = bugsbyBug.Version   // Map Version to Release
	existingBug.Component = bugsbyBug.Component
	existingBug.SyncStatus = "synced"
	existingBug.LastSyncedAt = &now

	// Update description
	if bugsbyBug.Description != "" {
		existingBug.Description = &bugsbyBug.Description
	}

	// Update Assignee
	if bugsbyBug.Assignee != "" && userEmailToIDMap != nil {
		if userID, ok := userEmailToIDMap[bugsbyBug.Assignee]; ok {
			existingBug.AssignedTo = &userID
		}
	}

	// Note: We don't update Status (our internal status) as it's managed by our workflow
}
