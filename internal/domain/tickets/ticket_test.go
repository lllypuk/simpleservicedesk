package tickets_test

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "simpleservicedesk/internal/domain/tickets"
)

func TestNewTicket(t *testing.T) {
	organizationID := uuid.New()
	authorID := uuid.New()
	categoryID := uuid.New()

	tests := getNewTicketTestCases(organizationID, authorID, categoryID)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket, err := domain.NewTicket(
				uuid.New(),
				tt.title,
				tt.description,
				tt.priority,
				tt.organizationID,
				tt.authorID,
				tt.categoryID,
			)

			if tt.expectError {
				require.Error(t, err, "expected error but got none")
				return
			}

			require.NoError(t, err, "unexpected error creating ticket")
			validateTicketSuccess(t, ticket, tt.title, tt.priority, tt.organizationID, tt.authorID)
		})
	}
}

// validateTicketSuccess checks a successfully created ticket
func validateTicketSuccess(t *testing.T,
	ticket *domain.Ticket,
	expectedTitle string,
	expectedPriority domain.Priority,
	expectedOrgID, expectedAuthorID uuid.UUID) {
	assert.Equal(t, domain.StatusNew, ticket.Status())
	assert.Equal(t, expectedTitle, ticket.Title())
	assert.Equal(t, expectedPriority, ticket.Priority())
	assert.Equal(t, expectedOrgID, ticket.OrganizationID())
	assert.Equal(t, expectedAuthorID, ticket.AuthorID())
	assert.Empty(t, ticket.Comments())
	assert.Empty(t, ticket.Attachments())
}

// getNewTicketTestCases returns test cases for ticket creation
func getNewTicketTestCases(organizationID, authorID, categoryID uuid.UUID) []struct {
	name           string
	title          string
	description    string
	priority       domain.Priority
	organizationID uuid.UUID
	authorID       uuid.UUID
	categoryID     *uuid.UUID
	expectError    bool
} {
	return []struct {
		name           string
		title          string
		description    string
		priority       domain.Priority
		organizationID uuid.UUID
		authorID       uuid.UUID
		categoryID     *uuid.UUID
		expectError    bool
	}{
		{
			name:           "valid ticket",
			title:          "Test ticket",
			description:    "Test description",
			priority:       domain.PriorityNormal,
			organizationID: organizationID,
			authorID:       authorID,
			categoryID:     &categoryID,
			expectError:    false,
		},
		{
			name:           "empty title",
			title:          "",
			description:    "Test description",
			priority:       domain.PriorityNormal,
			organizationID: organizationID,
			authorID:       authorID,
			categoryID:     nil,
			expectError:    true,
		},
		{
			name:           "title too short",
			title:          "Hi",
			description:    "Test description",
			priority:       domain.PriorityNormal,
			organizationID: organizationID,
			authorID:       authorID,
			categoryID:     nil,
			expectError:    true,
		},
		{
			name:           "invalid priority",
			title:          "Test ticket",
			description:    "Test description",
			priority:       domain.Priority("invalid"),
			organizationID: organizationID,
			authorID:       authorID,
			categoryID:     nil,
			expectError:    true,
		},
		{
			name:           "empty organization ID",
			title:          "Test ticket",
			description:    "Test description",
			priority:       domain.PriorityNormal,
			organizationID: uuid.Nil,
			authorID:       authorID,
			categoryID:     nil,
			expectError:    true,
		},
		{
			name:           "empty author ID",
			title:          "Test ticket",
			description:    "Test description",
			priority:       domain.PriorityNormal,
			organizationID: organizationID,
			authorID:       uuid.Nil,
			categoryID:     nil,
			expectError:    true,
		},
	}
}

func TestTicket_ChangeStatus(t *testing.T) {
	ticket := createTestTicket(t)
	tests := getChangeStatusTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket.ResetToInitialStatus(tt.fromStatus)

			err := ticket.ChangeStatus(tt.toStatus)

			if tt.expectError {
				require.Error(t, err, "expected error but got none")
				return
			}

			require.NoError(t, err, "unexpected error changing status")
			validateStatusChange(t, ticket, tt.toStatus)
		})
	}
}

// validateStatusChange validates a successful status change
func validateStatusChange(t *testing.T, ticket *domain.Ticket, expectedStatus domain.Status) {
	assert.Equal(t, expectedStatus, ticket.Status())

	// Проверяем установку времени решения/закрытия
	if expectedStatus == domain.StatusResolved {
		assert.NotNil(t, ticket.ResolvedAt(), "expected resolvedAt to be set")
	}

	if expectedStatus == domain.StatusClosed {
		assert.NotNil(t, ticket.ClosedAt(), "expected closedAt to be set")
	}
}

// getChangeStatusTestCases returns test cases for status changes
func getChangeStatusTestCases() []struct {
	name        string
	fromStatus  domain.Status
	toStatus    domain.Status
	expectError bool
} {
	return []struct {
		name        string
		fromStatus  domain.Status
		toStatus    domain.Status
		expectError bool
	}{
		{
			name:        "new to in_progress",
			fromStatus:  domain.StatusNew,
			toStatus:    domain.StatusInProgress,
			expectError: false,
		},
		{
			name:        "new to resolved - should fail",
			fromStatus:  domain.StatusNew,
			toStatus:    domain.StatusResolved,
			expectError: true,
		},
		{
			name:        "in_progress to resolved",
			fromStatus:  domain.StatusInProgress,
			toStatus:    domain.StatusResolved,
			expectError: false,
		},
		{
			name:        "resolved to closed",
			fromStatus:  domain.StatusResolved,
			toStatus:    domain.StatusClosed,
			expectError: false,
		},
		{
			name:        "closed to in_progress (reopen)",
			fromStatus:  domain.StatusClosed,
			toStatus:    domain.StatusInProgress,
			expectError: false,
		},
		{
			name:        "invalid status",
			fromStatus:  domain.StatusNew,
			toStatus:    domain.Status("invalid"),
			expectError: true,
		},
	}
}

func TestTicket_UpdatePriority(t *testing.T) {
	ticket := createTestTicket(t)
	oldUpdatedAt := ticket.UpdatedAt()

	// Ждем немного, чтобы время обновления изменилось
	time.Sleep(time.Millisecond)

	err := ticket.UpdatePriority(domain.PriorityHigh)
	require.NoError(t, err)

	assert.Equal(t, domain.PriorityHigh, ticket.Priority())
	assert.True(t, ticket.UpdatedAt().After(oldUpdatedAt), "expected updatedAt to be updated")

	// Тест с невалидным приоритетом
	err = ticket.UpdatePriority(domain.Priority("invalid"))
	require.Error(t, err, "expected error for invalid priority")
}

func TestTicket_AssignTo(t *testing.T) {
	ticket := createTestTicket(t)
	assigneeID := uuid.New()

	err := ticket.AssignTo(assigneeID)
	require.NoError(t, err)

	assert.True(t, ticket.IsAssigned(), "expected ticket to be assigned")
	require.NotNil(t, ticket.AssigneeID())
	assert.Equal(t, assigneeID, *ticket.AssigneeID())

	// Тест с пустым ID
	err = ticket.AssignTo(uuid.Nil)
	require.Error(t, err, "expected error for empty assignee ID")
}

func TestTicket_Unassign(t *testing.T) {
	ticket := createTestTicket(t)
	assigneeID := uuid.New()

	// Сначала назначаем
	err := ticket.AssignTo(assigneeID)
	require.NoError(t, err)
	assert.True(t, ticket.IsAssigned(), "expected ticket to be assigned")

	// Затем снимаем назначение
	ticket.Unassign()
	assert.False(t, ticket.IsAssigned(), "expected ticket to be unassigned")
	assert.Nil(t, ticket.AssigneeID(), "expected assigneeID to be nil")
}

func TestTicket_AddComment(t *testing.T) {
	ticket := createTestTicket(t)
	authorID := uuid.New()
	content := "Test comment"

	err := ticket.AddComment(authorID, content, false)
	require.NoError(t, err)

	comments := ticket.Comments()
	require.Len(t, comments, 1, "expected 1 comment")

	comment := comments[0]
	assert.Equal(t, authorID, comment.AuthorID)
	assert.Equal(t, content, comment.Content)
	assert.False(t, comment.IsInternal, "expected isInternal to be false")

	// Тест с пустым комментарием
	err = ticket.AddComment(authorID, "", false)
	require.Error(t, err, "expected error for empty comment")

	// Тест с слишком длинным комментарием
	longContent := strings.Repeat("a", domain.MaxCommentLength+1)
	err = ticket.AddComment(authorID, longContent, false)
	require.Error(t, err, "expected error for too long comment")
}

func TestTicket_AddAttachment(t *testing.T) {
	ticket := createTestTicket(t)
	uploadedBy := uuid.New()

	err := ticket.AddAttachment("test.txt", 1024, "text/plain", "/path/to/file", uploadedBy)
	require.NoError(t, err)

	attachments := ticket.Attachments()
	require.Len(t, attachments, 1, "expected 1 attachment")

	attachment := attachments[0]
	assert.Equal(t, "test.txt", attachment.FileName)
	assert.Equal(t, int64(1024), attachment.FileSize)

	// Тест с пустым именем файла
	err = ticket.AddAttachment("", 1024, "text/plain", "/path/to/file", uploadedBy)
	require.Error(t, err, "expected error for empty file name")

	// Тест с нулевым размером файла
	err = ticket.AddAttachment("test.txt", 0, "text/plain", "/path/to/file", uploadedBy)
	require.Error(t, err, "expected error for zero file size")
}

func TestTicket_IsOverdue(t *testing.T) {
	// Создаем заявку с критическим приоритетом (SLA 2 часа)
	ticket := createTestTicketWithPriority(t, domain.PriorityCritical)

	// Устанавливаем время создания в прошлом (3 часа назад)
	ticket.SetCreatedAt(time.Now().Add(-3 * time.Hour))

	assert.True(t, ticket.IsOverdue(), "expected ticket to be overdue")

	// Решенная заявка не должна быть просроченной
	// Сначала переводим в "в работе", затем в "решена"
	err := ticket.ChangeStatus(domain.StatusInProgress)
	require.NoError(t, err)

	err = ticket.ChangeStatus(domain.StatusResolved)
	require.NoError(t, err)

	assert.False(t, ticket.IsOverdue(), "resolved ticket should not be overdue")
}

func TestTicket_GetPublicComments(t *testing.T) {
	ticket := createTestTicket(t)
	authorID := uuid.New()

	// Добавляем публичный комментарий
	err := ticket.AddComment(authorID, "Public comment", false)
	require.NoError(t, err)

	// Добавляем внутренний комментарий
	err = ticket.AddComment(authorID, "Internal comment", true)
	require.NoError(t, err)

	publicComments := ticket.GetPublicComments()
	require.Len(t, publicComments, 1, "expected 1 public comment")
	assert.Equal(t, "Public comment", publicComments[0].Content)
}

// Helper functions for tests
func createTestTicket(t *testing.T) *domain.Ticket {
	return createTestTicketWithPriority(t, domain.PriorityNormal)
}

func createTestTicketWithPriority(t *testing.T, priority domain.Priority) *domain.Ticket {
	ticket, err := domain.NewTicket(
		uuid.New(),
		"Test ticket",
		"Test description",
		priority,
		uuid.New(),
		uuid.New(),
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create test ticket: %v", err)
	}
	return ticket
}

// Comprehensive Workflow Transition Tests

func TestTicket_WorkflowTransitions_CompleteLifecycle(t *testing.T) {
	ticket := createTestTicket(t)

	// Initial state
	require.Equal(t, domain.StatusNew, ticket.Status())
	require.Nil(t, ticket.ResolvedAt())
	require.Nil(t, ticket.ClosedAt())

	// New -> In Progress
	originalUpdatedAt := ticket.UpdatedAt()
	time.Sleep(time.Millisecond)

	err := ticket.ChangeStatus(domain.StatusInProgress)
	require.NoError(t, err)
	require.Equal(t, domain.StatusInProgress, ticket.Status())
	require.True(t, ticket.UpdatedAt().After(originalUpdatedAt))
	require.Nil(t, ticket.ResolvedAt())
	require.Nil(t, ticket.ClosedAt())

	// In Progress -> Waiting
	err = ticket.ChangeStatus(domain.StatusWaiting)
	require.NoError(t, err)
	require.Equal(t, domain.StatusWaiting, ticket.Status())

	// Waiting -> In Progress (back to work)
	err = ticket.ChangeStatus(domain.StatusInProgress)
	require.NoError(t, err)
	require.Equal(t, domain.StatusInProgress, ticket.Status())

	// In Progress -> Resolved
	beforeResolve := time.Now()
	err = ticket.ChangeStatus(domain.StatusResolved)
	require.NoError(t, err)
	require.Equal(t, domain.StatusResolved, ticket.Status())
	require.NotNil(t, ticket.ResolvedAt())
	require.True(t, ticket.ResolvedAt().After(beforeResolve))
	require.Nil(t, ticket.ClosedAt())
	require.True(t, ticket.IsResolved())
	require.False(t, ticket.IsClosed())

	// Resolved -> Closed
	beforeClose := time.Now()
	time.Sleep(time.Millisecond)
	err = ticket.ChangeStatus(domain.StatusClosed)
	require.NoError(t, err)
	require.Equal(t, domain.StatusClosed, ticket.Status())
	require.NotNil(t, ticket.ResolvedAt()) // Should remain set
	require.NotNil(t, ticket.ClosedAt())
	require.True(t, ticket.ClosedAt().After(beforeClose))
	require.True(t, ticket.IsResolved())
	require.True(t, ticket.IsClosed())
}

func TestTicket_WorkflowTransitions_ReopenScenarios(t *testing.T) {
	ticket := createTestTicket(t)

	// Go through to resolved state
	require.NoError(t, ticket.ChangeStatus(domain.StatusInProgress))
	require.NoError(t, ticket.ChangeStatus(domain.StatusResolved))

	originalResolvedAt := ticket.ResolvedAt()
	require.NotNil(t, originalResolvedAt)

	// Reopen from resolved
	err := ticket.ChangeStatus(domain.StatusInProgress)
	require.NoError(t, err)
	require.Equal(t, domain.StatusInProgress, ticket.Status())
	require.Equal(t, originalResolvedAt, ticket.ResolvedAt()) // Should preserve original resolved time
	require.Nil(t, ticket.ClosedAt())

	// Go to closed
	require.NoError(t, ticket.ChangeStatus(domain.StatusResolved))
	require.NoError(t, ticket.ChangeStatus(domain.StatusClosed))

	// Reopen from closed
	err = ticket.ChangeStatus(domain.StatusInProgress)
	require.NoError(t, err)
	require.Equal(t, domain.StatusInProgress, ticket.Status())
	require.NotNil(t, ticket.ResolvedAt()) // Should still have resolved time
	require.NotNil(t, ticket.ClosedAt())   // Should still have closed time
}

func TestTicket_WorkflowTransitions_DirectClosureScenarios(t *testing.T) {
	tests := []struct {
		name        string
		fromStatus  domain.Status
		description string
	}{
		{"new_to_closed", domain.StatusNew, "Direct closure from new (e.g., spam)"},
		{"in_progress_to_closed", domain.StatusInProgress, "Direct closure from in progress"},
		{"waiting_to_closed", domain.StatusWaiting, "Direct closure from waiting"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket := createTestTicket(t)

			// Move to initial state if needed
			if tt.fromStatus != domain.StatusNew {
				require.NoError(t, ticket.ChangeStatus(domain.StatusInProgress))
				if tt.fromStatus == domain.StatusWaiting {
					require.NoError(t, ticket.ChangeStatus(domain.StatusWaiting))
				}
			}
			require.Equal(t, tt.fromStatus, ticket.Status())

			// Direct closure
			beforeClose := time.Now()
			err := ticket.ChangeStatus(domain.StatusClosed)
			require.NoError(t, err)
			require.Equal(t, domain.StatusClosed, ticket.Status())

			// Should set closed time but NOT resolved time for direct closure
			require.NotNil(t, ticket.ClosedAt())
			require.True(t, ticket.ClosedAt().After(beforeClose))

			if tt.fromStatus != domain.StatusResolved {
				require.Nil(t, ticket.ResolvedAt(), "Direct closure should not set resolved time")
			}
		})
	}
}

func TestTicket_WorkflowTransitions_InvalidTransitions(t *testing.T) {
	tests := []struct {
		name         string
		currentState domain.Status
		targetState  domain.Status
		setupFunc    func(*domain.Ticket) error
	}{
		{
			name:         "new_to_resolved_directly",
			currentState: domain.StatusNew,
			targetState:  domain.StatusResolved,
			setupFunc:    func(_ *domain.Ticket) error { return nil },
		},
		{
			name:         "closed_to_new",
			currentState: domain.StatusClosed,
			targetState:  domain.StatusNew,
			setupFunc: func(t *domain.Ticket) error {
				if err := t.ChangeStatus(domain.StatusInProgress); err != nil {
					return err
				}
				if err := t.ChangeStatus(domain.StatusResolved); err != nil {
					return err
				}
				return t.ChangeStatus(domain.StatusClosed)
			},
		},
		{
			name:         "closed_to_waiting",
			currentState: domain.StatusClosed,
			targetState:  domain.StatusWaiting,
			setupFunc: func(t *domain.Ticket) error {
				if err := t.ChangeStatus(domain.StatusInProgress); err != nil {
					return err
				}
				if err := t.ChangeStatus(domain.StatusResolved); err != nil {
					return err
				}
				return t.ChangeStatus(domain.StatusClosed)
			},
		},
		{
			name:         "closed_to_resolved",
			currentState: domain.StatusClosed,
			targetState:  domain.StatusResolved,
			setupFunc: func(t *domain.Ticket) error {
				if err := t.ChangeStatus(domain.StatusInProgress); err != nil {
					return err
				}
				if err := t.ChangeStatus(domain.StatusResolved); err != nil {
					return err
				}
				return t.ChangeStatus(domain.StatusClosed)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket := createTestTicket(t)

			// Setup the current state
			require.NoError(t, tt.setupFunc(ticket))
			require.Equal(t, tt.currentState, ticket.Status())

			// Attempt invalid transition
			err := ticket.ChangeStatus(tt.targetState)
			require.Error(t, err)
			require.ErrorIs(t, err, domain.ErrInvalidTransition)

			// Status should remain unchanged
			require.Equal(t, tt.currentState, ticket.Status())
		})
	}
}

// Business Rules Tests

func TestTicket_BusinessRules_SLACalculation(t *testing.T) {
	tests := []struct {
		name           string
		priority       domain.Priority
		expectedSLA    int
		hoursToOverdue int
	}{
		{"critical_priority", domain.PriorityCritical, 2, 3},
		{"high_priority", domain.PriorityHigh, 8, 9},
		{"normal_priority", domain.PriorityNormal, 24, 25},
		{"low_priority", domain.PriorityLow, 72, 73},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket := createTestTicketWithPriority(t, tt.priority)

			// Verify SLA calculation
			require.Equal(t, tt.expectedSLA, ticket.GetSLAHours())

			// Ticket should not be overdue initially
			require.False(t, ticket.IsOverdue())

			// Set created time to past to make it overdue
			ticket.SetCreatedAt(time.Now().Add(-time.Duration(tt.hoursToOverdue) * time.Hour))
			require.True(t, ticket.IsOverdue())

			// Resolved/closed tickets should not be overdue
			require.NoError(t, ticket.ChangeStatus(domain.StatusInProgress))
			require.NoError(t, ticket.ChangeStatus(domain.StatusResolved))
			require.False(t, ticket.IsOverdue(), "Resolved ticket should not be overdue")

			require.NoError(t, ticket.ChangeStatus(domain.StatusClosed))
			require.False(t, ticket.IsOverdue(), "Closed ticket should not be overdue")
		})
	}
}

func TestTicket_BusinessRules_AssignmentRules(t *testing.T) {
	ticket := createTestTicket(t)
	assigneeID := uuid.New()

	// Initially unassigned
	require.False(t, ticket.IsAssigned())
	require.Nil(t, ticket.AssigneeID())

	// Assign to someone
	require.NoError(t, ticket.AssignTo(assigneeID))
	require.True(t, ticket.IsAssigned())
	require.Equal(t, assigneeID, *ticket.AssigneeID())

	// Can reassign to different person
	newAssigneeID := uuid.New()
	require.NoError(t, ticket.AssignTo(newAssigneeID))
	require.Equal(t, newAssigneeID, *ticket.AssigneeID())

	// Unassign
	ticket.Unassign()
	require.False(t, ticket.IsAssigned())
	require.Nil(t, ticket.AssigneeID())

	// Cannot assign to nil UUID
	err := ticket.AssignTo(uuid.Nil)
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrTicketValidation)
}

func TestTicket_BusinessRules_CommentManagement(t *testing.T) {
	ticket := createTestTicket(t)
	authorID := uuid.New()

	// Add public comment
	require.NoError(t, ticket.AddComment(authorID, "Public comment", false))
	require.Len(t, ticket.Comments(), 1)
	require.Len(t, ticket.GetPublicComments(), 1)

	// Add internal comment
	require.NoError(t, ticket.AddComment(authorID, "Internal comment", true))
	require.Len(t, ticket.Comments(), 2)
	require.Len(t, ticket.GetPublicComments(), 1) // Still only 1 public

	// Add another public comment
	require.NoError(t, ticket.AddComment(authorID, "Another public comment", false))
	require.Len(t, ticket.Comments(), 3)
	require.Len(t, ticket.GetPublicComments(), 2) // Now 2 public

	// Verify comment content and metadata
	comments := ticket.Comments()
	require.Equal(t, "Public comment", comments[0].Content)
	require.False(t, comments[0].IsInternal)
	require.Equal(t, authorID, comments[0].AuthorID)
	require.Equal(t, ticket.ID(), comments[0].TicketID)

	require.Equal(t, "Internal comment", comments[1].Content)
	require.True(t, comments[1].IsInternal)

	// Comment validation
	require.Error(t, ticket.AddComment(authorID, "", false), "Empty comment should fail")
	require.Error(t, ticket.AddComment(authorID, "   ", false), "Whitespace-only comment should fail")
	require.Error(t, ticket.AddComment(uuid.Nil, "Valid comment", false), "Nil author should fail")

	// Very long comment should fail
	longComment := strings.Repeat("a", domain.MaxCommentLength+1)
	require.Error(t, ticket.AddComment(authorID, longComment, false), "Too long comment should fail")
}

func TestTicket_BusinessRules_AttachmentManagement(t *testing.T) {
	ticket := createTestTicket(t)
	uploaderID := uuid.New()

	// Add valid attachment
	require.NoError(t, ticket.AddAttachment(
		"document.pdf", 1024*1024, "application/pdf", "/uploads/document.pdf", uploaderID,
	))
	require.Len(t, ticket.Attachments(), 1)

	attachment := ticket.Attachments()[0]
	require.Equal(t, "document.pdf", attachment.FileName)
	require.Equal(t, int64(1024*1024), attachment.FileSize)
	require.Equal(t, "application/pdf", attachment.MimeType)
	require.Equal(t, "/uploads/document.pdf", attachment.FilePath)
	require.Equal(t, uploaderID, attachment.UploadedBy)
	require.Equal(t, ticket.ID(), attachment.TicketID)

	// Add multiple attachments
	require.NoError(t, ticket.AddAttachment("image.jpg", 512*1024, "image/jpeg", "/uploads/image.jpg", uploaderID))
	require.Len(t, ticket.Attachments(), 2)

	// Attachment validation
	require.Error(t, ticket.AddAttachment("", 1024, "text/plain", "/path", uploaderID), "Empty filename should fail")
	require.Error(t, ticket.AddAttachment("   ", 1024, "text/plain", "/path", uploaderID),
		"Whitespace filename should fail")
	require.Error(t, ticket.AddAttachment("file.txt", 0, "text/plain", "/path", uploaderID), "Zero size should fail")
	require.Error(t, ticket.AddAttachment("file.txt", -1, "text/plain", "/path", uploaderID),
		"Negative size should fail")
	require.Error(t, ticket.AddAttachment("file.txt", 1024, "text/plain", "/path", uuid.Nil),
		"Nil uploader should fail")
}

func TestTicket_BusinessRules_CategoryManagement(t *testing.T) {
	categoryID := uuid.New()
	ticket := createTestTicket(t)

	// Initially no category
	require.Nil(t, ticket.CategoryID())

	// Set category
	originalUpdatedAt := ticket.UpdatedAt()
	time.Sleep(time.Millisecond)
	ticket.SetCategory(&categoryID)
	require.Equal(t, categoryID, *ticket.CategoryID())
	require.True(t, ticket.UpdatedAt().After(originalUpdatedAt))

	// Change category
	newCategoryID := uuid.New()
	ticket.SetCategory(&newCategoryID)
	require.Equal(t, newCategoryID, *ticket.CategoryID())

	// Remove category
	ticket.SetCategory(nil)
	require.Nil(t, ticket.CategoryID())
}

func TestTicket_BusinessRules_StatusTrackingTimes(t *testing.T) {
	ticket := createTestTicket(t)

	// Track times for resolve/close events
	require.NoError(t, ticket.ChangeStatus(domain.StatusInProgress))

	// First resolve
	beforeFirstResolve := time.Now()
	require.NoError(t, ticket.ChangeStatus(domain.StatusResolved))
	firstResolvedAt := ticket.ResolvedAt()
	require.NotNil(t, firstResolvedAt)
	require.True(t, firstResolvedAt.After(beforeFirstResolve))

	// Reopen and resolve again - should update resolved time (new resolution)
	require.NoError(t, ticket.ChangeStatus(domain.StatusInProgress))
	time.Sleep(time.Millisecond)
	beforeSecondResolve := time.Now()
	require.NoError(t, ticket.ChangeStatus(domain.StatusResolved))
	secondResolvedAt := ticket.ResolvedAt()
	require.NotNil(t, secondResolvedAt)
	require.True(t, secondResolvedAt.After(beforeSecondResolve), "Second resolve should update resolved time")
	require.True(t, secondResolvedAt.After(*firstResolvedAt), "Second resolved time should be later than first")

	// Close
	beforeClose := time.Now()
	require.NoError(t, ticket.ChangeStatus(domain.StatusClosed))
	closedAt := ticket.ClosedAt()
	require.NotNil(t, closedAt)
	require.True(t, closedAt.After(beforeClose))

	// Reopen from closed and close again - should update closed time (new closure)
	require.NoError(t, ticket.ChangeStatus(domain.StatusInProgress))
	require.NoError(t, ticket.ChangeStatus(domain.StatusResolved))
	time.Sleep(time.Millisecond)
	beforeSecondClose := time.Now()
	require.NoError(t, ticket.ChangeStatus(domain.StatusClosed))
	secondClosedAt := ticket.ClosedAt()
	require.NotNil(t, secondClosedAt)
	require.True(t, secondClosedAt.After(beforeSecondClose), "Second close should update closed time")
	require.True(t, secondClosedAt.After(*closedAt), "Second closed time should be later than first")
}

// Ticket Relationships Tests

func TestTicket_Relationships_OrganizationIsolation(t *testing.T) {
	org1ID := uuid.New()
	org2ID := uuid.New()
	authorID := uuid.New()

	// Create tickets in different organizations
	ticket1, err := domain.NewTicket(
		uuid.New(),
		"Ticket 1",
		"Description",
		domain.PriorityNormal,
		org1ID,
		authorID,
		nil,
	)
	require.NoError(t, err)

	ticket2, err := domain.NewTicket(
		uuid.New(),
		"Ticket 2",
		"Description",
		domain.PriorityNormal,
		org2ID,
		authorID,
		nil,
	)
	require.NoError(t, err)

	// Verify organizational isolation
	require.Equal(t, org1ID, ticket1.OrganizationID())
	require.Equal(t, org2ID, ticket2.OrganizationID())
	require.NotEqual(t, ticket1.OrganizationID(), ticket2.OrganizationID())

	// Same author can create tickets in different organizations
	require.Equal(t, authorID, ticket1.AuthorID())
	require.Equal(t, authorID, ticket2.AuthorID())
}

func TestTicket_Relationships_CategoryLinking(t *testing.T) {
	orgID := uuid.New()
	authorID := uuid.New()
	categoryID := uuid.New()

	// Ticket with category
	ticketWithCategory, err := domain.NewTicket(
		uuid.New(),
		"Categorized Ticket",
		"Description",
		domain.PriorityNormal,
		orgID,
		authorID,
		&categoryID,
	)
	require.NoError(t, err)
	require.Equal(t, categoryID, *ticketWithCategory.CategoryID())

	// Ticket without category
	ticketWithoutCategory, err := domain.NewTicket(
		uuid.New(),
		"Uncategorized Ticket",
		"Description",
		domain.PriorityNormal,
		orgID,
		authorID,
		nil,
	)
	require.NoError(t, err)
	require.Nil(t, ticketWithoutCategory.CategoryID())

	// Move ticket between categories
	newCategoryID := uuid.New()
	ticketWithoutCategory.SetCategory(&newCategoryID)
	require.Equal(t, newCategoryID, *ticketWithoutCategory.CategoryID())

	// Remove category
	ticketWithCategory.SetCategory(nil)
	require.Nil(t, ticketWithCategory.CategoryID())
}

func TestTicket_Relationships_UserRoles(t *testing.T) {
	orgID := uuid.New()
	authorID := uuid.New()
	assigneeID := uuid.New()
	commenterID := uuid.New()

	ticket, err := domain.NewTicket(
		uuid.New(),
		"Multi-user Ticket",
		"Description",
		domain.PriorityNormal,
		orgID,
		authorID,
		nil,
	)
	require.NoError(t, err)

	// Verify author
	require.Equal(t, authorID, ticket.AuthorID())

	// Assign to different user
	require.NoError(t, ticket.AssignTo(assigneeID))
	require.Equal(t, assigneeID, *ticket.AssigneeID())
	require.NotEqual(t, ticket.AuthorID(), *ticket.AssigneeID())

	// Comments from different users
	require.NoError(t, ticket.AddComment(authorID, "Author comment", false))
	require.NoError(t, ticket.AddComment(assigneeID, "Assignee comment", true))
	require.NoError(t, ticket.AddComment(commenterID, "Third party comment", false))

	comments := ticket.Comments()
	require.Len(t, comments, 3)
	require.Equal(t, authorID, comments[0].AuthorID)
	require.Equal(t, assigneeID, comments[1].AuthorID)
	require.Equal(t, commenterID, comments[2].AuthorID)

	// Attachments from different users
	require.NoError(t, ticket.AddAttachment("author_file.txt", 1024, "text/plain", "/path1", authorID))
	require.NoError(t, ticket.AddAttachment("assignee_file.pdf", 2048, "application/pdf", "/path2", assigneeID))

	attachments := ticket.Attachments()
	require.Len(t, attachments, 2)
	require.Equal(t, authorID, attachments[0].UploadedBy)
	require.Equal(t, assigneeID, attachments[1].UploadedBy)
}
