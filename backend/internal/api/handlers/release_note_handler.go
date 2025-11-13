package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/omnikam04/release-notes-generator/internal/dto"
	"github.com/omnikam04/release-notes-generator/internal/logger"
	"github.com/omnikam04/release-notes-generator/internal/repository"
	"github.com/omnikam04/release-notes-generator/internal/service"
)

type ReleaseNoteHandler struct {
	releaseNoteService service.ReleaseNoteService
}

func NewReleaseNoteHandler(releaseNoteService service.ReleaseNoteService) *ReleaseNoteHandler {
	return &ReleaseNoteHandler{
		releaseNoteService: releaseNoteService,
	}
}

// GetPendingBugs gets bugs without release notes
// GET /api/v1/release-notes/pending
func (h *ReleaseNoteHandler) GetPendingBugs(c *fiber.Ctx) error {
	// Get current user from context
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}

	// Parse query parameters
	var req dto.GetPendingBugsRequest
	if err := c.QueryParser(&req); err != nil {
		logger.Error().Err(err).Msg("Invalid query parameters")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid query parameters",
		})
	}

	// Set defaults
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	// Build filters
	filters := &service.PendingBugsFilters{
		Release:   req.Release,
		Status:    req.Status,
		Severity:  req.Severity,
		Component: req.Component,
	}

	// If assigned_to_me is true (default), filter by current user
	if req.AssignedToMe {
		filters.AssignedTo = &userID
	}

	// Build pagination
	pagination := &repository.Pagination{
		Page:      req.Page,
		Limit:     req.Limit,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
	}

	// Get pending bugs
	result, err := h.releaseNoteService.GetPendingBugs(c.Context(), userID, filters, pagination)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get pending bugs")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: "Failed to retrieve pending bugs",
		})
	}

	// Convert to response
	// Ensure limit is at least 1 to avoid division by zero
	limit := req.Limit
	if limit < 1 {
		limit = 20 // Default limit
	}

	totalPages := int(result.Total) / limit
	if int(result.Total)%limit != 0 {
		totalPages++
	}

	response := &dto.PendingBugsResponse{
		Bugs:       make([]dto.BugResponse, 0, len(result.Bugs)),
		Total:      result.Total,
		Page:       req.Page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	for _, bug := range result.Bugs {
		if bugResp := dto.ToBugResponse(bug); bugResp != nil {
			response.Bugs = append(response.Bugs, *bugResp)
		}
	}

	return c.Status(fiber.StatusOK).JSON(dto.SuccessResponse{
		Success: true,
		Data:    response,
	})
}

// GetBugContext gets bug details with commit information
// GET /api/v1/release-notes/bug/:bug_id/context
func (h *ReleaseNoteHandler) GetBugContext(c *fiber.Ctx) error {
	bugIDStr := c.Params("bug_id")
	bugID, err := uuid.Parse(bugIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid bug ID",
		})
	}

	// Get bug context
	context, err := h.releaseNoteService.GetBugContext(c.Context(), bugID)
	if err != nil {
		logger.Error().Err(err).Str("bug_id", bugIDStr).Msg("Failed to get bug context")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: "Failed to retrieve bug context",
		})
	}

	// Convert to response
	response := &dto.BugContextResponse{
		Bug:              dto.ToBugResponse(context.Bug),
		Comments:         make([]dto.CommitInfoResponse, 0, len(context.Comments)),
		CommitCount:      context.CommitCount,
		ReadyForGenerate: context.CommitCount > 0,
	}

	for _, commit := range context.Comments {
		if commitResp := dto.ToCommitInfoResponse(commit); commitResp != nil {
			response.Comments = append(response.Comments, *commitResp)
		}
	}

	return c.Status(fiber.StatusOK).JSON(dto.SuccessResponse{
		Success: true,
		Data:    response,
	})
}

// GenerateReleaseNote generates a release note for a bug
// POST /api/v1/release-notes/generate
func (h *ReleaseNoteHandler) GenerateReleaseNote(c *fiber.Ctx) error {
	// Get current user from context
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}

	// Parse request body
	var req dto.GenerateReleaseNoteRequest
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

	// Generate release note
	note, err := h.releaseNoteService.GenerateReleaseNote(c.Context(), req.BugID, userID, req.ManualContent)
	if err != nil {
		logger.Error().Err(err).Str("bug_id", req.BugID.String()).Msg("Failed to generate release note")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "generation_failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(dto.SuccessResponse{
		Success: true,
		Data:    dto.ToReleaseNoteDetailResponse(note),
		Message: "Release note generated successfully",
	})
}

// GetReleaseNoteByBugID gets release note for a bug
// GET /api/v1/release-notes/bug/:bug_id
func (h *ReleaseNoteHandler) GetReleaseNoteByBugID(c *fiber.Ctx) error {
	bugIDStr := c.Params("bug_id")
	bugID, err := uuid.Parse(bugIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid bug ID",
		})
	}

	// Get release note
	note, err := h.releaseNoteService.GetReleaseNoteByBugID(c.Context(), bugID)
	if err != nil {
		logger.Error().Err(err).Str("bug_id", bugIDStr).Msg("Release note not found")
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "not_found",
			Message: "Release note not found for this bug",
		})
	}

	return c.Status(fiber.StatusOK).JSON(dto.SuccessResponse{
		Success: true,
		Data:    dto.ToReleaseNoteDetailResponse(note),
	})
}

// UpdateReleaseNote updates a release note
// PUT /api/v1/release-notes/:id
func (h *ReleaseNoteHandler) UpdateReleaseNote(c *fiber.Ctx) error {
	// Get current user from context
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}

	// Parse ID
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid release note ID",
		})
	}

	// Parse request body
	var req dto.UpdateReleaseNoteRequest
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

	// Update release note
	note, err := h.releaseNoteService.UpdateReleaseNote(c.Context(), id, req.Content, req.Status, userID)
	if err != nil {
		logger.Error().Err(err).Str("note_id", idStr).Msg("Failed to update release note")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(dto.SuccessResponse{
		Success: true,
		Data:    dto.ToReleaseNoteDetailResponse(note),
		Message: "Release note updated successfully",
	})
}

// BulkGenerateReleaseNotes generates release notes for multiple bugs
// POST /api/v1/release-notes/bulk-generate
func (h *ReleaseNoteHandler) BulkGenerateReleaseNotes(c *fiber.Ctx) error {
	// Get current user from context
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}

	// Parse request body
	var req dto.BulkGenerateRequest
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

	// Bulk generate
	result, err := h.releaseNoteService.BulkGenerateReleaseNotes(c.Context(), req.BugIDs, userID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to bulk generate release notes")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "bulk_generation_failed",
			Message: err.Error(),
		})
	}

	// Convert to response
	response := &dto.BulkGenerateResponse{
		Total:     result.Total,
		Generated: result.Generated,
		Failed:    result.Failed,
		Results:   make([]dto.BulkGenerateItemResponse, 0, len(result.Results)),
	}

	for _, item := range result.Results {
		response.Results = append(response.Results, dto.BulkGenerateItemResponse{
			BugID:         item.BugID,
			ReleaseNoteID: item.ReleaseNoteID,
			Status:        item.Status,
			Error:         item.Error,
		})
	}

	return c.Status(fiber.StatusOK).JSON(dto.SuccessResponse{
		Success: true,
		Data:    response,
		Message: "Generated release notes successfully",
	})
}

// ApproveReleaseNote approves or rejects a release note (manager only)
// POST /api/v1/release-notes/:id/approve
func (h *ReleaseNoteHandler) ApproveReleaseNote(c *fiber.Ctx) error {
	// Get current user from context
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}

	// Parse ID
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid release note ID",
		})
	}

	// Parse request body
	var req dto.ApproveReleaseNoteRequest
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

	// Approve or reject
	if req.Action == "approve" {
		err = h.releaseNoteService.ApproveReleaseNote(c.Context(), id, userID, req.Feedback)
	} else {
		feedbackStr := ""
		if req.Feedback != nil {
			feedbackStr = *req.Feedback
		}
		err = h.releaseNoteService.RejectReleaseNote(c.Context(), id, userID, feedbackStr)
	}

	if err != nil {
		logger.Error().Err(err).Str("note_id", idStr).Str("action", req.Action).Msg("Failed to process approval")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "approval_failed",
			Message: err.Error(),
		})
	}

	message := "Release note approved successfully"
	if req.Action == "reject" {
		message = "Release note rejected"
	}

	return c.Status(fiber.StatusOK).JSON(dto.SuccessResponse{
		Success: true,
		Message: message,
	})
}
