package tickets_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	domain "simpleservicedesk/internal/domain/tickets"
)

func TestStatus_IsValid(t *testing.T) {
	tests := []struct {
		status domain.Status
		valid  bool
	}{
		{domain.StatusNew, true},
		{domain.StatusInProgress, true},
		{domain.StatusWaiting, true},
		{domain.StatusResolved, true},
		{domain.StatusClosed, true},
		{domain.Status("invalid"), false},
		{domain.Status(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			require.Equal(t, tt.valid, tt.status.IsValid())
		})
	}
}

func TestStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		from    domain.Status
		to      domain.Status
		canMove bool
	}{
		// Из New
		{domain.StatusNew, domain.StatusInProgress, true},
		{domain.StatusNew, domain.StatusWaiting, true},
		{domain.StatusNew, domain.StatusClosed, true},
		{domain.StatusNew, domain.StatusResolved, false},

		// Из InProgress
		{domain.StatusInProgress, domain.StatusWaiting, true},
		{domain.StatusInProgress, domain.StatusResolved, true},
		{domain.StatusInProgress, domain.StatusClosed, true},
		{domain.StatusInProgress, domain.StatusNew, false},

		// Из Resolved
		{domain.StatusResolved, domain.StatusClosed, true},
		{domain.StatusResolved, domain.StatusInProgress, true}, // Переоткрытие
		{domain.StatusResolved, domain.StatusNew, false},

		// Из Closed
		{domain.StatusClosed, domain.StatusInProgress, true}, // Переоткрытие
		{domain.StatusClosed, domain.StatusNew, false},
		{domain.StatusClosed, domain.StatusWaiting, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.from)+"_to_"+string(tt.to), func(t *testing.T) {
			require.Equal(t, tt.canMove, tt.from.CanTransitionTo(tt.to))
		})
	}
}

func TestStatus_IsOpenStatus(t *testing.T) {
	tests := []struct {
		status domain.Status
		isOpen bool
	}{
		{domain.StatusNew, true},
		{domain.StatusInProgress, true},
		{domain.StatusWaiting, true},
		{domain.StatusResolved, false},
		{domain.StatusClosed, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			require.Equal(t, tt.isOpen, tt.status.IsOpenStatus())
		})
	}
}

func TestParseStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected domain.Status
		hasError bool
	}{
		{"new", domain.StatusNew, false},
		{"IN_PROGRESS", domain.StatusInProgress, false},
		{"  waiting  ", domain.StatusWaiting, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := domain.ParseStatus(tt.input)
			if tt.hasError {
				require.Error(t, err)
				require.Equal(t, domain.ErrInvalidStatus, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}
