# Текущая задача: Phase 1 - Service Layer Refactoring

## Обзор

Первая фаза внедрения HTMX интерфейса в проект SimpleServiceDesk. Основная цель - создание сервисного слоя для устранения дублирования кода между API и будущими веб-хэндлерами.

## Цели Phase 1

1. **Создать сервисные интерфейсы** для бизнес-логики
2. **Рефакторинг существующих API хэндлеров** для использования сервисов
3. **Обновить HTTP сервер** для поддержки сервисного слоя
4. **Сохранить совместимость** существующих API контрактов

## Детальный план выполнения

### Шаг 1: Создание сервисных интерфейсов

**Создать директорию**: `internal/application/services/`

**Файлы для создания**:
- `internal/application/services/interfaces.go` - Общие интерфейсы и типы
- `internal/application/services/user_service.go` - Сервис пользователей
- `internal/application/services/ticket_service.go` - Сервис тикетов
- `internal/application/services/category_service.go` - Сервис категорий
- `internal/application/services/organization_service.go` - Сервис организаций

**Структура UserService (пример)**:
```go
type UserService interface {
    CreateUser(ctx context.Context, req CreateUserRequest) (*domain.User, error)
    GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error)
    UpdateUser(ctx context.Context, id uuid.UUID, req UpdateUserRequest) (*domain.User, error)
    DeleteUser(ctx context.Context, id uuid.UUID) error
    ListUsers(ctx context.Context, filter queries.UserFilter) ([]*domain.User, int64, error)
    UpdateUserRole(ctx context.Context, id uuid.UUID, role domain.Role) (*domain.User, error)
    GetUserTickets(ctx context.Context, userID uuid.UUID, filter queries.TicketFilter) ([]*domain.Ticket, int64, error)
}
```

### Шаг 2: Реализация сервисов

**Переместить бизнес-логику** из существующих хэндлеров в сервисы:

**Из `internal/application/users/create.go`**:
- Валидация пароля
- Хэширование пароля
- Создание пользователя через репозиторий
- Обработка ошибок домена

**Из `internal/application/users/update.go`**:
- Валидация обновлений
- Логика обновления полей
- Проверка прав доступа

**Из других хэндлеров**:
- Аналогично для tickets, categories, organizations

### Шаг 3: Рефакторинг существующих хэндлеров

**Создать новую структуру директорий**:
```
internal/application/handlers/
├── api/                    # Перемещенные API хэндлеры
│   ├── users/
│   ├── tickets/
│   ├── categories/
│   └── organizations/
```

**Обновить API хэндлеры**:
- Убрать бизнес-логику
- Оставить только HTTP-специфичную логику (binding, response formatting)
- Использовать сервисы для выполнения операций

**Пример рефакторинга** `PostUsers`:
```go
// Было: прямая работа с репозиторием + бизнес-логика
func (h UserHandlers) PostUsers(c echo.Context) error {
    // Валидация, хэширование пароля, создание пользователя
}

// Стало: делегирование сервису
func (h UserHandlers) PostUsers(c echo.Context) error {
    var req openapi.CreateUserRequest
    if err := c.Bind(&req); err != nil {
        return err
    }

    user, err := h.userService.CreateUser(ctx, CreateUserRequest{
        Name:     req.Name,
        Email:    string(req.Email),
        Password: req.Password,
    })

    // Обработка ошибок и формирование ответа
}
```

### Шаг 4: Обновление HTTP сервера

**Модификация `internal/application/http_server.go`**:
- Добавить создание сервисов
- Передавать сервисы в хэндлеры вместо репозиториев
- Подготовить структуру для будущих веб-хэндлеров

**Новая структура SetupHTTPServer**:
```go
func SetupHTTPServer(
    userRepo UserRepository,
    ticketRepo TicketRepository,
    organizationRepo OrganizationRepository,
    categoryRepo CategoryRepository,
) *echo.Echo {
    // Создание сервисов
    userService := services.NewUserService(userRepo)
    ticketService := services.NewTicketService(ticketRepo)
    // ...

    // Создание хэндлеров с сервисами
    apiHandlers := setupAPIHandlers(userService, ticketService, ...)

    // Регистрация маршрутов
    registerAPIRoutes(e, apiHandlers)

    return e
}
```

## Критерии завершения

### ✅ Сервисы созданы и работают
- [ ] Все 4 сервиса реализованы (Users, Tickets, Categories, Organizations)
- [ ] Бизнес-логика перенесена из хэндлеров в сервисы
- [ ] Сервисы используют репозитории для доступа к данным

### ✅ API хэндлеры рефакторены
- [ ] Хэндлеры перемещены в `internal/application/handlers/api/`
- [ ] Хэндлеры используют сервисы вместо прямого доступа к репозиториям
- [ ] Сохранена совместимость API контрактов

### ✅ HTTP сервер обновлен
- [ ] `http_server.go` использует сервисный слой
- [ ] Dependency injection настроен правильно
- [ ] Подготовлена структура для будущих веб-хэндлеров

### ✅ Тесты проходят
- [ ] `make test` проходит успешно
- [ ] `make test-integration` проходит успешно
- [ ] Существующие API тесты продолжают работать

### ✅ Код качество
- [ ] `make lint` проходит без ошибок
- [ ] Код соответствует стандартам проекта
- [ ] Комментарии и документация обновлены

## Задачи для выполнения

### Неделя 1: Создание сервисного слоя
1. **День 1-2**: Создать интерфейсы сервисов и базовые структуры
2. **День 3-4**: Реализовать UserService с переносом логики из хэндлеров
3. **День 5**: Создать тесты для UserService

### Неделя 2: Расширение сервисов
1. **День 1-2**: Реализовать TicketService
2. **День 3**: Реализовать CategoryService и OrganizationService
3. **День 4-5**: Рефакторинг всех API хэндлеров

### Неделя 3: Интеграция и тестирование
1. **День 1-2**: Обновить HTTP сервер и dependency injection
2. **День 3-4**: Обновить все тесты
3. **День 5**: Финальное тестирование и исправление багов

## Риски и митигация

**Риск**: Нарушение существующих API контрактов
**Митигация**: Тщательное тестирование интеграционными тестами

**Риск**: Усложнение кода без реальной пользы
**Митигация**: Четкое разделение ответственности, простые интерфейсы

**Риск**: Ошибки при переносе бизнес-логики
**Митигация**: Пошаговый рефакторинг, проверка тестами на каждом этапе

## Следующие шаги

После завершения Phase 1:
- **Phase 2**: Создание веб-инфраструктуры (шаблоны, статические файлы)
- **Phase 3**: Реализация основных веб-страниц
- **Phase 4**: Продвинутые HTMX функции

## Документация

Обновить документацию:
- README.md - добавить информацию о сервисном слое
- CLAUDE.md - обновить архитектурную секцию
- Добавить примеры использования сервисов в коде
