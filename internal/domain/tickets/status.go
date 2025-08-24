package tickets

import (
	"errors"
	"strings"
)

var (
	ErrInvalidStatus = errors.New("invalid ticket status")
)

// Status представляет статус заявки
type Status string

const (
	StatusNew        Status = "new"         // Новая заявка
	StatusInProgress Status = "in_progress" // В работе
	StatusWaiting    Status = "waiting"     // Ожидание (клиента/информации)
	StatusResolved   Status = "resolved"    // Решена
	StatusClosed     Status = "closed"      // Закрыта
)

// AllStatuses возвращает все возможные статусы
func AllStatuses() []Status {
	return []Status{
		StatusNew,
		StatusInProgress,
		StatusWaiting,
		StatusResolved,
		StatusClosed,
	}
}

// String возвращает строковое представление статуса
func (s Status) String() string {
	return string(s)
}

// IsValid проверяет, является ли статус валидным
func (s Status) IsValid() bool {
	for _, status := range AllStatuses() {
		if s == status {
			return true
		}
	}
	return false
}

// CanTransitionTo проверяет, возможен ли переход из текущего статуса в новый
func (s Status) CanTransitionTo(newStatus Status) bool {
	if !s.IsValid() || !newStatus.IsValid() {
		return false
	}

	// Определяем допустимые переходы статусов согласно бизнес-логике
	allowedTransitions := map[Status][]Status{
		StatusNew: {
			StatusInProgress,
			StatusWaiting,
			StatusClosed,
		},
		StatusInProgress: {
			StatusWaiting,
			StatusResolved,
			StatusClosed,
		},
		StatusWaiting: {
			StatusInProgress,
			StatusResolved,
			StatusClosed,
		},
		StatusResolved: {
			StatusClosed,
			StatusInProgress, // Переоткрытие заявки
		},
		StatusClosed: {
			StatusInProgress, // Переоткрытие закрытой заявки
		},
	}

	allowedStatuses, exists := allowedTransitions[s]
	if !exists {
		return false
	}

	for _, allowed := range allowedStatuses {
		if allowed == newStatus {
			return true
		}
	}

	return false
}

// ParseStatus преобразует строку в статус
func ParseStatus(s string) (Status, error) {
	status := Status(strings.ToLower(strings.TrimSpace(s)))
	if !status.IsValid() {
		return "", ErrInvalidStatus
	}
	return status, nil
}

// IsOpenStatus проверяет, является ли статус "открытым" (заявка в работе)
func (s Status) IsOpenStatus() bool {
	return s == StatusNew || s == StatusInProgress || s == StatusWaiting
}

// IsClosedStatus проверяет, является ли статус "закрытым"
func (s Status) IsClosedStatus() bool {
	return s == StatusResolved || s == StatusClosed
}

// DisplayName возвращает человекочитаемое название статуса
func (s Status) DisplayName() string {
	names := map[Status]string{
		StatusNew:        "Новая",
		StatusInProgress: "В работе",
		StatusWaiting:    "Ожидание",
		StatusResolved:   "Решена",
		StatusClosed:     "Закрыта",
	}

	if name, exists := names[s]; exists {
		return name
	}
	return string(s)
}

// Color возвращает цвет для отображения статуса в UI
func (s Status) Color() string {
	colors := map[Status]string{
		StatusNew:        "blue",
		StatusInProgress: "orange",
		StatusWaiting:    "yellow",
		StatusResolved:   "green",
		StatusClosed:     "gray",
	}

	if color, exists := colors[s]; exists {
		return color
	}
	return "gray"
}

// IsTerminal проверяет, является ли статус финальным
func (s Status) IsTerminal() bool {
	return s == StatusClosed
}

// IsActive проверяет, является ли статус активным (требует работы)
func (s Status) IsActive() bool {
	return s == StatusNew || s == StatusInProgress || s == StatusWaiting
}
