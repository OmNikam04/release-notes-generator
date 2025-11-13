package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/omnikam04/release-notes-generator/internal/external/bugsby"
	"github.com/omnikam04/release-notes-generator/internal/logger"
	"github.com/omnikam04/release-notes-generator/internal/models"
	"github.com/omnikam04/release-notes-generator/internal/repository"
	"gorm.io/gorm"
)

// BugsbySyncService handles syncing bugs from Bugsby API to our database
type BugsbySyncService interface {
	SyncRelease(ctx context.Context, release string, filters *bugsby.BugFilters) (*SyncResult, error)
	SyncBugByID(ctx context.Context, bugsbyID int) (*models.Bug, error)
	GetSyncStatus(release string) (*SyncStatus, error)
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	TotalFetched int       `json:"total_fetched"`
	NewBugs      int       `json:"new_bugs"`
	UpdatedBugs  int       `json:"updated_bugs"`
	FailedBugs   int       `json:"failed_bugs"`
	SyncedAt     time.Time `json:"synced_at"`
	Errors       []string  `json:"errors,omitempty"`
}

// SyncStatus represents the sync status for a release
type SyncStatus struct {
	Release      string     `json:"release"`
	TotalBugs    int        `json:"total_bugs"`
	SyncedBugs   int        `json:"synced_bugs"`
	PendingBugs  int        `json:"pending_bugs"`
	FailedBugs   int        `json:"failed_bugs"`
	LastSyncedAt *time.Time `json:"last_synced_at"`
}

type bugsbySyncService struct {
	bugsbyClient   bugsby.Client
	bugRepository  repository.BugRepository
	userRepository repository.UserRepository
}

// NewBugsbySyncService creates a new Bugsby sync service
func NewBugsbySyncService(
	bugsbyClient bugsby.Client,
	bugRepository repository.BugRepository,
	userRepository repository.UserRepository,
) BugsbySyncService {
	return &bugsbySyncService{
		bugsbyClient:   bugsbyClient,
		bugRepository:  bugRepository,
		userRepository: userRepository,
	}
}

// SyncRelease syncs all bugs for a specific release from Bugsby
func (s *bugsbySyncService) SyncRelease(ctx context.Context, release string, filters *bugsby.BugFilters) (*SyncResult, error) {
	logger.Info().Str("release", release).Msg("Starting Bugsby sync for release")

	result := &SyncResult{
		SyncedAt: time.Now(),
		Errors:   []string{},
	}

	// Fetch bugs from Bugsby
	bugsbyResp, err := s.bugsbyClient.GetBugsByRelease(ctx, release, filters)
	if err != nil {
		logger.Error().Err(err).Str("release", release).Msg("Failed to fetch bugs from Bugsby")
		return nil, fmt.Errorf("failed to fetch bugs from Bugsby: %w", err)
	}

	result.TotalFetched = len(bugsbyResp.Bugs)
	logger.Info().Int("count", result.TotalFetched).Msg("Fetched bugs from Bugsby")

	if result.TotalFetched == 0 {
		logger.Info().Msg("No bugs found for release")
		return result, nil
	}

	// Extract unique emails and ensure users exist
	emails := bugsby.ExtractUniqueEmails(bugsbyResp.Bugs)
	userEmailToIDMap, err := s.ensureUsersExist(emails)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to ensure users exist")
		// Continue with sync even if user mapping fails
	}

	// Process each bug
	for i := range bugsbyResp.Bugs {
		bugsbyBug := &bugsbyResp.Bugs[i]

		if err := s.syncSingleBug(bugsbyBug, userEmailToIDMap); err != nil {
			result.FailedBugs++
			result.Errors = append(result.Errors, fmt.Sprintf("Bug %d: %v", bugsbyBug.ID, err))
			logger.Error().
				Err(err).
				Int("bugsby_id", bugsbyBug.ID).
				Msg("Failed to sync bug")
			continue
		}

		// Check if it was a new bug or update
		exists, _ := s.bugRepository.BugsbyIDExists(fmt.Sprintf("%d", bugsbyBug.ID))
		if exists {
			result.UpdatedBugs++
		} else {
			result.NewBugs++
		}
	}

	logger.Info().
		Int("total", result.TotalFetched).
		Int("new", result.NewBugs).
		Int("updated", result.UpdatedBugs).
		Int("failed", result.FailedBugs).
		Msg("Bugsby sync completed")

	return result, nil
}

// SyncBugByID syncs a single bug by its Bugsby ID
func (s *bugsbySyncService) SyncBugByID(ctx context.Context, bugsbyID int) (*models.Bug, error) {
	logger.Info().Int("bugsby_id", bugsbyID).Msg("Syncing single bug from Bugsby")

	// Fetch bug from Bugsby
	bugsbyBug, err := s.bugsbyClient.GetBugByID(ctx, bugsbyID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bug from Bugsby: %w", err)
	}

	// Extract emails and ensure users exist
	emails := []string{}
	if bugsbyBug.Assignee != "" {
		emails = append(emails, bugsbyBug.Assignee)
	}
	// Note: Manager field doesn't exist in Bugsby v3 API

	userEmailToIDMap, err := s.ensureUsersExist(emails)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to ensure users exist")
	}

	// Sync the bug
	if err := s.syncSingleBug(bugsbyBug, userEmailToIDMap); err != nil {
		return nil, err
	}

	// Fetch and return the synced bug
	bug, err := s.bugRepository.FindByBugsbyID(fmt.Sprintf("%d", bugsbyID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch synced bug: %w", err)
	}

	return bug, nil
}

// GetSyncStatus returns the sync status for a release
func (s *bugsbySyncService) GetSyncStatus(release string) (*SyncStatus, error) {
	filters := &repository.BugFilters{
		Release: release,
	}

	// Get all bugs for the release
	bugs, _, err := s.bugRepository.List(filters, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get bugs for release: %w", err)
	}

	status := &SyncStatus{
		Release:   release,
		TotalBugs: len(bugs),
	}

	// Count bugs by sync status
	for _, bug := range bugs {
		switch bug.SyncStatus {
		case "synced":
			status.SyncedBugs++
		case "pending":
			status.PendingBugs++
		case "failed":
			status.FailedBugs++
		}

		// Track last synced time
		if bug.LastSyncedAt != nil {
			if status.LastSyncedAt == nil || bug.LastSyncedAt.After(*status.LastSyncedAt) {
				status.LastSyncedAt = bug.LastSyncedAt
			}
		}
	}

	return status, nil
}

// syncSingleBug syncs a single Bugsby bug to our database
func (s *bugsbySyncService) syncSingleBug(bugsbyBug *bugsby.BugsbyBug, userEmailToIDMap map[string]uuid.UUID) error {
	bugsbyIDStr := fmt.Sprintf("%d", bugsbyBug.ID)

	// Check if bug already exists
	existingBug, err := s.bugRepository.FindByBugsbyID(bugsbyIDStr)
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check if bug exists: %w", err)
	}

	if err == gorm.ErrRecordNotFound {
		// Create new bug
		newBug := bugsby.MapBugsbyBugToModel(bugsbyBug, userEmailToIDMap)
		if err := s.bugRepository.Create(newBug); err != nil {
			return fmt.Errorf("failed to create bug: %w", err)
		}
		logger.Debug().Str("bugsby_id", bugsbyIDStr).Msg("Created new bug")
	} else {
		// Update existing bug
		bugsby.MergeBugData(existingBug, bugsbyBug, userEmailToIDMap)
		if err := s.bugRepository.Update(existingBug); err != nil {
			return fmt.Errorf("failed to update bug: %w", err)
		}
		logger.Debug().Str("bugsby_id", bugsbyIDStr).Msg("Updated existing bug")
	}

	return nil
}

// ensureUsersExist ensures that users with the given emails exist in the database
// Returns a map of email -> user ID
func (s *bugsbySyncService) ensureUsersExist(emails []string) (map[string]uuid.UUID, error) {
	emailToIDMap := make(map[string]uuid.UUID)

	for _, email := range emails {
		if email == "" {
			continue
		}

		// Try to find existing user
		user, err := s.userRepository.FindByEmail(email)
		if err != nil && err != gorm.ErrRecordNotFound {
			logger.Error().Err(err).Str("email", email).Msg("Failed to find user")
			continue
		}

		if err == gorm.ErrRecordNotFound {
			// Create new user with developer role by default
			newUser := &models.User{
				Email: email,
				Role:  "developer", // Default role
			}
			if err := s.userRepository.CreateUser(newUser); err != nil {
				logger.Error().Err(err).Str("email", email).Msg("Failed to create user")
				continue
			}
			emailToIDMap[email] = newUser.ID
			logger.Debug().Str("email", email).Msg("Created new user from Bugsby sync")
		} else {
			emailToIDMap[email] = user.ID
		}
	}

	return emailToIDMap, nil
}
