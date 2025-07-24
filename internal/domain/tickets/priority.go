package tickets

import (
	"errors"
	"strings"
)

var (
	ErrInvalidPriority = errors.New("invalid ticket priority")
)

// Priority представляет приоритет заявки
type Priority string

const (
	PriorityLow      Priority = "low"      // Низкий приоритет
	PriorityNormal   Priority = "normal"   // Обычный приоритет
	PriorityHigh     Priority = "high"     // Высокий приоритет
	PriorityCritical Priority = "critical" // Критический приоритет
)

// AllPriorities возвращает все возможные приоритеты
func AllPriorities() []Priority {
	return []Priority{
		PriorityLow,
		PriorityNormal,
		PriorityHigh,
		PriorityCritical,
	}
}

// String возвращает строковое представление приоритета
func (p Priority) String() string {
	return string(p)
}

// IsValid проверяет, является ли приоритет валидным
func (p Priority) IsValid() bool {
	for _, priority := range AllPriorities() {
		if p == priority {
			return true
		}
	}
	return false
}

// ParsePriority преобразует строку в приоритет
func ParsePriority(s string) (Priority, error) {
	priority := Priority(strings.ToLower(strings.TrimSpace(s)))
	if !priority.IsValid() {
		return "", ErrInvalidPriority
	}
	return priority, nil
}

// Weight возвращает числовой вес приоритета для сортировки
// Чем больше число, тем выше приоритет
func (p Priority) Weight() int {
	weights := map[Priority]int{
		PriorityLow:      1,
		PriorityNormal:   2,
		PriorityHigh:     3,
		PriorityCritical: 4,
	}

	if weight, exists := weights[p]; exists {
		return weight
	}
	return 0
}

// DisplayName возвращает человекочитаемое название приоритета
func (p Priority) DisplayName() string {
	names := map[Priority]string{
		PriorityLow:      "Низкий",
		PriorityNormal:   "Обычный",
		PriorityHigh:     "Высокий",
		PriorityCritical: "Критический",
	}

	if name, exists := names[p]; exists {
		return name
	}
	return string(p)
}

// Color возвращает цвет для отображения приоритета в UI
func (p Priority) Color() string {
	colors := map[Priority]string{
		PriorityLow:      "green",
		PriorityNormal:   "blue",
		PriorityHigh:     "orange",
		PriorityCritical: "red",
	}

	if color, exists := colors[p]; exists {
		return color
	}
	return "gray"
}

// SLA возвращает целевое время решения в часах для данного приоритета
func (p Priority) SLA() int {
	sla := map[Priority]int{
		PriorityLow:      72, // 3 дня
		PriorityNormal:   24, // 1 день
		PriorityHigh:     8,  // 8 часов
		PriorityCritical: 2,  // 2 часа
	}

	if hours, exists := sla[p]; exists {
		return hours
	}
	return 24 // по умолчанию 24 часа
}
