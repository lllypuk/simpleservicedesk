package tickets

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTicketNotFound     = errors.New("ticket not found")
	ErrInvalidTicket      = errors.New("invalid ticket")
	ErrTicketValidation   = errors.New("ticket validation error")
	ErrUnauthorizedAccess = errors.New("unauthorized access to ticket")
	ErrInvalidTransition  = errors.New("invalid status transition")
)

const (
	MinTitleLength       = 3
	MaxTitleLength       = 200
	MaxDescriptionLength = 5000
	MaxCommentLength     = 2000
	formatError          = "%w: %s"
)

// Ticket представляет заявку в системе
type Ticket struct {
	id             uuid.UUID
	title          string
	description    string
	status         Status
	priority       Priority
	organizationID uuid.UUID
	categoryID     *uuid.UUID // Может быть nil, если категория не указана
	authorID       uuid.UUID  // ID создателя заявки
	assigneeID     *uuid.UUID // ID исполнителя, может быть nil
	comments       []Comment
	attachments    []Attachment
	createdAt      time.Time
	updatedAt      time.Time
	resolvedAt     *time.Time // Время решения заявки
	closedAt       *time.Time // Время закрытия заявки
}

// Comment представляет комментарий к заявке
type Comment struct {
	ID         uuid.UUID `json:"id"`
	TicketID   uuid.UUID `json:"ticket_id"`
	AuthorID   uuid.UUID `json:"author_id"`
	Content    string    `json:"content"`
	IsInternal bool      `json:"is_internal"` // Внутренний комментарий (не видим клиенту)
	CreatedAt  time.Time `json:"created_at"`
}

// Attachment представляет вложение к заявке
type Attachment struct {
	ID         uuid.UUID `json:"id"`
	TicketID   uuid.UUID `json:"ticket_id"`
	FileName   string    `json:"file_name"`
	FileSize   int64     `json:"file_size"`
	MimeType   string    `json:"mime_type"`
	FilePath   string    `json:"file_path"` // Путь к файлу в хранилище
	UploadedBy uuid.UUID `json:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at"`
}

// NewTicket создает новую заявку
func NewTicket(
	id uuid.UUID,
	title, description string,
	priority Priority,
	organizationID, authorID uuid.UUID,
	categoryID *uuid.UUID,
) (*Ticket, error) {
	title, err := validateTitle(title)
	if err != nil {
		return nil, err
	}

	description, err = validateDescription(description)
	if err != nil {
		return nil, err
	}

	if !priority.IsValid() {
		return nil, fmt.Errorf(formatError, ErrInvalidPriority, priority)
	}

	if err = validateUUID(organizationID, "organization_id"); err != nil {
		return nil, err
	}

	if err = validateUUID(authorID, "author_id"); err != nil {
		return nil, err
	}

	now := time.Now()
	ticket := &Ticket{
		id:             id,
		title:          title,
		description:    description,
		status:         StatusNew, // Новые заявки всегда имеют статус "new"
		priority:       priority,
		organizationID: organizationID,
		categoryID:     categoryID,
		authorID:       authorID,
		assigneeID:     nil,
		comments:       make([]Comment, 0),
		attachments:    make([]Attachment, 0),
		createdAt:      now,
		updatedAt:      now,
		resolvedAt:     nil,
		closedAt:       nil,
	}

	return ticket, nil
}

// CreateTicket создает новую заявку с автогенерацией ID
func CreateTicket(
	title, description string,
	priority Priority,
	organizationID, authorID uuid.UUID,
	categoryID *uuid.UUID,
) (*Ticket, error) {
	return NewTicket(
		uuid.New(),
		title,
		description,
		priority,
		organizationID,
		authorID,
		categoryID,
	)
}

func (t *Ticket) ID() uuid.UUID             { return t.id }
func (t *Ticket) Title() string             { return t.title }
func (t *Ticket) Description() string       { return t.description }
func (t *Ticket) Status() Status            { return t.status }
func (t *Ticket) Priority() Priority        { return t.priority }
func (t *Ticket) OrganizationID() uuid.UUID { return t.organizationID }
func (t *Ticket) CategoryID() *uuid.UUID    { return t.categoryID }
func (t *Ticket) AuthorID() uuid.UUID       { return t.authorID }
func (t *Ticket) AssigneeID() *uuid.UUID    { return t.assigneeID }
func (t *Ticket) Comments() []Comment       { return t.comments }
func (t *Ticket) Attachments() []Attachment { return t.attachments }
func (t *Ticket) CreatedAt() time.Time      { return t.createdAt }
func (t *Ticket) UpdatedAt() time.Time      { return t.updatedAt }
func (t *Ticket) ResolvedAt() *time.Time    { return t.resolvedAt }
func (t *Ticket) ClosedAt() *time.Time      { return t.closedAt }

func (t *Ticket) SetCreatedAt(newCreatedAt time.Time) {
	t.createdAt = newCreatedAt
}

// UpdateTitle обновляет заголовок заявки
func (t *Ticket) UpdateTitle(title string) error {
	validatedTitle, err := validateTitle(title)
	if err != nil {
		return err
	}
	t.title = validatedTitle
	t.updatedAt = time.Now()
	return nil
}

// UpdateDescription обновляет описание заявки
func (t *Ticket) UpdateDescription(description string) error {
	validatedDescription, err := validateDescription(description)
	if err != nil {
		return err
	}
	t.description = validatedDescription
	t.updatedAt = time.Now()
	return nil
}

// UpdatePriority обновляет приоритет заявки
func (t *Ticket) UpdatePriority(priority Priority) error {
	if !priority.IsValid() {
		return fmt.Errorf(formatError, ErrInvalidPriority, priority)
	}
	t.priority = priority
	t.updatedAt = time.Now()
	return nil
}

// ChangeStatus изменяет статус заявки с проверкой валидности перехода
func (t *Ticket) ChangeStatus(newStatus Status) error {
	if !newStatus.IsValid() {
		return fmt.Errorf(formatError, ErrInvalidStatus, newStatus)
	}

	if !t.status.CanTransitionTo(newStatus) {
		return fmt.Errorf("%w: cannot transition from %s to %s",
			ErrInvalidTransition, t.status, newStatus)
	}

	oldStatus := t.status
	t.status = newStatus
	t.updatedAt = time.Now()

	// Устанавливаем время решения/закрытия
	now := time.Now()
	if newStatus == StatusResolved && oldStatus != StatusResolved {
		t.resolvedAt = &now
	}
	if newStatus == StatusClosed && oldStatus != StatusClosed {
		t.closedAt = &now
	}

	return nil
}

// ResetToInitialStatus устанавливает начальный статус заявки
func (t *Ticket) ResetToInitialStatus(status Status) {
	t.status = status
	t.resolvedAt = nil
	t.closedAt = nil
	t.updatedAt = time.Now()
}

// AssignTo назначает заявку исполнителю
func (t *Ticket) AssignTo(assigneeID uuid.UUID) error {
	if err := validateUUID(assigneeID, "assignee_id"); err != nil {
		return err
	}
	t.assigneeID = &assigneeID
	t.updatedAt = time.Now()
	return nil
}

// Unassign снимает назначение с заявки
func (t *Ticket) Unassign() {
	t.assigneeID = nil
	t.updatedAt = time.Now()
}

// SetCategory устанавливает категорию заявки
func (t *Ticket) SetCategory(categoryID *uuid.UUID) {
	t.categoryID = categoryID
	t.updatedAt = time.Now()
}

// AddComment добавляет комментарий к заявке
func (t *Ticket) AddComment(authorID uuid.UUID, content string, isInternal bool) error {
	if err := validateUUID(authorID, "author_id"); err != nil {
		return err
	}

	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return fmt.Errorf("%w: comment content cannot be empty", ErrTicketValidation)
	}
	if len(content) > MaxCommentLength {
		return fmt.Errorf("%w: comment content too long (max %d characters)",
			ErrTicketValidation, MaxCommentLength)
	}

	comment := Comment{
		ID:         uuid.New(),
		TicketID:   t.id,
		AuthorID:   authorID,
		Content:    content,
		IsInternal: isInternal,
		CreatedAt:  time.Now(),
	}

	t.comments = append(t.comments, comment)
	t.updatedAt = time.Now()
	return nil
}

// AddAttachment добавляет вложение к заявке
func (t *Ticket) AddAttachment(fileName string, fileSize int64, mimeType, filePath string, uploadedBy uuid.UUID) error {
	if err := validateUUID(uploadedBy, "uploaded_by"); err != nil {
		return err
	}

	fileName = strings.TrimSpace(fileName)
	if len(fileName) == 0 {
		return fmt.Errorf("%w: file name cannot be empty", ErrTicketValidation)
	}

	if fileSize <= 0 {
		return fmt.Errorf("%w: file size must be positive", ErrTicketValidation)
	}

	attachment := Attachment{
		ID:         uuid.New(),
		TicketID:   t.id,
		FileName:   fileName,
		FileSize:   fileSize,
		MimeType:   mimeType,
		FilePath:   filePath,
		UploadedBy: uploadedBy,
		CreatedAt:  time.Now(),
	}

	t.attachments = append(t.attachments, attachment)
	t.updatedAt = time.Now()
	return nil
}

// IsAssigned проверяет, назначена ли заявка
func (t *Ticket) IsAssigned() bool {
	return t.assigneeID != nil
}

// IsResolved проверяет, решена ли заявка
func (t *Ticket) IsResolved() bool {
	return t.status == StatusResolved || t.status == StatusClosed
}

// IsClosed проверяет, закрыта ли заявка
func (t *Ticket) IsClosed() bool {
	return t.status == StatusClosed
}

// GetSLAHours возвращает количество часов SLA для данного приоритета
func (t *Ticket) GetSLAHours() int {
	return t.priority.GetSLAHours()
}

// IsOverdue проверяет, просрочена ли заявка
func (t *Ticket) IsOverdue() bool {
	if t.IsResolved() {
		return false
	}

	slaHours := t.GetSLAHours()
	deadline := t.createdAt.Add(time.Duration(slaHours) * time.Hour)
	return time.Now().After(deadline)
}

// GetPublicComments возвращает только публичные комментарии
func (t *Ticket) GetPublicComments() []Comment {
	var publicComments []Comment
	for _, comment := range t.comments {
		if !comment.IsInternal {
			publicComments = append(publicComments, comment)
		}
	}
	return publicComments
}

// Валидационные функции
func validateTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if len(title) < MinTitleLength {
		return "", fmt.Errorf("%w: title too short (min %d characters)",
			ErrTicketValidation, MinTitleLength)
	}
	if len(title) > MaxTitleLength {
		return "", fmt.Errorf("%w: title too long (max %d characters)",
			ErrTicketValidation, MaxTitleLength)
	}
	return title, nil
}

func validateDescription(description string) (string, error) {
	description = strings.TrimSpace(description)
	if len(description) > MaxDescriptionLength {
		return "", fmt.Errorf("%w: description too long (max %d characters)",
			ErrTicketValidation, MaxDescriptionLength)
	}
	return description, nil
}

func validateUUID(id uuid.UUID, fieldName string) error {
	if id == uuid.Nil {
		return fmt.Errorf("%w: %s cannot be empty", ErrTicketValidation, fieldName)
	}
	return nil
}
