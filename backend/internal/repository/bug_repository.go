package repository

import (
	"github.com/google/uuid"
	"github.com/omnikam04/release-notes-generator/internal/models"
	"gorm.io/gorm"
)

// BugRepository defines the interface for bug data operations
type BugRepository interface {
	Create(bug *models.Bug) error
	CreateBatch(bugs []*models.Bug) error
	FindByID(id uuid.UUID) (*models.Bug, error)
	FindByBugsbyID(bugsbyID string) (*models.Bug, error)
	Update(bug *models.Bug) error
	Delete(id uuid.UUID) error
	List(filters *BugFilters, pagination *Pagination) ([]*models.Bug, int64, error)
	FindByRelease(release string) ([]*models.Bug, error)
	BugsbyIDExists(bugsbyID string) (bool, error)
}

// BugFilters represents filter options for querying bugs
type BugFilters struct {
	Release        string
	Status         []string
	AssignedTo     *uuid.UUID
	ManagerID      *uuid.UUID
	Severity       []string
	BugType        []string
	Component      string
	HasReleaseNote *bool
	SyncStatus     string
}

// Pagination represents pagination parameters
type Pagination struct {
	Page      int
	Limit     int
	SortBy    string
	SortOrder string // "asc" or "desc"
}

// bugRepository is the concrete implementation of BugRepository
type bugRepository struct {
	db *gorm.DB
}

// NewBugRepository creates a new bug repository instance
func NewBugRepository(db *gorm.DB) BugRepository {
	return &bugRepository{db: db}
}

// Create creates a new bug
func (r *bugRepository) Create(bug *models.Bug) error {
	return r.db.Create(bug).Error
}

// CreateBatch creates multiple bugs in a single transaction
func (r *bugRepository) CreateBatch(bugs []*models.Bug) error {
	if len(bugs) == 0 {
		return nil
	}
	return r.db.Create(&bugs).Error
}

// FindByID finds a bug by its UUID
func (r *bugRepository) FindByID(id uuid.UUID) (*models.Bug, error) {
	var bug models.Bug
	err := r.db.Preload("ReleaseNote").Where("id = ?", id).First(&bug).Error
	return &bug, err
}

// FindByBugsbyID finds a bug by its Bugsby ID
func (r *bugRepository) FindByBugsbyID(bugsbyID string) (*models.Bug, error) {
	var bug models.Bug
	err := r.db.Preload("ReleaseNote").Where("bugsby_id = ?", bugsbyID).First(&bug).Error
	return &bug, err
}

// Update updates an existing bug
func (r *bugRepository) Update(bug *models.Bug) error {
	return r.db.Save(bug).Error
}

// Delete soft deletes a bug
func (r *bugRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Bug{}, "id = ?", id).Error
}

// List retrieves bugs with filters and pagination
func (r *bugRepository) List(filters *BugFilters, pagination *Pagination) ([]*models.Bug, int64, error) {
	var bugs []*models.Bug
	var total int64

	query := r.db.Model(&models.Bug{})

	// Apply filters
	if filters != nil {
		query = r.applyFilters(query, filters)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if pagination != nil {
		query = r.applyPagination(query, pagination)
	}

	// Preload relationships
	query = query.Preload("ReleaseNote")

	// Execute query
	if err := query.Find(&bugs).Error; err != nil {
		return nil, 0, err
	}

	return bugs, total, nil
}

// FindByRelease finds all bugs for a specific release
func (r *bugRepository) FindByRelease(release string) ([]*models.Bug, error) {
	var bugs []*models.Bug
	err := r.db.Preload("ReleaseNote").
		Where("release = ?", release).
		Order("created_at DESC").
		Find(&bugs).Error
	return bugs, err
}

// BugsbyIDExists checks if a bug with the given Bugsby ID exists
func (r *bugRepository) BugsbyIDExists(bugsbyID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Bug{}).Where("bugsby_id = ?", bugsbyID).Count(&count).Error
	return count > 0, err
}

// applyFilters applies filter conditions to the query
func (r *bugRepository) applyFilters(query *gorm.DB, filters *BugFilters) *gorm.DB {
	if filters.Release != "" {
		query = query.Where("release = ?", filters.Release)
	}

	if len(filters.Status) > 0 {
		query = query.Where("status IN ?", filters.Status)
	}

	if filters.AssignedTo != nil {
		query = query.Where("assigned_to = ?", *filters.AssignedTo)
	}

	if filters.ManagerID != nil {
		query = query.Where("manager_id = ?", *filters.ManagerID)
	}

	if len(filters.Severity) > 0 {
		query = query.Where("severity IN ?", filters.Severity)
	}

	if len(filters.BugType) > 0 {
		query = query.Where("bug_type IN ?", filters.BugType)
	}

	if filters.Component != "" {
		query = query.Where("component = ?", filters.Component)
	}

	if filters.SyncStatus != "" {
		query = query.Where("sync_status = ?", filters.SyncStatus)
	}

	if filters.HasReleaseNote != nil {
		if *filters.HasReleaseNote {
			query = query.Joins("INNER JOIN release_notes ON release_notes.bug_id = bugs.id AND release_notes.deleted_at IS NULL")
		} else {
			query = query.Joins("LEFT JOIN release_notes ON release_notes.bug_id = bugs.id AND release_notes.deleted_at IS NULL").
				Where("release_notes.id IS NULL")
		}
	}

	return query
}

// applyPagination applies pagination and sorting to the query
func (r *bugRepository) applyPagination(query *gorm.DB, pagination *Pagination) *gorm.DB {
	// Set defaults
	page := pagination.Page
	if page < 1 {
		page = 1
	}

	limit := pagination.Limit
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	offset := (page - 1) * limit

	// Apply sorting
	sortBy := pagination.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}

	sortOrder := pagination.SortOrder
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	query = query.Order(sortBy + " " + sortOrder)

	// Apply pagination
	return query.Offset(offset).Limit(limit)
}
