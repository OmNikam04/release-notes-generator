package handlers

import (
	"encoding/json"
	"fmt"
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
	bugsbyClient      bugsby.Client
}

func NewBugHandler(bugsbySyncService service.BugsbySyncService, bugRepository repository.BugRepository, bugsbyClient bugsby.Client) *BugHandler {
	return &BugHandler{
		bugsbySyncService: bugsbySyncService,
		bugRepository:     bugRepository,
		bugsbyClient:      bugsbyClient,
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

// SyncByQuery syncs bugs using a custom Bugsby query
// POST /api/v1/bugsby/sync-by-query
func (h *BugHandler) SyncByQuery(c *fiber.Ctx) error {
	var req dto.SyncByQueryRequest

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

	// Set default limit if not provided
	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}

	// Perform sync
	result, err := h.bugsbySyncService.SyncByQuery(c.Context(), req.Query, limit)
	if err != nil {
		logger.Error().Err(err).Str("query", req.Query).Msg("Failed to sync bugs by query")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "sync_failed",
			Message: err.Error(),
		})
	}

	logger.Info().
		Str("query", req.Query).
		Int("total", result.TotalFetched).
		Int("new", result.NewBugs).
		Int("updated", result.UpdatedBugs).
		Int("failed", result.FailedBugs).
		Msg("Bugs synced successfully by query")

	response := &dto.SyncResultResponse{
		TotalFetched: result.TotalFetched,
		NewBugs:      result.NewBugs,
		UpdatedBugs:  result.UpdatedBugs,
		FailedBugs:   result.FailedBugs,
		SyncedAt:     result.SyncedAt,
		Errors:       result.Errors,
	}

	return c.Status(fiber.StatusOK).JSON(dto.SuccessResponse{
		Success: true,
		Message: "Bugs synced successfully",
		Data:    response,
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

// GetBugsByAssignee fetches bugs from Bugsby API for a specific assignee
// GET /api/v1/bugsby/bugs/assignee/:email
// Query params: limit, sortBy, order, cursor (all optional)
func (h *BugHandler) GetBugsByAssignee(c *fiber.Ctx) error {
	email := c.Params("email")
	if email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "missing_email",
			Message: "Email parameter is required",
		})
	}

	// Build query parameters with full control
	params := map[string]string{
		"q":                     fmt.Sprintf("assignee==%s", email),
		"limit":                 c.Query("limit", "100"),
		"sortBy":                c.Query("sortBy", "id"),
		"order":                 c.Query("order", "asc"),
		"source":                c.Query("source", "mysql"),
		"textQueryMode":         c.Query("textQueryMode", "default"),
		"auxiliaryUserLimit":    c.Query("auxiliaryUserLimit", "200"),
		"auxiliaryProductLimit": c.Query("auxiliaryProductLimit", "200"),
		"auxiliaryPackageLimit": c.Query("auxiliaryPackageLimit", "200"),
		"auxiliaryBugLimit":     c.Query("auxiliaryBugLimit", "200"),
		"auxiliaryReleaseLimit": c.Query("auxiliaryReleaseLimit", "200"),
		"auxiliaryBugTagLimit":  c.Query("auxiliaryBugTagLimit", "200"),
	}

	// Add cursor if provided (for pagination)
	if cursor := c.Query("cursor"); cursor != "" {
		params["cursor"] = cursor
	}

	logger.Info().
		Str("email", email).
		Str("limit", params["limit"]).
		Str("sortBy", params["sortBy"]).
		Msg("Fetching bugs from Bugsby API")

	// Make GET request to Bugsby API
	resp, err := h.bugsbyClient.Get(c.Context(), "bugs", params)
	if err != nil {
		logger.Error().Err(err).Str("email", email).Msg("Failed to fetch bugs from Bugsby")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "bugsby_fetch_failed",
			Message: fmt.Sprintf("Failed to fetch bugs from Bugsby: %v", err),
		})
	}
	defer resp.Body.Close()

	// Parse Bugsby response
	var bugsbyResp bugsby.BugsbyResponse
	if err := json.NewDecoder(resp.Body).Decode(&bugsbyResp); err != nil {
		logger.Error().Err(err).Msg("Failed to decode Bugsby response")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "decode_failed",
			Message: "Failed to parse Bugsby response",
		})
	}

	logger.Info().
		Str("email", email).
		Int("bugs_returned", len(bugsbyResp.Bugs)).
		Bool("has_next", bugsbyResp.Metadata.HasNext).
		Msg("Successfully fetched bugs from Bugsby")

	// Return response with metadata
	// Note: Bugsby API doesn't return total count, only paginated results
	return c.Status(fiber.StatusOK).JSON(dto.SuccessResponse{
		Success: true,
		Message: fmt.Sprintf("Found %d bugs for %s (page result)", len(bugsbyResp.Bugs), email),
		Data: fiber.Map{
			"bugs":      bugsbyResp.Bugs,
			"count":     len(bugsbyResp.Bugs),
			"has_next":  bugsbyResp.Metadata.HasNext,
			"cursor":    bugsbyResp.Metadata.Cursor,
			"next_link": bugsbyResp.Metadata.Links.Next,
		},
	})
}

// GetBugsByCustomQuery allows testing any Bugsby query
// POST /api/v1/bugsby/bugs/query
// Body: { "query": "assignee==john.doe@arista.com AND status==ASSIGNED", "limit": 50, ... }
func (h *BugHandler) GetBugsByCustomQuery(c *fiber.Ctx) error {
	var req struct {
		Query                 string `json:"query" validate:"required"`
		Limit                 string `json:"limit"`
		SortBy                string `json:"sortBy"`
		Order                 string `json:"order"`
		Source                string `json:"source"`
		TextQueryMode         string `json:"textQueryMode"`
		AuxiliaryUserLimit    string `json:"auxiliaryUserLimit"`
		AuxiliaryProductLimit string `json:"auxiliaryProductLimit"`
		AuxiliaryPackageLimit string `json:"auxiliaryPackageLimit"`
		AuxiliaryBugLimit     string `json:"auxiliaryBugLimit"`
		AuxiliaryReleaseLimit string `json:"auxiliaryReleaseLimit"`
		AuxiliaryBugTagLimit  string `json:"auxiliaryBugTagLimit"`
		Cursor                string `json:"cursor"`
	}

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

	// Build query parameters with defaults
	params := map[string]string{
		"q": req.Query,
	}

	// Add optional parameters
	if req.Limit != "" {
		params["limit"] = req.Limit
	} else {
		params["limit"] = "100"
	}
	if req.SortBy != "" {
		params["sortBy"] = req.SortBy
	}
	if req.Order != "" {
		params["order"] = req.Order
	}
	if req.Source != "" {
		params["source"] = req.Source
	}
	if req.TextQueryMode != "" {
		params["textQueryMode"] = req.TextQueryMode
	}
	if req.AuxiliaryUserLimit != "" {
		params["auxiliaryUserLimit"] = req.AuxiliaryUserLimit
	}
	if req.AuxiliaryProductLimit != "" {
		params["auxiliaryProductLimit"] = req.AuxiliaryProductLimit
	}
	if req.AuxiliaryPackageLimit != "" {
		params["auxiliaryPackageLimit"] = req.AuxiliaryPackageLimit
	}
	if req.AuxiliaryBugLimit != "" {
		params["auxiliaryBugLimit"] = req.AuxiliaryBugLimit
	}
	if req.AuxiliaryReleaseLimit != "" {
		params["auxiliaryReleaseLimit"] = req.AuxiliaryReleaseLimit
	}
	if req.AuxiliaryBugTagLimit != "" {
		params["auxiliaryBugTagLimit"] = req.AuxiliaryBugTagLimit
	}
	if req.Cursor != "" {
		params["cursor"] = req.Cursor
	}

	logger.Info().
		Str("query", req.Query).
		Str("limit", params["limit"]).
		Msg("Executing custom Bugsby query")

	// Make GET request to Bugsby API
	resp, err := h.bugsbyClient.Get(c.Context(), "bugs", params)
	if err != nil {
		logger.Error().Err(err).Str("query", req.Query).Msg("Failed to execute Bugsby query")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "bugsby_query_failed",
			Message: fmt.Sprintf("Failed to execute Bugsby query: %v", err),
		})
	}
	defer resp.Body.Close()

	// Parse Bugsby response
	var bugsbyResp bugsby.BugsbyResponse
	if err := json.NewDecoder(resp.Body).Decode(&bugsbyResp); err != nil {
		logger.Error().Err(err).Msg("Failed to decode Bugsby response")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "decode_failed",
			Message: "Failed to parse Bugsby response",
		})
	}

	logger.Info().
		Str("query", req.Query).
		Int("bugs_returned", len(bugsbyResp.Bugs)).
		Bool("has_next", bugsbyResp.Metadata.HasNext).
		Msg("Successfully executed Bugsby query")

	// Return response with metadata
	// Note: Bugsby API doesn't return total count, only paginated results
	return c.Status(fiber.StatusOK).JSON(dto.SuccessResponse{
		Success: true,
		Message: fmt.Sprintf("Found %d bugs (page result)", len(bugsbyResp.Bugs)),
		Data: fiber.Map{
			"bugs":      bugsbyResp.Bugs,
			"count":     len(bugsbyResp.Bugs),
			"has_next":  bugsbyResp.Metadata.HasNext,
			"cursor":    bugsbyResp.Metadata.Cursor,
			"next_link": bugsbyResp.Metadata.Links.Next,
			"query":     req.Query,
		},
	})
}
