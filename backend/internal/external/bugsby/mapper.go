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
		BugType:      bugsbyBug.BugType,
		Release:      bugsbyBug.Release,
		Component:    bugsbyBug.Component,
		Status:       "pending", // Our internal status, not Bugsby's status
		SyncStatus:   "synced",
		LastSyncedAt: &now,
	}

	// Set description (nullable)
	if bugsbyBug.Description != "" {
		bug.Description = &bugsbyBug.Description
	}

	// Set CVE number (nullable)
	if bugsbyBug.CVE != "" {
		bug.CVENumber = &bugsbyBug.CVE
	}

	// Map AssignedTo email to user ID
	if bugsbyBug.AssignedTo != "" && userEmailToIDMap != nil {
		if userID, ok := userEmailToIDMap[bugsbyBug.AssignedTo]; ok {
			bug.AssignedTo = &userID
		}
	}

	// Map Manager email to user ID
	if bugsbyBug.Manager != "" && userEmailToIDMap != nil {
		if userID, ok := userEmailToIDMap[bugsbyBug.Manager]; ok {
			bug.ManagerID = &userID
		}
	}

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
		if bug.AssignedTo != "" {
			emailSet[bug.AssignedTo] = true
		}
		if bug.Manager != "" {
			emailSet[bug.Manager] = true
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
	existingBug.BugType = bugsbyBug.BugType
	existingBug.Release = bugsbyBug.Release
	existingBug.Component = bugsbyBug.Component
	existingBug.SyncStatus = "synced"
	existingBug.LastSyncedAt = &now

	// Update description
	if bugsbyBug.Description != "" {
		existingBug.Description = &bugsbyBug.Description
	}

	// Update CVE number
	if bugsbyBug.CVE != "" {
		existingBug.CVENumber = &bugsbyBug.CVE
	}

	// Update AssignedTo
	if bugsbyBug.AssignedTo != "" && userEmailToIDMap != nil {
		if userID, ok := userEmailToIDMap[bugsbyBug.AssignedTo]; ok {
			existingBug.AssignedTo = &userID
		}
	}

	// Update Manager
	if bugsbyBug.Manager != "" && userEmailToIDMap != nil {
		if userID, ok := userEmailToIDMap[bugsbyBug.Manager]; ok {
			existingBug.ManagerID = &userID
		}
	}

	// Note: We don't update Status (our internal status) as it's managed by our workflow
}

