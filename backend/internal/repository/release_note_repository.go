package repository

import (
	"github.com/google/uuid"
	"github.com/omnikam04/release-notes-generator/internal/models"
	"gorm.io/gorm"
)

// ReleaseNoteRepository defines the interface for release note data operations
type ReleaseNoteRepository interface {
	Create(note *models.ReleaseNote) error
	CreateBatch(notes []*models.ReleaseNote) error
	FindByID(id uuid.UUID) (*models.ReleaseNote, error)
	FindByBugID(bugID uuid.UUID) (*models.ReleaseNote, error)
	Update(note *models.ReleaseNote) error
	Delete(id uuid.UUID) error
	List(filters *ReleaseNoteFilters, pagination *Pagination) ([]*models.ReleaseNote, int64, error)
	ListPendingBugs(filters *PendingBugsFilters, pagination *Pagination) ([]*models.Bug, int64, error)
}

// ReleaseNoteFilters represents filter options for querying release notes
type ReleaseNoteFilters struct {
	BugID         *uuid.UUID
	Status        []string
	GeneratedBy   string
	CreatedByID   *uuid.UUID
	ApprovedByDev *uuid.UUID
	ApprovedByMgr *uuid.UUID
	// Bug-related filters (requires join with bugs table)
	AssignedTo *uuid.UUID // Filter by bug's assigned developer
	ManagerID  *uuid.UUID // Filter by bug's manager
	Release    string     // Filter by bug's release
	Component  string     // Filter by bug's component
}

// PendingBugsFilters represents filter options for querying bugs without release notes
type PendingBugsFilters struct {
	AssignedTo *uuid.UUID
	ManagerID  *uuid.UUID
	Release    string
	Status     []string // Bug status filter
	Severity   []string
	Component  string
}

// releaseNoteRepository is the concrete implementation of ReleaseNoteRepository
type releaseNoteRepository struct {
	db *gorm.DB
}

// NewReleaseNoteRepository creates a new release note repository instance
func NewReleaseNoteRepository(db *gorm.DB) ReleaseNoteRepository {
	return &releaseNoteRepository{db: db}
}

// Create creates a new release note
func (r *releaseNoteRepository) Create(note *models.ReleaseNote) error {
	return r.db.Create(note).Error
}

// CreateBatch creates multiple release notes in a single transaction
func (r *releaseNoteRepository) CreateBatch(notes []*models.ReleaseNote) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, note := range notes {
			if err := tx.Create(note).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// FindByID finds a release note by its ID
func (r *releaseNoteRepository) FindByID(id uuid.UUID) (*models.ReleaseNote, error) {
	var note models.ReleaseNote
	err := r.db.Preload("Bug").First(&note, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &note, nil
}

// FindByBugID finds a release note by bug ID
func (r *releaseNoteRepository) FindByBugID(bugID uuid.UUID) (*models.ReleaseNote, error) {
	var note models.ReleaseNote
	err := r.db.Preload("Bug").First(&note, "bug_id = ?", bugID).Error
	if err != nil {
		return nil, err
	}
	return &note, nil
}

// Update updates an existing release note
func (r *releaseNoteRepository) Update(note *models.ReleaseNote) error {
	return r.db.Save(note).Error
}

// Delete deletes a release note by ID
func (r *releaseNoteRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.ReleaseNote{}, "id = ?", id).Error
}

// List retrieves release notes with filters and pagination
func (r *releaseNoteRepository) List(filters *ReleaseNoteFilters, pagination *Pagination) ([]*models.ReleaseNote, int64, error) {
	var notes []*models.ReleaseNote
	var total int64

	query := r.db.Model(&models.ReleaseNote{})

	// Check if we need to join with bugs table
	needsBugJoin := false
	if filters != nil {
		if filters.AssignedTo != nil || filters.ManagerID != nil || filters.Release != "" || filters.Component != "" {
			needsBugJoin = true
		}
	}

	// Join with bugs table if needed
	if needsBugJoin {
		query = query.Joins("JOIN bugs ON bugs.id = release_notes.bug_id")
	}

	// Apply filters
	if filters != nil {
		if filters.BugID != nil {
			query = query.Where("release_notes.bug_id = ?", *filters.BugID)
		}
		if len(filters.Status) > 0 {
			query = query.Where("release_notes.status IN ?", filters.Status)
		}
		if filters.GeneratedBy != "" {
			query = query.Where("release_notes.generated_by = ?", filters.GeneratedBy)
		}
		if filters.CreatedByID != nil {
			query = query.Where("release_notes.created_by_id = ?", *filters.CreatedByID)
		}
		if filters.ApprovedByDev != nil {
			query = query.Where("release_notes.approved_by_dev_id = ?", *filters.ApprovedByDev)
		}
		if filters.ApprovedByMgr != nil {
			query = query.Where("release_notes.approved_by_mgr_id = ?", *filters.ApprovedByMgr)
		}
		// Bug-related filters
		if filters.AssignedTo != nil {
			query = query.Where("bugs.assigned_to = ?", *filters.AssignedTo)
		}
		if filters.ManagerID != nil {
			query = query.Where("bugs.manager_id = ?", *filters.ManagerID)
		}
		if filters.Release != "" {
			query = query.Where("bugs.release = ?", filters.Release)
		}
		if filters.Component != "" {
			query = query.Where("bugs.component = ?", filters.Component)
		}
	}

	// Count total (need to select distinct release_notes.id to avoid duplicates from join)
	if needsBugJoin {
		if err := query.Distinct("release_notes.id").Count(&total).Error; err != nil {
			return nil, 0, err
		}
	} else {
		if err := query.Count(&total).Error; err != nil {
			return nil, 0, err
		}
	}

	// Apply pagination
	if pagination != nil {
		offset := (pagination.Page - 1) * pagination.Limit
		query = query.Offset(offset).Limit(pagination.Limit)

		// Apply sorting
		if pagination.SortBy != "" {
			order := "release_notes." + pagination.SortBy
			if pagination.SortOrder == "desc" {
				order += " DESC"
			} else {
				order += " ASC"
			}
			query = query.Order(order)
		} else {
			query = query.Order("release_notes.created_at DESC")
		}
	}

	// Execute query with preloading
	// Need to select distinct to avoid duplicates from join
	if needsBugJoin {
		err := query.Distinct("release_notes.*").Preload("Bug").Find(&notes).Error
		return notes, total, err
	} else {
		err := query.Preload("Bug").Find(&notes).Error
		return notes, total, err
	}
}

// ListPendingBugs retrieves bugs that don't have release notes yet
func (r *releaseNoteRepository) ListPendingBugs(filters *PendingBugsFilters, pagination *Pagination) ([]*models.Bug, int64, error) {
	var bugs []*models.Bug
	var total int64

	// Query bugs that don't have release notes
	query := r.db.Model(&models.Bug{}).
		Joins("LEFT JOIN release_notes ON bugs.id = release_notes.bug_id").
		Where("release_notes.id IS NULL")

	// Apply filters
	if filters != nil {
		if filters.AssignedTo != nil {
			query = query.Where("bugs.assigned_to = ?", *filters.AssignedTo)
		}
		if filters.ManagerID != nil {
			query = query.Where("bugs.manager_id = ?", *filters.ManagerID)
		}
		if filters.Release != "" {
			query = query.Where("bugs.release = ?", filters.Release)
		}
		if len(filters.Status) > 0 {
			query = query.Where("bugs.status IN ?", filters.Status)
		}
		if len(filters.Severity) > 0 {
			query = query.Where("bugs.severity IN ?", filters.Severity)
		}
		if filters.Component != "" {
			query = query.Where("bugs.component = ?", filters.Component)
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if pagination != nil {
		offset := (pagination.Page - 1) * pagination.Limit
		query = query.Offset(offset).Limit(pagination.Limit)

		// Apply sorting
		if pagination.SortBy != "" {
			order := "bugs." + pagination.SortBy
			if pagination.SortOrder == "desc" {
				order += " DESC"
			} else {
				order += " ASC"
			}
			query = query.Order(order)
		} else {
			query = query.Order("bugs.created_at DESC")
		}
	}

	// Execute query
	err := query.Find(&bugs).Error
	return bugs, total, err
}
