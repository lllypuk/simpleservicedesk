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

// Константы для весов приоритетов
const (
	WeightLow      = 1
	WeightNormal   = 2
	WeightHigh     = 3
	WeightCritical = 4
)

// Константы для SLA времен (в часах)
const (
	SLALowHours      = 72 // 3 дня
	SLANormalHours   = 24 // 1 день
	SLAHighHours     = 8  // 8 часов
	SLACriticalHours = 2  // 2 часа
	SLADefaultHours  = 24 // по умолчанию 24 часа
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
		PriorityLow:      WeightLow,
		PriorityNormal:   WeightNormal,
		PriorityHigh:     WeightHigh,
		PriorityCritical: WeightCritical,
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
		PriorityLow:      SLALowHours,
		PriorityNormal:   SLANormalHours,
		PriorityHigh:     SLAHighHours,
		PriorityCritical: SLACriticalHours,
	}

	if hours, exists := sla[p]; exists {
		return hours
	}
	return SLADefaultHours
}

// GetSLAHours возвращает количество часов SLA для приоритета
func (p Priority) GetSLAHours() int {
	switch p {
	case PriorityLow:
		return SLALowHours
	case PriorityNormal:
		return SLANormalHours
	case PriorityHigh:
		return SLAHighHours
	case PriorityCritical:
		return SLACriticalHours
	default:
		return SLADefaultHours
	}
}

// GetWeight возвращает вес приоритета для сортировки
func (p Priority) GetWeight() int {
	switch p {
	case PriorityLow:
		return WeightLow
	case PriorityNormal:
		return WeightNormal
	case PriorityHigh:
		return WeightHigh
	case PriorityCritical:
		return WeightCritical
	default:
		return WeightNormal
	}
}
