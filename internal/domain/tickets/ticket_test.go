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
