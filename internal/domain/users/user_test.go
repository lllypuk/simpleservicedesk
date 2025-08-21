package users_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	domain "simpleservicedesk/internal/domain/users"
)

func TestNewUser_Valid(t *testing.T) {
	id := uuid.New()
	name := "Alice"
	email := "alice@example.com"
	password := "password123"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	user, err := domain.NewUser(id, name, email, passwordHash)
	require.NoError(t, err)
	require.Equal(t, id, user.ID())
	require.Equal(t, name, user.Name())
	require.Equal(t, email, user.Email())
	require.True(t, user.CheckPassword(password))
}

func TestNewUser_InvalidName(t *testing.T) {
	id := uuid.New()
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	_, err := domain.NewUser(id, "", "test@example.com", passwordHash)
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrUserValidation)
}

func TestNewUser_InvalidEmail(t *testing.T) {
	id := uuid.New()
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	_, err := domain.NewUser(id, "Bob", "", passwordHash)
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrUserValidation)
}

func TestNewUser_InvalidPasswordHash_Empty(t *testing.T) {
	id := uuid.New()
	_, err := domain.NewUser(id, "Bob", "bob@example.com", []byte{})
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrUserValidation)
}

func TestNewUser_InvalidPasswordHash_Nil(t *testing.T) {
	id := uuid.New()
	_, err := domain.NewUser(id, "Bob", "bob@example.com", nil)
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrUserValidation)
}

func TestCreateUser_Valid(t *testing.T) {
	name := "Charlie"
	email := "charlie@example.com"
	password := "securepass"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	user, err := domain.CreateUser(name, email, passwordHash)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, user.ID())
	require.Equal(t, name, user.Name())
	require.Equal(t, email, user.Email())
	require.True(t, user.CheckPassword(password))
}

func TestUser_CheckPassword(t *testing.T) {
	password := "mypassword"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	user, err := domain.CreateUser("Dave", "dave@example.com", passwordHash)
	require.NoError(t, err)

	// Правильный пароль
	require.True(t, user.CheckPassword(password))

	// Неправильный пароль
	require.False(t, user.CheckPassword("wrongpassword"))
	require.False(t, user.CheckPassword(""))
}

func TestUser_ChangePassword(t *testing.T) {
	oldPassword := "oldpassword"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(oldPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	user, err := domain.CreateUser("Eve", "eve@example.com", passwordHash)
	require.NoError(t, err)

	// Проверяем, что старый пароль работает
	require.True(t, user.CheckPassword(oldPassword))

	// Меняем пароль
	newPassword := "newpassword123"
	err = user.ChangePassword(newPassword)
	require.NoError(t, err)

	// Проверяем, что новый пароль работает, а старый - нет
	require.True(t, user.CheckPassword(newPassword))
	require.False(t, user.CheckPassword(oldPassword))
}

func TestUser_ChangePassword_Invalid(t *testing.T) {
	password := "validpass"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	user, err := domain.CreateUser("Frank", "frank@example.com", passwordHash)
	require.NoError(t, err)

	// Попытка установить пустой пароль
	err = user.ChangePassword("")
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrUserValidation)

	// Попытка установить слишком короткий пароль
	err = user.ChangePassword("123")
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrUserValidation)

	// Проверяем, что старый пароль все еще работает
	require.True(t, user.CheckPassword(password))
}

func TestUser_ChangeEmail(t *testing.T) {
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user, err := domain.CreateUser("Grace", "grace@example.com", passwordHash)
	require.NoError(t, err)

	newEmail := "grace.new@example.com"
	err = user.ChangeEmail(newEmail)
	require.NoError(t, err)
	require.Equal(t, newEmail, user.Email())
}

func TestUser_ChangeEmail_Invalid(t *testing.T) {
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user, err := domain.CreateUser("Henry", "henry@example.com", passwordHash)
	require.NoError(t, err)

	err = user.ChangeEmail("")
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrUserValidation)

	// Проверяем, что email не изменился
	require.Equal(t, "henry@example.com", user.Email())
}

// Edge Cases Tests

func TestNewUser_EdgeCases_Username(t *testing.T) {
	id := uuid.New()
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name     string
		username string
		hasError bool
	}{
		{"empty string", "", true},
		{"whitespace only", "   ", false}, // current validation only checks for empty, not whitespace
		{"single space", " ", false},
		{"tab character", "\t", false},
		{"newline character", "\n", false},
		{"mixed whitespace", " \t\n ", false},
		{"unicode spaces", "\u00A0\u2000\u2028", false},
		{"single character", "a", false},
		{"very long name", string(make([]byte, 1000)), false},
		{"unicode characters", "José Μαρία 中文", false},
		{"special characters", "John-O'Connor_Jr.", false},
		{"numbers", "User123", false},
		{"emojis", "User😀👍", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewUser(id, tt.username, "test@example.com", passwordHash)
			if tt.hasError {
				require.Error(t, err)
				require.ErrorIs(t, err, domain.ErrUserValidation)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewUser_EdgeCases_Email(t *testing.T) {
	id := uuid.New()
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name     string
		email    string
		hasError bool
	}{
		{"empty string", "", true},
		{"whitespace only", "   ", false}, // current validation only checks for empty, not whitespace
		{"single space", " ", false},
		{"tab character", "\t", false},
		{"newline character", "\n", false},
		{"mixed whitespace", " \t\n ", false},
		{"unicode spaces", "\u00A0\u2000\u2028", false},
		{"simple email", "test@example.com", false},
		{"email with plus", "user+tag@example.com", false},
		{"email with subdomain", "user@mail.example.com", false},
		{"email with numbers", "user123@example123.com", false},
		{"email with hyphens", "user-name@ex-ample.com", false},
		{"email with dots", "user.name@example.co.uk", false},
		{"very long email", "user@" + string(make([]byte, 250)) + ".com", false},
		{"unicode in local part", "José@example.com", false},
		{"uppercase email", "USER@EXAMPLE.COM", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewUser(id, "TestUser", tt.email, passwordHash)
			if tt.hasError {
				require.Error(t, err)
				require.ErrorIs(t, err, domain.ErrUserValidation)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewUser_EdgeCases_PasswordHash(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name         string
		passwordHash []byte
		hasError     bool
	}{
		{"nil hash", nil, true},
		{"empty hash", []byte{}, true},
		{"single byte", []byte{0x01}, false},
		{"valid bcrypt hash", []byte("$2a$10$hash"), false},
		{"invalid bcrypt format", []byte("invalidhash"), false}, // validation accepts any non-empty hash
		{"very long hash", make([]byte, 1000), false},
		{"binary data", []byte{0x00, 0x01, 0xFF, 0xFE}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewUser(id, "TestUser", "test@example.com", tt.passwordHash)
			if tt.hasError {
				require.Error(t, err)
				require.ErrorIs(t, err, domain.ErrUserValidation)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewUser_EdgeCases_UUID(t *testing.T) {
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name string
		id   uuid.UUID
	}{
		{"nil UUID", uuid.Nil},
		{"random UUID", uuid.New()},
		{"specific UUID", uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")},
		{"max UUID", uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := domain.NewUser(tt.id, "TestUser", "test@example.com", passwordHash)
			require.NoError(t, err)
			require.Equal(t, tt.id, user.ID())
		})
	}
}

func TestUser_ChangePassword_EdgeCases(t *testing.T) {
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user, err := domain.CreateUser("TestUser", "test@example.com", passwordHash)
	require.NoError(t, err)

	tests := []struct {
		name        string
		newPassword string
		hasError    bool
	}{
		{"minimum length", "123456", false},
		{"just under minimum", "12345", true},
		{"empty password", "", true},
		{"whitespace only 6 chars", "      ", false}, // current validation only checks length and empty, not content
		{"tab characters 6 chars", "\t\t\t\t\t\t", false},
		{"newline characters 6 chars", "\n\n\n\n\n\n", false},
		{"mixed whitespace 6 chars", " \t\n \t\n", false},
		{"password with spaces", "pass word", false},
		{"unicode password", "пароль123", false},
		{"emoji password", "😀👍🔒🔑🛡️⚡", false},
		{"long password under 72 bytes", string(make([]rune, 70)), false}, // bcrypt limit is 72 bytes
		{"password at 72 bytes", string(make([]byte, 72)), false},         // bcrypt limit
		{"password over 72 bytes", string(make([]byte, 73)), true},        // should fail due to bcrypt limit
		{"special characters", "!@#$%^&*()_+-=[]{}|;':\",./<>?", false},
		{"binary-like data", string([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passwordErr := user.ChangePassword(tt.newPassword)
			if tt.hasError {
				require.Error(t, passwordErr)
				// Note: bcrypt limit errors are not wrapped as validation errors
				if tt.name == "password over 72 bytes" {
					require.Contains(t, passwordErr.Error(), "bcrypt")
				} else {
					require.ErrorIs(t, passwordErr, domain.ErrUserValidation)
				}
			} else {
				require.NoError(t, passwordErr)
				require.True(t, user.CheckPassword(tt.newPassword))
			}
		})
	}
}

func TestUser_ChangeEmail_EdgeCases(t *testing.T) {
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user, err := domain.CreateUser("TestUser", "original@example.com", passwordHash)
	require.NoError(t, err)

	tests := []struct {
		name     string
		newEmail string
		hasError bool
	}{
		{"same email", "original@example.com", false},
		{"empty string", "", true},
		{"whitespace only", "   ", false}, // current validation only checks for empty, not whitespace
		{"tab characters", "\t\t\t", false},
		{"newline characters", "\n\n\n", false},
		{"mixed whitespace", " \t\n ", false},
		{"unicode spaces", "\u00A0\u2000", false},
		{"email with leading space", " test@example.com", false}, // trimming not implemented
		{"email with trailing space", "test@example.com ", false},
		{"uppercase email", "TEST@EXAMPLE.COM", false},
		{"email with unicode", "José@example.com", false},
		{"very long email", "user@" + string(make([]byte, 200)) + ".com", false},
	}

	originalEmail := user.Email()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emailErr := user.ChangeEmail(tt.newEmail)
			if tt.hasError {
				require.Error(t, emailErr)
				require.ErrorIs(t, emailErr, domain.ErrUserValidation)
				require.Equal(t, originalEmail, user.Email()) // Email should not change on error
			} else {
				require.NoError(t, emailErr)
				require.Equal(t, tt.newEmail, user.Email())
				// Reset for next test
				_ = user.ChangeEmail(originalEmail)
			}
		})
	}
}

func TestUser_CheckPassword_EdgeCases(t *testing.T) {
	password := "testPassword123"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	user, err := domain.CreateUser("TestUser", "test@example.com", passwordHash)
	require.NoError(t, err)

	tests := []struct {
		name           string
		inputPassword  string
		expectedResult bool
	}{
		{"correct password", password, true},
		{"empty password", "", false},
		{"wrong password", "wrongPassword", false},
		{"case sensitive", "TESTPASSWORD123", false},
		{"partial password", "testPassword", false},
		{"password with extra chars", "testPassword123extra", false},
		{"unicode password", "пароль", false},
		{"password with null byte", "test\x00pass", false},
		{"very long wrong password", string(make([]byte, 1000)), false},
		{"whitespace variations", " testPassword123 ", false},
		{"tab and newline", "\ttestPassword123\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := user.CheckPassword(tt.inputPassword)
			require.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestUser_CheckPassword_WithDifferentBcryptCosts(t *testing.T) {
	password := "testPassword123"
	// Use reasonable costs for testing - MaxCost (31) takes ~10 minutes
	costs := []int{bcrypt.MinCost, bcrypt.DefaultCost, 14} // 14 is high but reasonable for tests

	for _, cost := range costs {
		t.Run(fmt.Sprintf("bcrypt_cost_%d", cost), func(t *testing.T) {
			passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
			require.NoError(t, err)

			user, err := domain.CreateUser("TestUser", "test@example.com", passwordHash)
			require.NoError(t, err)

			require.True(t, user.CheckPassword(password))
			require.False(t, user.CheckPassword("wrongPassword"))
		})
	}
}

func TestUser_SendToEmail_NotImplemented(t *testing.T) {
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user, err := domain.CreateUser("TestUser", "test@example.com", passwordHash)
	require.NoError(t, err)

	// Test that SendToEmail returns "not implemented" error for any input
	tests := []string{
		"",
		"test message",
		"very long message " + string(make([]byte, 1000)),
		"unicode message: 中文 русский العربية",
		"message with\nnewlines\tand\ttabs",
	}

	for _, message := range tests {
		t.Run(fmt.Sprintf("message_len_%d", len(message)), func(t *testing.T) {
			sendErr := user.SendToEmail(message)
			require.Error(t, sendErr)
			require.Contains(t, sendErr.Error(), "not implemented")
		})
	}
}

// Concurrent Operations Tests

func TestUser_ConcurrentPasswordChange(t *testing.T) {
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("initialpass"), bcrypt.DefaultCost)
	user, err := domain.CreateUser("ConcurrentUser", "test@example.com", passwordHash)
	require.NoError(t, err)

	const numGoroutines = 10
	const numIterations = 5

	var wg sync.WaitGroup
	passwords := make([]string, numGoroutines)

	// Generate unique passwords for each goroutine
	for i := range numGoroutines {
		passwords[i] = fmt.Sprintf("password%d", i)
	}

	// Start concurrent password changes
	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(index int) {
			defer wg.Done()
			for j := range numIterations {
				newPassword := fmt.Sprintf("%s_%d", passwords[index], j)
				passwordErr := user.ChangePassword(newPassword)
				// Password change should succeed
				if passwordErr != nil {
					t.Errorf("ChangePassword failed in goroutine %d iteration %d: %v", index, j, passwordErr)
				}
				// Brief delay to increase contention
				time.Sleep(time.Microsecond)
			}
		}(i)
	}

	wg.Wait()

	// After all concurrent operations, user should still be in a valid state
	// We can't predict which password will be the final one, but we can verify
	// the user object is not corrupted by checking that some valid password works
	finalPasswordChanged := false
	for i := range numGoroutines {
		for j := range numIterations {
			testPassword := fmt.Sprintf("%s_%d", passwords[i], j)
			if user.CheckPassword(testPassword) {
				finalPasswordChanged = true
				break
			}
		}
		if finalPasswordChanged {
			break
		}
	}
	require.True(t, finalPasswordChanged, "User should have one of the concurrently set passwords")
}

func TestUser_ConcurrentEmailChange(t *testing.T) {
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user, err := domain.CreateUser("ConcurrentUser", "original@example.com", passwordHash)
	require.NoError(t, err)

	const numGoroutines = 10
	const numIterations = 5

	var wg sync.WaitGroup
	emails := make([]string, numGoroutines)

	// Generate unique emails for each goroutine
	for i := range numGoroutines {
		emails[i] = fmt.Sprintf("user%d@example.com", i)
	}

	// Start concurrent email changes
	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(index int) {
			defer wg.Done()
			for j := range numIterations {
				newEmail := fmt.Sprintf("user%d_%d@example.com", index, j)
				emailErr := user.ChangeEmail(newEmail)
				// Email change should succeed
				if emailErr != nil {
					t.Errorf("ChangeEmail failed in goroutine %d iteration %d: %v", index, j, emailErr)
				}
				// Brief delay to increase contention
				time.Sleep(time.Microsecond)
			}
		}(i)
	}

	wg.Wait()

	// After all concurrent operations, verify user has a valid email
	finalEmail := user.Email()
	require.NotEmpty(t, finalEmail)
	require.NotEqual(t, "original@example.com", finalEmail) // Should have changed

	// Verify the email follows our expected pattern
	require.Contains(t, finalEmail, "@example.com")
	require.Contains(t, finalEmail, "user")
}

func TestUser_ConcurrentPasswordCheck(t *testing.T) {
	password := "testpassword123"
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user, err := domain.CreateUser("ConcurrentUser", "test@example.com", passwordHash)
	require.NoError(t, err)

	const numGoroutines = 20
	const numChecks = 100

	var wg sync.WaitGroup
	results := make([][]bool, numGoroutines)

	// Start concurrent password checks
	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(index int) {
			defer wg.Done()
			results[index] = make([]bool, numChecks)
			for j := range numChecks {
				// Mix of correct and incorrect password checks
				if j%2 == 0 {
					results[index][j] = user.CheckPassword(password)
				} else {
					results[index][j] = user.CheckPassword("wrongpassword")
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify all results are as expected
	for i := range numGoroutines {
		for j := range numChecks {
			if j%2 == 0 {
				require.True(t, results[i][j], "Correct password should always return true")
			} else {
				require.False(t, results[i][j], "Wrong password should always return false")
			}
		}
	}
}

// Helper functions for user concurrent operations testing
func runPasswordOperations(user *domain.User, index int, passwordChanges *int32, mu *sync.Mutex) {
	for j := range 10 {
		newPassword := fmt.Sprintf("pass%d_%d", index, j)
		passwordErr := user.ChangePassword(newPassword)
		if passwordErr == nil {
			mu.Lock()
			(*passwordChanges)++
			mu.Unlock()
		}
		time.Sleep(time.Microsecond)
	}
}

func runEmailOperations(user *domain.User, index int, emailChanges *int32, mu *sync.Mutex) {
	for j := range 10 {
		newEmail := fmt.Sprintf("user%d_%d@example.com", index, j)
		emailErr := user.ChangeEmail(newEmail)
		if emailErr == nil {
			mu.Lock()
			(*emailChanges)++
			mu.Unlock()
		}
		time.Sleep(time.Microsecond)
	}
}

func runPasswordCheckOperations(user *domain.User, index int, numGoroutines int,
	passwordChecks *int32, mu *sync.Mutex) {
	for j := range 20 {
		testPassword := fmt.Sprintf("pass%d_%d", (index-2+numGoroutines)%numGoroutines, j%10)
		user.CheckPassword(testPassword)
		mu.Lock()
		(*passwordChecks)++
		mu.Unlock()
		time.Sleep(time.Microsecond)
	}
}

func TestUser_ConcurrentMixedOperations(t *testing.T) {
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("initialpass"), bcrypt.DefaultCost)
	user, err := domain.CreateUser("MixedOpsUser", "initial@example.com", passwordHash)
	require.NoError(t, err)

	const numGoroutines = 15
	var wg sync.WaitGroup

	// Track operations for verification
	var passwordChanges, emailChanges, passwordChecks int32
	var mu sync.Mutex

	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(index int) {
			defer wg.Done()

			switch index % 3 {
			case 0:
				runPasswordOperations(user, index, &passwordChanges, &mu)
			case 1:
				runEmailOperations(user, index, &emailChanges, &mu)
			case 2:
				runPasswordCheckOperations(user, index, numGoroutines, &passwordChecks, &mu)
			}
		}(i)
	}

	wg.Wait()

	// Verify operations completed
	mu.Lock()
	require.Positive(t, passwordChanges, "Some password changes should have succeeded")
	require.Positive(t, emailChanges, "Some email changes should have succeeded")
	require.Positive(t, passwordChecks, "Password checks should have completed")
	mu.Unlock()

	// Verify user is still in a valid state
	require.NotEmpty(t, user.ID())
	require.NotEmpty(t, user.Name())
	require.NotEmpty(t, user.Email())
	require.NotEqual(t, "initial@example.com", user.Email()) // Should have changed
}

func TestCreateUser_Concurrent(t *testing.T) {
	const numGoroutines = 20
	var wg sync.WaitGroup

	users := make([]*domain.User, numGoroutines)
	testErrors := make([]error, numGoroutines)

	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(index int) {
			defer wg.Done()

			passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			user, err := domain.CreateUser(
				fmt.Sprintf("User%d", index),
				fmt.Sprintf("user%d@example.com", index),
				passwordHash,
			)
			users[index] = user
			testErrors[index] = err
		}(i)
	}

	wg.Wait()

	// Verify all users were created successfully
	uniqueIDs := make(map[uuid.UUID]bool)
	for i := range numGoroutines {
		require.NoError(t, testErrors[i])
		require.NotNil(t, users[i])

		// Verify unique IDs
		id := users[i].ID()
		require.False(t, uniqueIDs[id], "User IDs should be unique")
		uniqueIDs[id] = true

		// Verify correct data
		require.Equal(t, fmt.Sprintf("User%d", i), users[i].Name())
		require.Equal(t, fmt.Sprintf("user%d@example.com", i), users[i].Email())
		require.True(t, users[i].CheckPassword("password123"))
	}
}

func TestUser_ConcurrentAccessors(t *testing.T) {
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user, err := domain.CreateUser("AccessorUser", "accessor@example.com", passwordHash)
	require.NoError(t, err)

	const numGoroutines = 50
	const numReads = 100
	var wg sync.WaitGroup

	// Store results to verify consistency
	results := make([][]any, numGoroutines)

	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(index int) {
			defer wg.Done()
			results[index] = make([]any, numReads*3) // ID, Name, Email per iteration

			for j := range numReads {
				// Read all accessors concurrently
				results[index][j*3] = user.ID()
				results[index][j*3+1] = user.Name()
				results[index][j*3+2] = user.Email()
			}
		}(i)
	}

	wg.Wait()

	// Verify all reads returned consistent values
	expectedID := user.ID()
	expectedName := user.Name()
	expectedEmail := user.Email()

	for i := range numGoroutines {
		for j := range numReads {
			require.Equal(t, expectedID, results[i][j*3], "ID should be consistent across concurrent reads")
			require.Equal(t, expectedName, results[i][j*3+1], "Name should be consistent across concurrent reads")
			require.Equal(t, expectedEmail, results[i][j*3+2], "Email should be consistent across concurrent reads")
		}
	}
}

// Negative Tests - Testing error conditions and invalid inputs

func TestNewUser_NegativeTests(t *testing.T) {
	validID := uuid.New()
	validName := "TestUser"
	validEmail := "test@example.com"
	validPasswordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name         string
		id           uuid.UUID
		username     string
		email        string
		passwordHash []byte
		expectedErr  error
	}{
		{
			name:         "nil ID is allowed",
			id:           uuid.Nil,
			username:     validName,
			email:        validEmail,
			passwordHash: validPasswordHash,
			expectedErr:  nil, // uuid.Nil is actually valid for NewUser
		},
		{
			name:         "empty username",
			id:           validID,
			username:     "",
			email:        validEmail,
			passwordHash: validPasswordHash,
			expectedErr:  domain.ErrUserValidation,
		},
		{
			name:         "empty email",
			id:           validID,
			username:     validName,
			email:        "",
			passwordHash: validPasswordHash,
			expectedErr:  domain.ErrUserValidation,
		},
		{
			name:         "nil password hash",
			id:           validID,
			username:     validName,
			email:        validEmail,
			passwordHash: nil,
			expectedErr:  domain.ErrUserValidation,
		},
		{
			name:         "empty password hash",
			id:           validID,
			username:     validName,
			email:        validEmail,
			passwordHash: []byte{},
			expectedErr:  domain.ErrUserValidation,
		},
		{
			name:         "all invalid fields",
			id:           uuid.Nil, // This is actually valid
			username:     "",
			email:        "",
			passwordHash: nil,
			expectedErr:  domain.ErrUserValidation, // First validation error wins
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := domain.NewUser(tt.id, tt.username, tt.email, tt.passwordHash)
			if tt.expectedErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectedErr)
				require.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
			}
		})
	}
}

func TestCreateUser_NegativeTests(t *testing.T) {
	validName := "TestUser"
	validEmail := "test@example.com"
	validPasswordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name         string
		username     string
		email        string
		passwordHash []byte
		expectedErr  error
	}{
		{
			name:         "empty username",
			username:     "",
			email:        validEmail,
			passwordHash: validPasswordHash,
			expectedErr:  domain.ErrUserValidation,
		},
		{
			name:         "empty email",
			username:     validName,
			email:        "",
			passwordHash: validPasswordHash,
			expectedErr:  domain.ErrUserValidation,
		},
		{
			name:         "nil password hash",
			username:     validName,
			email:        validEmail,
			passwordHash: nil,
			expectedErr:  domain.ErrUserValidation,
		},
		{
			name:         "empty password hash slice",
			username:     validName,
			email:        validEmail,
			passwordHash: []byte{},
			expectedErr:  domain.ErrUserValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := domain.CreateUser(tt.username, tt.email, tt.passwordHash)
			require.Error(t, err)
			require.ErrorIs(t, err, tt.expectedErr)
			require.Nil(t, user)
		})
	}
}

func TestUser_ChangePassword_NegativeTests(t *testing.T) {
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("validpassword"), bcrypt.DefaultCost)
	user, err := domain.CreateUser("TestUser", "test@example.com", passwordHash)
	require.NoError(t, err)

	tests := []struct {
		name        string
		newPassword string
		expectedErr error
	}{
		{
			name:        "empty password",
			newPassword: "",
			expectedErr: domain.ErrUserValidation,
		},
		{
			name:        "password too short",
			newPassword: "12345", // 5 chars, minimum is 6
			expectedErr: domain.ErrUserValidation,
		},
		{
			name:        "password exactly minimum length minus 1",
			newPassword: string(make([]byte, domain.MinPasswordLength-1)),
			expectedErr: domain.ErrUserValidation,
		},
		{
			name:        "password over bcrypt limit",
			newPassword: string(make([]byte, 73)), // bcrypt limit is 72 bytes
			expectedErr: errors.New("dummy"),      // This will fail with bcrypt error, not validation error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalPasswordWorks := user.CheckPassword("validpassword")
			require.True(t, originalPasswordWorks, "Original password should work before test")

			err = user.ChangePassword(tt.newPassword)

			if tt.expectedErr != nil {
				require.Error(t, err)
				if tt.name == "password over bcrypt limit" {
					require.Contains(t, err.Error(), "bcrypt")
				} else {
					require.ErrorIs(t, err, tt.expectedErr)
				}
				// Password should not have changed on error
				require.True(t, user.CheckPassword("validpassword"),
					"Original password should still work after failed change")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUser_ChangeEmail_NegativeTests(t *testing.T) {
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user, err := domain.CreateUser("TestUser", "original@example.com", passwordHash)
	require.NoError(t, err)

	tests := []struct {
		name        string
		newEmail    string
		expectedErr error
	}{
		{
			name:        "empty email",
			newEmail:    "",
			expectedErr: domain.ErrUserValidation,
		},
		// Note: Current implementation is very basic and only checks for empty strings
		// More sophisticated email validation would catch these, but current code doesn't
		{
			name:        "invalid email format - no @",
			newEmail:    "invalidemail",
			expectedErr: nil, // Current validation doesn't check format
		},
		{
			name:        "invalid email format - multiple @",
			newEmail:    "user@@example.com",
			expectedErr: nil, // Current validation doesn't check format
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalEmail := user.Email()

			err = user.ChangeEmail(tt.newEmail)

			if tt.expectedErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectedErr)
				// Email should not have changed on error
				require.Equal(t, originalEmail, user.Email(), "Email should not change on validation error")
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.newEmail, user.Email(), "Email should have changed")
				// Reset for next test
				_ = user.ChangeEmail(originalEmail)
			}
		})
	}
}

func TestUser_CheckPassword_NegativeTests(t *testing.T) {
	password := "correctpassword"
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user, err := domain.CreateUser("TestUser", "test@example.com", passwordHash)
	require.NoError(t, err)

	// All these should return false (no errors thrown, just wrong passwords)
	negativeTests := []struct {
		name     string
		password string
	}{
		{"empty password", ""},
		{"wrong password", "wrongpassword"},
		{"password with different case", "CORRECTPASSWORD"},
		{"password with extra characters", "correctpassword123"},
		{"password substring", "correct"},
		{"password with leading space", " correctpassword"},
		{"password with trailing space", "correctpassword "},
		{"password with unicode", "correctpasswörd"},
		{"completely different", "totallydifferent"},
		{"numeric password", "123456789"},
		{"password with null byte", "correct\x00password"},
		{"very long wrong password", string(make([]byte, 1000))},
	}

	for _, tt := range negativeTests {
		t.Run(tt.name, func(t *testing.T) {
			result := user.CheckPassword(tt.password)
			require.False(t, result, "Wrong password should return false")
		})
	}
}

func TestUser_SendToEmail_AlwaysReturnsError(t *testing.T) {
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user, err := domain.CreateUser("TestUser", "test@example.com", passwordHash)
	require.NoError(t, err)

	// This method always returns "not implemented" error regardless of input
	tests := []string{
		"",
		"valid message",
		"message with\nnewlines",
		"unicode: 中文",
		"very long: " + string(make([]byte, 10000)),
		"null byte: \x00",
	}

	for _, message := range tests {
		t.Run(fmt.Sprintf("len_%d", len(message)), func(t *testing.T) {
			err = user.SendToEmail(message)
			require.Error(t, err)
			require.Contains(t, err.Error(), "not implemented")
		})
	}
}

func TestValidatePassword_NegativeTests(t *testing.T) {
	// Testing the internal validation function through ChangePassword
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("validpass"), bcrypt.DefaultCost)
	user, err := domain.CreateUser("TestUser", "test@example.com", passwordHash)
	require.NoError(t, err)

	invalidPasswords := []struct {
		name     string
		password string
		reason   string
	}{
		{"empty", "", "empty password"},
		{"single char", "a", "too short"},
		{"two chars", "ab", "too short"},
		{"three chars", "abc", "too short"},
		{"four chars", "abcd", "too short"},
		{"five chars", "abcde", "too short - exactly one under minimum"},
		{"whitespace only under minimum", "     ", "whitespace only under minimum length"},
	}

	for _, tt := range invalidPasswords {
		t.Run(tt.name, func(t *testing.T) {
			err = user.ChangePassword(tt.password)
			require.Error(t, err, "Password validation should fail for: %s", tt.reason)
			require.ErrorIs(t, err, domain.ErrUserValidation)

			// Verify original password still works
			require.True(t, user.CheckPassword("validpass"), "Original password should still work")
		})
	}
}

func TestUser_StateCorruption_Negative(t *testing.T) {
	// Test that user object maintains integrity even after failed operations
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("initialpass"), bcrypt.DefaultCost)
	user, err := domain.CreateUser("StateTestUser", "state@example.com", passwordHash)
	require.NoError(t, err)

	// Store initial state
	initialID := user.ID()
	initialName := user.Name()
	initialEmail := user.Email()

	// Try invalid operations that should fail
	err = user.ChangePassword("") // Empty password
	require.Error(t, err)

	err = user.ChangePassword("123") // Too short
	require.Error(t, err)

	err = user.ChangeEmail("") // Empty email
	require.Error(t, err)

	// Verify state hasn't been corrupted
	require.Equal(t, initialID, user.ID(), "ID should not change after failed operations")
	require.Equal(t, initialName, user.Name(), "Name should not change after failed operations")
	require.Equal(t, initialEmail, user.Email(), "Email should not change after failed operations")
	require.True(t, user.CheckPassword("initialpass"), "Original password should still work after failed operations")

	// Verify user is still functional with valid operations
	err = user.ChangePassword("newvalidpassword")
	require.NoError(t, err)
	require.True(t, user.CheckPassword("newvalidpassword"))
	require.False(t, user.CheckPassword("initialpass"))

	err = user.ChangeEmail("newemail@example.com")
	require.NoError(t, err)
	require.Equal(t, "newemail@example.com", user.Email())
}
