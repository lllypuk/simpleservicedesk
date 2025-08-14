# Bug Fix Workflow - Процесс исправления багов

## 🎯 Когда использовать
- Обнаружена ошибка в существующей функциональности
- Пользователи сообщили о проблеме
- Automated tests выявили regression
- Code review обнаружил потенциальную проблему

## 📋 Пошаговый процесс

### 1. 🔍 Анализ и воспроизведение

#### Сбор информации
```bash
# Проверь логи для понимания проблемы
grep "ERROR" logs/app.log | tail -20

# Проверь текущее состояние
make test
make lint
```

#### Создание воспроизводимого примера
```go
// Создай тест, который воспроизводит баг
func TestBugRepro_DescriptionOfIssue(t *testing.T) {
    // Arrange - настройка условий, при которых возникает баг

    // Act - выполнение действия, вызывающего баг

    // Assert - проверка, что баг действительно происходит
    t.Skip("This test reproduces the bug - will be fixed")
}
```

### 2. 🏥 Диагностика root cause

#### Анализ кода
```bash
# Найди связанный код
rg "function_name" --type go
rg "error_message" --type go

# Проверь git history
git log --oneline -n 10 path/to/file.go
git blame path/to/file.go
```

#### Проверка архитектурных слоев
- **Domain Layer**: Проблема в бизнес-логике?
- **Application Layer**: Ошибка в координации между компонентами?
- **Infrastructure Layer**: Проблема с внешними зависимостями?
- **API Layer**: Неправильная обработка HTTP requests?

### 3. 🔧 Планирование исправления

#### Определи стратегию
- [ ] **Quick Fix**: Минимальное изменение для устранения симптома
- [ ] **Root Cause Fix**: Полное устранение первопричины
- [ ] **Refactoring**: Улучшение архитектуры для предотвращения подобных багов

#### Оценка воздействия
```go
// Определи затронутые компоненты
var affectedComponents = []string{
    "internal/domain/users",
    "internal/application/users",
    "api/users endpoints",
}

// Проверь backward compatibility
// Могут ли изменения сломать существующий API?
```

### 4. 🛠️ Реализация исправления

#### Создай ветку для исправления
```bash
# Naming convention: bugfix/issue-description
git checkout -b bugfix/user-validation-error
```

#### Test-Driven Development
```go
// 1. Сначала исправь тест, чтобы он показывал желаемое поведение
func TestUserValidation_ShouldWork(t *testing.T) {
    // Arrange
    user := &User{Name: "Valid Name", Email: "valid@email.com"}

    // Act
    err := validateUser(user)

    // Assert - теперь ожидаем success вместо ошибки
    assert.NoError(t, err) // Тест должен проходить после фикса
}

// 2. Убедись, что тест failing без исправления
go test -v ./internal/domain/users/...

// 3. Реализуй минимальное исправление
func validateUser(user *User) error {
    // Fix implementation here
    if strings.TrimSpace(user.Name) == "" {
        return ValidationError{Field: "name", Message: "name is required"}
    }
    // ... rest of validation
    return nil
}

// 4. Убедись, что тест теперь проходит
go test -v ./internal/domain/users/...
```

#### Применяй паттерны error handling
```go
// Используй существующие паттерны из patterns/error_handling.md
func (h *Handler) handleUserCreation(c echo.Context) error {
    user, err := h.createUser(req)
    if err != nil {
        // Логируй для debugging
        slog.Error("Failed to create user",
            "error", err,
            "email", req.Email,
            "request_id", getRequestID(c))

        // Возвращай structured error response
        return h.handleError(c, err)
    }

    return c.JSON(http.StatusCreated, user)
}
```

### 5. ✅ Проверка исправления

#### Запуск тестов
```bash
# Unit tests
make unit_test

# Integration tests
make integration_test

# Полный test suite
make test

# Проверь покрытие
make coverage_report
```

#### Проверка качества кода
```bash
# Линтинг и форматирование
make lint

# Проверь, что generated code актуален
make generate
git diff --exit-code # Должно быть пусто
```

#### Manual testing
```bash
# Запусти приложение локально
make run

# Проверь исправление через API
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Test User", "email": "test@example.com", "password": "password123"}'
```

### 6. 📝 Документирование

#### Обновление тестов
```go
// Добавь регрессионный тест
func TestUserCreation_RegressionTest_Issue123(t *testing.T) {
    // Этот тест предотвращает возврат бага
    // Описание проблемы: User validation failed for valid emails

    user := &User{
        Name:  "John Doe",
        Email: "john.doe+test@example.com", // Тест-кейс, который вызывал баг
    }

    err := validateUser(user)
    assert.NoError(t, err, "Should accept valid email with plus sign")
}
```

#### Git commit message
```bash
git add .
git commit -m "fix: resolve user validation error for emails with plus signs

- Fixed regex pattern in email validation
- Added test case to prevent regression
- Resolves issue where emails with '+' were rejected

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

### 7. 🔍 Code Review

#### Self review checklist
- [ ] Исправление addresses root cause, не только симптом
- [ ] Все тесты проходят
- [ ] Код следует project coding standards
- [ ] Error handling соответствует установленным паттернам
- [ ] Нет breaking changes в API
- [ ] Performance не пострадал

#### Подготовка для review
```bash
# Создай PR с описательным заголовком
git push origin bugfix/user-validation-error

# В описании PR включи:
# - Описание бага
# - Root cause analysis
# - Решение и его обоснование
# - Testing strategy
# - Затронутые компоненты
```

### 8. 🚀 Деплой и мониторинг

#### После merge в main
```bash
# Убедись, что CI/CD pipeline прошел успешно
# Мониторь логи после деплоя

# Проверь метрики
# - Error rates должны снизиться
# - Response times не должны ухудшиться
# - User complaints должны прекратиться
```

#### Follow-up actions
- [ ] Обнови документацию, если нужно
- [ ] Закрой связанные issues
- [ ] Уведоми заинтересованные стороны
- [ ] Добавь monitoring/alerting если проблема может повториться

## 🔧 Полезные команды

### Debugging commands
```bash
# Поиск похожих проблем в коде
rg "pattern" --type go -A 5 -B 5

# Анализ git истории
git log --grep="keyword" --oneline
git log -p path/to/file.go

# Профилирование производительности
make cpu_profile
make mem_profile
```

### Testing commands
```bash
# Запуск конкретного теста
go test -run TestSpecificFunction ./internal/domain/users/

# Запуск с verbose output
go test -v -run TestUserValidation ./internal/...

# Benchmark проблемной функции
go test -bench=BenchmarkUserValidation ./internal/domain/users/
```

## ⚠️ Распространенные ошибки

### Что НЕ делать
- ❌ Исправлять только симптом без анализа root cause
- ❌ Большие рефакторинги в bug fix'ах
- ❌ Изменения, которые могут сломать другую функциональность
- ❌ Отключение failing tests вместо исправления
- ❌ Коммиты без тестов

### Best practices
- ✅ Сначала создай репродуцирующий тест
- ✅ Минимальные изменения для исправления
- ✅ Comprehensive testing затронутой функциональности
- ✅ Clear commit messages с описанием проблемы и решения
- ✅ Documentation обновляется при необходимости

---

> 💡 **Помни**: Хороший bug fix не только устраняет проблему, но и предотвращает ее повторное возникновение через tests и improved error handling.
