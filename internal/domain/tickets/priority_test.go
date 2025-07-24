package tickets_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	domain "simpleservicedesk/internal/domain/tickets"
)

func TestPriority_IsValid(t *testing.T) {
	tests := []struct {
		priority domain.Priority
		valid    bool
	}{
		{domain.PriorityLow, true},
		{domain.PriorityNormal, true},
		{domain.PriorityHigh, true},
		{domain.PriorityCritical, true},
		{domain.Priority("invalid"), false},
		{domain.Priority(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.priority), func(t *testing.T) {
			require.Equal(t, tt.valid, tt.priority.IsValid())
		})
	}
}

func TestPriority_Weight(t *testing.T) {
	tests := []struct {
		priority domain.Priority
		weight   int
	}{
		{domain.PriorityLow, 1},
		{domain.PriorityNormal, 2},
		{domain.PriorityHigh, 3},
		{domain.PriorityCritical, 4},
	}

	for _, tt := range tests {
		t.Run(string(tt.priority), func(t *testing.T) {
			require.Equal(t, tt.weight, tt.priority.Weight())
		})
	}
}

func TestPriority_SLA(t *testing.T) {
	tests := []struct {
		priority domain.Priority
		slaHours int
	}{
		{domain.PriorityLow, 72},     // 3 дня
		{domain.PriorityNormal, 24},  // 1 день
		{domain.PriorityHigh, 8},     // 8 часов
		{domain.PriorityCritical, 2}, // 2 часа
	}

	for _, tt := range tests {
		t.Run(string(tt.priority), func(t *testing.T) {
			require.Equal(t, tt.slaHours, tt.priority.SLA())
		})
	}
}

func TestPriority_Color(t *testing.T) {
	tests := []struct {
		priority domain.Priority
		color    string
	}{
		{domain.PriorityLow, "green"},
		{domain.PriorityNormal, "blue"},
		{domain.PriorityHigh, "orange"},
		{domain.PriorityCritical, "red"},
	}

	for _, tt := range tests {
		t.Run(string(tt.priority), func(t *testing.T) {
			require.Equal(t, tt.color, tt.priority.Color())
		})
	}
}

func TestParsePriority(t *testing.T) {
	tests := []struct {
		input    string
		expected domain.Priority
		hasError bool
	}{
		{"low", domain.PriorityLow, false},
		{"NORMAL", domain.PriorityNormal, false},
		{"  high  ", domain.PriorityHigh, false},
		{"critical", domain.PriorityCritical, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := domain.ParsePriority(tt.input)
			if tt.hasError {
				require.Error(t, err)
				require.Equal(t, domain.ErrInvalidPriority, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}
