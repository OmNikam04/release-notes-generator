package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/omnikam04/release-notes-generator/internal/dto"
	"github.com/omnikam04/release-notes-generator/internal/external/bugsby"
	"github.com/omnikam04/release-notes-generator/internal/logger"
	"github.com/omnikam04/release-notes-generator/internal/repository"
	"github.com/omnikam04/release-notes-generator/internal/service"
)

type BugHandler struct {
	bugsbySyncService service.BugsbySyncService
	bugRepository     repository.BugRepository
}

func NewBugHandler(bugsbySyncService service.BugsbySyncService, bugRepository repository.BugRepository) *BugHandler {
	return &BugHandler{
		bugsbySyncService: bugsbySyncService,
		bugRepository:     bugRepository,
	}
}

// SyncRelease syncs bugs for a release from Bugsby
// POST /api/v1/bugsby/sync
func (h *BugHandler) SyncRelease(c *fiber.Ctx) error {
	var req dto.SyncReleaseRequest

	if err := c.BodyParser(&req); err != nil {
		logger.Error().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate request
	if err := ValidateStruct(c, &req); err != nil {
		return err
	}

	// Build Bugsby filters
	filters := &bugsby.BugFilters{
		Status:    req.Status,
		Severity:  req.Severity,
		BugType:   req.BugType,
		Component: req.Component,
	}

	// Perform sync
	result, err := h.bugsbySyncService.SyncRelease(c.Context(), req.Release, filters)
	if err != nil {
		logger.Error().Err(err).Str("release", req.Release).Msg("Failed to sync release")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "sync_failed",
			Message: err.Error(),
		})
	}

	// Convert to response DTO
	response := &dto.SyncResultResponse{
		TotalFetched: result.TotalFetched,
		NewBugs:      result.NewBugs,
		UpdatedBugs:  result.UpdatedBugs,
		FailedBugs:   result.FailedBugs,
		SyncedAt:     result.SyncedAt,
		Errors:       result.Errors,
	}

	logger.Info().
		Str("release", req.Release).
		Int("total", result.TotalFetched).
		Int("new", result.NewBugs).
		Int("updated", result.UpdatedBugs).
		Msg("Release sync completed")

	return c.Status(fiber.StatusOK).JSON(dto.SuccessResponse{
		Success: true,
		Message: "Release synced successfully",
		Data:    response,
	})
}

// SyncBugByID syncs a single bug by its Bugsby ID
// POST /api/v1/bugsby/sync/:bugsby_id
func (h *BugHandler) SyncBugByID(c *fiber.Ctx) error {
	bugsbyIDStr := c.Params("bugsby_id")
	bugsbyID, err := strconv.Atoi(bugsbyIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_bugsby_id",
			Message: "Bugsby ID must be a valid integer",
		})
	}

	// Perform sync
	bug, err := h.bugsbySyncService.SyncBugByID(c.Context(), bugsbyID)
	if err != nil {
		logger.Error().Err(err).Int("bugsby_id", bugsbyID).Msg("Failed to sync bug")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "sync_failed",
			Message: err.Error(),
		})
	}

	logger.Info().Int("bugsby_id", bugsbyID).Msg("Bug synced successfully")

	return c.Status(fiber.StatusOK).JSON(dto.SuccessResponse{
		Success: true,
		Message: "Bug synced successfully",
		Data:    dto.ToBugResponse(bug),
	})
}

// GetSyncStatus gets the sync status for a release
// GET /api/v1/bugsby/status?release=wifi-ooty
func (h *BugHandler) GetSyncStatus(c *fiber.Ctx) error {
	release := c.Query("release")
	if release == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "missing_release",
			Message: "Release parameter is required",
		})
	}

	status, err := h.bugsbySyncService.GetSyncStatus(release)
	if err != nil {
		logger.Error().Err(err).Str("release", release).Msg("Failed to get sync status")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "status_failed",
			Message: err.Error(),
		})
	}

	response := &dto.SyncStatusResponse{
		Release:      status.Release,
		TotalBugs:    status.TotalBugs,
		SyncedBugs:   status.SyncedBugs,
		PendingBugs:  status.PendingBugs,
		FailedBugs:   status.FailedBugs,
		LastSyncedAt: status.LastSyncedAt,
	}

	return c.Status(fiber.StatusOK).JSON(dto.SuccessResponse{
		Success: true,
		Data:    response,
	})
}

// ListBugs lists bugs with filters and pagination
// GET /api/v1/bugs
func (h *BugHandler) ListBugs(c *fiber.Ctx) error {
	var filterReq dto.BugFiltersRequest

	// Parse query parameters
	if err := c.QueryParser(&filterReq); err != nil {
		logger.Error().Err(err).Msg("Failed to parse query parameters")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_query",
			Message: "Invalid query parameters",
		})
	}

	// Build repository filters
	filters := &repository.BugFilters{
		Release:        filterReq.Release,
		Status:         filterReq.Status,
		Severity:       filterReq.Severity,
		BugType:        filterReq.BugType,
		Component:      filterReq.Component,
		HasReleaseNote: filterReq.HasReleaseNote,
	}

	// Parse UUID filters
	if filterReq.AssignedTo != "" {
		if assignedToID, err := uuid.Parse(filterReq.AssignedTo); err == nil {
			filters.AssignedTo = &assignedToID
		}
	}
	if filterReq.ManagerID != "" {
		if managerID, err := uuid.Parse(filterReq.ManagerID); err == nil {
			filters.ManagerID = &managerID
		}
	}

	// Build pagination
	pagination := &repository.Pagination{
		Page:      filterReq.Page,
		Limit:     filterReq.Limit,
		SortBy:    filterReq.SortBy,
		SortOrder: filterReq.SortOrder,
	}

	// Fetch bugs
	bugs, total, err := h.bugRepository.List(filters, pagination)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to list bugs")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "list_failed",
			Message: "Failed to retrieve bugs",
		})
	}

	// Convert to response
	response := dto.ToBugListResponse(bugs, total, pagination.Page, pagination.Limit)

	return c.Status(fiber.StatusOK).JSON(dto.SuccessResponse{
		Success: true,
		Data:    response,
	})
}

// GetBug gets a single bug by ID
// GET /api/v1/bugs/:id
func (h *BugHandler) GetBug(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid bug ID",
		})
	}

	bug, err := h.bugRepository.FindByID(id)
	if err != nil {
		logger.Error().Err(err).Str("bug_id", idStr).Msg("Bug not found")
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "not_found",
			Message: "Bug not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(dto.SuccessResponse{
		Success: true,
		Data:    dto.ToBugResponse(bug),
	})
}

// UpdateBug updates a bug
// PATCH /api/v1/bugs/:id
func (h *BugHandler) UpdateBug(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid bug ID",
		})
	}

	var req dto.UpdateBugRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Error().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Fetch existing bug
	bug, err := h.bugRepository.FindByID(id)
	if err != nil {
		logger.Error().Err(err).Str("bug_id", idStr).Msg("Bug not found")
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "not_found",
			Message: "Bug not found",
		})
	}

	// Update fields
	if req.Status != nil {
		bug.Status = *req.Status
	}
	if req.AssignedTo != nil {
		bug.AssignedTo = req.AssignedTo
	}
	if req.ManagerID != nil {
		bug.ManagerID = req.ManagerID
	}

	// Save changes
	if err := h.bugRepository.Update(bug); err != nil {
		logger.Error().Err(err).Str("bug_id", idStr).Msg("Failed to update bug")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "update_failed",
			Message: "Failed to update bug",
		})
	}

	logger.Info().Str("bug_id", idStr).Msg("Bug updated successfully")

	return c.Status(fiber.StatusOK).JSON(dto.SuccessResponse{
		Success: true,
		Message: "Bug updated successfully",
		Data:    dto.ToBugResponse(bug),
	})
}

// DeleteBug soft deletes a bug
// DELETE /api/v1/bugs/:id
func (h *BugHandler) DeleteBug(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid bug ID",
		})
	}

	if err := h.bugRepository.Delete(id); err != nil {
		logger.Error().Err(err).Str("bug_id", idStr).Msg("Failed to delete bug")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "delete_failed",
			Message: "Failed to delete bug",
		})
	}

	logger.Info().Str("bug_id", idStr).Msg("Bug deleted successfully")

	return c.Status(fiber.StatusOK).JSON(dto.SuccessResponse{
		Success: true,
		Message: "Bug deleted successfully",
	})
}
