package mocks_test

import (
	"context"
	"testing"

	"simpleservicedesk/internal/application/mocks"
	"simpleservicedesk/internal/domain/users"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestUserRepositoryMock_BasicUsage(t *testing.T) {
	// Создаем мок с автоматическим поведением (по умолчанию)
	mockRepo := mocks.NewUserRepositoryMock()

	// Пример 1: Тест с автоматическим поведением (мок ведет себя как настоящий репозиторий)
	ctx := context.Background()
	email := "test@example.com"
	password := "password123"
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	// Создаем пользователя через мок
	user, err := mockRepo.CreateUser(ctx, email, passwordHash, func() (*users.User, error) {
		return users.CreateUser("Test User", email, passwordHash)
	})

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, email, user.Email())
	assert.Equal(t, "Test User", user.Name())

	// Получаем пользователя
	retrievedUser, err := mockRepo.GetUser(ctx, user.ID())
	require.NoError(t, err)
	assert.Equal(t, user.ID(), retrievedUser.ID())
	assert.Equal(t, user.Email(), retrievedUser.Email())
}

func TestUserRepositoryMock_WithExpectations(t *testing.T) {
	// Пример 2: Тест с явными ожиданиями
	mockRepo := mocks.NewUserRepositoryMock()
	mockRepo.EnableMockExpectations() // Включаем режим expectations
	defer mockRepo.AssertExpectations(t)

	ctx := context.Background()
	userID := uuid.New()
	expectedUser, _ := users.CreateUser("Expected User", "expected@example.com", []byte("hash"))

	// Настраиваем ожидания
	mockRepo.On("GetUser", ctx, userID).Return(expectedUser, nil).Once()

	// Вызываем метод
	result, err := mockRepo.GetUser(ctx, userID)

	// Проверяем результат
	require.NoError(t, err)
	assert.Equal(t, expectedUser, result)
}

func TestUserRepositoryMock_ErrorCases(t *testing.T) {
	// Пример 3: Тестирование ошибок с expectations
	mockRepo := mocks.NewUserRepositoryMock()
	mockRepo.EnableMockExpectations()
	defer mockRepo.AssertExpectations(t)

	ctx := context.Background()
	nonExistentID := uuid.New()

	// Настраиваем ожидание возврата ошибки
	mockRepo.On("GetUser", ctx, nonExistentID).Return((*users.User)(nil), users.ErrUserNotFound).Once()

	// Вызываем метод
	result, err := mockRepo.GetUser(ctx, nonExistentID)

	// Проверяем ошибку
	assert.Nil(t, result)
	assert.ErrorIs(t, err, users.ErrUserNotFound)
}

func TestUserRepositoryMock_UpdateUser(t *testing.T) {
	// Пример 4: Тестирование обновления пользователя с автоматическим поведением
	mockRepo := mocks.NewUserRepositoryMock()

	ctx := context.Background()
	email := "update@example.com"
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	// Сначала создаем пользователя
	user, err := mockRepo.CreateUser(ctx, email, passwordHash, func() (*users.User, error) {
		return users.CreateUser("Original Name", email, passwordHash)
	})
	require.NoError(t, err)

	// Обновляем пользователя
	updatedUser, err := mockRepo.UpdateUser(ctx, user.ID(), func(u *users.User) (bool, error) {
		err = u.ChangeEmail("updated@example.com")
		return err == nil, err
	})

	require.NoError(t, err)
	assert.Equal(t, "updated@example.com", updatedUser.Email())
}

func TestUserRepositoryMock_DuplicateEmail(t *testing.T) {
	// Пример 5: Тестирование дублирования email с автоматическим поведением
	mockRepo := mocks.NewUserRepositoryMock()

	ctx := context.Background()
	email := "duplicate@example.com"
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	// Создаем первого пользователя
	_, err := mockRepo.CreateUser(ctx, email, passwordHash, func() (*users.User, error) {
		return users.CreateUser("First User", email, passwordHash)
	})
	require.NoError(t, err)

	// Пытаемся создать второго пользователя с тем же email
	_, err = mockRepo.CreateUser(ctx, email, passwordHash, func() (*users.User, error) {
		return users.CreateUser("Second User", email, passwordHash)
	})

	// Должна быть ошибка дублирования
	assert.ErrorIs(t, err, users.ErrUserAlreadyExist)
}

func TestUserRepositoryMock_Reset(t *testing.T) {
	// Пример 6: Тестирование сброса мока
	mockRepo := mocks.NewUserRepositoryMock()

	ctx := context.Background()
	email := "reset@example.com"
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	// Создаем пользователя
	user, err := mockRepo.CreateUser(ctx, email, passwordHash, func() (*users.User, error) {
		return users.CreateUser("Test User", email, passwordHash)
	})
	require.NoError(t, err)

	// Проверяем, что пользователь существует
	_, err = mockRepo.GetUser(ctx, user.ID())
	require.NoError(t, err)

	// Сбрасываем мок
	mockRepo.Reset()

	// Проверяем, что пользователь больше не существует
	_, err = mockRepo.GetUser(ctx, user.ID())
	assert.ErrorIs(t, err, users.ErrUserNotFound)
}

func TestUserRepositoryMock_MixedExpectations(t *testing.T) {
	// Пример 7: Смешанный тест с expectations для одних методов и автоматическим поведением для других
	mockRepo := mocks.NewUserRepositoryMock()

	ctx := context.Background()
	email := "mixed@example.com"
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	// Сначала создаем пользователя с автоматическим поведением
	user, err := mockRepo.CreateUser(ctx, email, passwordHash, func() (*users.User, error) {
		return users.CreateUser("Mixed User", email, passwordHash)
	})
	require.NoError(t, err)

	// Теперь переключаемся на режим expectations для GetUser
	mockRepo.EnableMockExpectations()
	defer mockRepo.AssertExpectations(t)

	// Настраиваем ожидание для GetUser
	mockRepo.On("GetUser", ctx, user.ID()).Return(user, nil).Once()

	// Вызываем GetUser
	result, err := mockRepo.GetUser(ctx, user.ID())
	require.NoError(t, err)
	assert.Equal(t, user.ID(), result.ID())
}
