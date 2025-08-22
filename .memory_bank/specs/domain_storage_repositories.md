# Domain Storage Repositories - Реализация хранения для всех доменных сущностей

## 🎯 Описание фичи

### Проблема

В настоящее время в системе реализовано только хранение пользователей (Users), но отсутствуют репозитории и слои
хранения для остальных ключевых доменных сущностей:

- **Tickets**: Заявки в службу поддержки - основная сущность системы
- **Categories**: Система категоризации заявок для организации и фильтрации
- **Organizations**: Организационная структура для группировки пользователей и заявок

Без этих репозиториев невозможно:

- Создавать и управлять заявками
- Организовать иерархическую структуру категорий
- Группировать пользователей по организациям
- Полноценно тестировать business logic

### Решение

Реализация полного набора репозиториев следуя существующему паттерну:

- **Единообразная архитектура**: Применение тех же принципов, что и в UserRepository
- **MongoDB**: Основная реализация на MongoDB
- **Транзакционность**: Поддержка ACID операций где необходимо
- **Тестируемость**: Полное покрытие unit и integration тестами

## 📋 Требования

### Функциональные требования

- [ ] **FR-1**: TicketRepository с CRUD операциями и поиском по статусу/приоритету
- [ ] **FR-2**: CategoryRepository с поддержкой иерархической структуры (parent-child)
- [ ] **FR-3**: OrganizationRepository с иерархической структурой организаций
- [ ] **FR-4**: Связи между сущностями (Ticket-User, Ticket-Category, User-Organization)
- [ ] **FR-5**: Поддержка фильтрации и сортировки для всех репозиториев
- [ ] **FR-6**: Batch операции для эффективной работы с множественными записями

### Нефункциональные требования

- [ ] **NFR-1**: Производительность (время отклика < 200ms для простых запросов)
- [ ] **NFR-2**: Масштабируемость (поддержка 10K+ записей в каждой коллекции)
- [ ] **NFR-3**: Безопасность (валидация данных на уровне репозиториев)
- [ ] **NFR-4**: Надежность (graceful handling MongoDB connection issues)
- [ ] **NFR-5**: Совместимость (единый интерфейс для MongoDB реализации)

## 🏗️ Техническая архитектура

### Затронутые компоненты

- **Application Layer**: `internal/application/interfaces.go` - новые интерфейсы репозиториев
- **Infrastructure Layer**:
    - `internal/infrastructure/tickets/` - MongoDB реализация
    - `internal/infrastructure/categories/` - MongoDB реализация
    - `internal/infrastructure/organizations/` - MongoDB реализация
- **Testing**: Моки для всех новых репозиториев в `internal/application/mocks/`

### Новые интерфейсы репозиториев

```go
// TicketRepository - управление заявками
type TicketRepository interface {
CreateTicket(ctx context.Context, createFn func () (*tickets.Ticket, error)) (*tickets.Ticket, error)
UpdateTicket(ctx context.Context, id uuid.UUID, updateFn func (*tickets.Ticket) (bool, error)) (*tickets.Ticket, error)
GetTicket(ctx context.Context, id uuid.UUID) (*tickets.Ticket, error)
ListTickets(ctx context.Context, filter TicketFilter) ([]*tickets.Ticket, error)
DeleteTicket(ctx context.Context, id uuid.UUID) error
}

// CategoryRepository - управление категориями
type CategoryRepository interface {
CreateCategory(ctx context.Context, createFn func () (*categories.Category, error)) (*categories.Category, error)
UpdateCategory(ctx context.Context, id uuid.UUID, updateFn func (*categories.Category) (bool, error)) (*categories.Category, error)
GetCategory(ctx context.Context, id uuid.UUID) (*categories.Category, error)
ListCategories(ctx context.Context, filter CategoryFilter) ([]*categories.Category, error)
GetCategoryHierarchy(ctx context.Context, rootID uuid.UUID) (*categories.CategoryTree, error)
DeleteCategory(ctx context.Context, id uuid.UUID) error
}

// OrganizationRepository - управление организациями
type OrganizationRepository interface {
CreateOrganization(ctx context.Context, createFn func () (*organizations.Organization, error)) (*organizations.Organization, error)
UpdateOrganization(ctx context.Context, id uuid.UUID, updateFn func (*organizations.Organization) (bool, error)) (*organizations.Organization, error)
GetOrganization(ctx context.Context, id uuid.UUID) (*organizations.Organization, error)
ListOrganizations(ctx context.Context, filter OrganizationFilter) ([]*organizations.Organization, error)
GetOrganizationHierarchy(ctx context.Context, rootID uuid.UUID) (*organizations.OrganizationTree, error)
DeleteOrganization(ctx context.Context, id uuid.UUID) error
}
```

### MongoDB коллекции

- **tickets**: Основная коллекция заявок
- **categories**: Категории с parent_id для иерархии
- **organizations**: Организации с parent_id для иерархии

### Новые индексы

```javascript
// tickets collection
db.tickets.createIndex({"status": 1})
db.tickets.createIndex({"priority": 1})
db.tickets.createIndex({"assigned_to": 1})
db.tickets.createIndex({"category_id": 1})
db.tickets.createIndex({"created_at": -1})

// categories collection
db.categories.createIndex({"parent_id": 1})
db.categories.createIndex({"name": 1}, {unique: true})

// organizations collection  
db.organizations.createIndex({"parent_id": 1})
db.organizations.createIndex({"name": 1}, {unique: true})
```

## 🔄 User Stories

### История 1: Agent хочет управлять заявками

**Как** агент службы поддержки  
**Я хочу** создавать, обновлять и искать заявки  
**Чтобы** эффективно обрабатывать запросы пользователей

**Критерии приемки:**

- [ ] Given валидные данные заявки, when создаю заявку, then заявка сохраняется с уникальным ID
- [ ] Given существующая заявка, when обновляю статус, then статус изменяется согласно business rules
- [ ] Given критерии поиска, when ищу заявки, then получаю отфильтрованный список

### История 2: Admin хочет управлять категориями

**Как** администратор системы  
**Я хочу** создавать иерархию категорий  
**Чтобы** организовать классификацию заявок

**Критерии приемки:**

- [ ] Given родительскую категорию, when создаю подкатегорию, then устанавливается parent-child связь
- [ ] Given категорию с подкategorиями, when получаю иерархию, then возвращается полное дерево
- [ ] Given категорию с заявками, when пытаюсь удалить, then операция блокируется

### История 3: Admin хочет управлять организационной структурой

**Как** администратор системы  
**Я хочу** создавать иерархию организаций  
**Чтобы** группировать пользователей и контролировать доступ

**Критерии приемки:**

- [ ] Given организацию, when создаю подразделение, then устанавливается иерархическая связь
- [ ] Given организацию, when получаю всех пользователей, then возвращаются пользователи всех подразделений
- [ ] Given организацию с пользователями, when пытаюсь удалить, then операция блокируется

## 🧪 План тестирования

### Unit Tests

- [ ] **Domain logic**: Валидация бизнес-правил для каждой сущности
- [ ] **Repository methods**: Тестирование CRUD операций с моками
- [ ] **Filter logic**: Проверка корректности фильтрации и сортировки
- [ ] **Error handling**: Обработка всех возможных ошибочных ситуаций

### Integration Tests

- [ ] **MongoDB repositories**: Тестирование с реальной БД через testcontainers
- [ ] **Cross-repository operations**: Операции затрагивающие несколько репозиториев
- [ ] **Transaction handling**: Проверка ACID свойств для сложных операций
- [ ] **Performance**: Load testing для критичных операций

### Test Coverage Goals

- **Unit tests**: > 90% для репозиториев
- **Integration tests**: 100% покрытие API endpoints
- **Error scenarios**: Все error paths покрыты тестами

## 📊 Метрики успеха

### Технические метрики

- [ ] Repository method response time < 200ms (95th percentile)
- [ ] Error rate < 0.5% для всех операций
- [ ] Test coverage > 85% для нового кода
- [ ] Zero memory leaks в integration tests

### Продуктовые метрики

- [ ] Возможность создавать 1000+ заявок без деградации производительности
- [ ] Поддержка иерархий до 10 уровней глубиной
- [ ] Concurrent operations без data corruption

## 🚀 План внедрения

### Phase 1: TicketRepository (Priority: High)

**Срок**: 3-5 дней

- [ ] Интерфейс TicketRepository в interfaces.go
- [ ] MongoDB реализация с основными CRUD операциями
- [ ] Unit tests для domain logic
- [ ] Integration tests с MongoDB
- [ ] Mock для TicketRepository

### Phase 2: CategoryRepository (Priority: Medium)

**Срок**: 2-3 дня

- [ ] Интерфейс CategoryRepository с поддержкой иерархий
- [ ] MongoDB реализация с рекурсивными запросами
- [ ] Специализированные тесты для иерархических операций
- [ ] Mock для CategoryRepository

### Phase 3: OrganizationRepository (Priority: Medium)

**Срок**: 2-3 дня

- [ ] Интерфейс OrganizationRepository
- [ ] MongoDB реализация с иерархической логикой
- [ ] Тесты для организационной структуры
- [ ] Mock для OrganizationRepository

### Phase 4: Integration & Optimization (Priority: Low)

**Срок**: 2-3 дня

- [ ] Cross-repository операции
- [ ] Performance optimizations
- [ ] Advanced filtering capabilities
- [ ] Comprehensive error handling
- [ ] Documentation updates

### Definition of Done для каждой фазы

- [ ] Код следует установленным conventions (English comments, clean architecture)
- [ ] `make lint` проходит без ошибок
- [ ] `make test` показывает > 85% coverage для новых компонентов
- [ ] Integration tests с MongoDB проходят
- [ ] Моки генерируются и используются в тестах
- [ ] Интерфейсы добавлены в interfaces.go
- [ ] Performance requirements выполнены

## 🔗 Зависимости

### Внешние зависимости

- **MongoDB Driver**: Уже используется для UserRepository
- **Testcontainers**: Уже настроен для integration testing
- **UUID library**: github.com/google/uuid уже используется

### Внутренние зависимости

- **Domain models**: Все domain entities уже реализованы
- **Existing patterns**: Следование паттернам из UserRepository
- **Mock generation**: Использование существующего подхода к генерации моков

## ⚠️ Риски и ограничения

### Технические риски

- **Performance impact**: Добавление индексов может замедлить write operations
- **Concurrent access**: Race conditions при concurrent updates

### Митигация рисков

- **Batch operations**: Реализация batch методов для bulk operations
- **Connection pooling**: Правильная настройка MongoDB connection pool
- **Optimistic locking**: Версионирование записей для предотвращения data corruption
- **Graceful degradation**: Fallback механизмы при недоступности MongoDB

### Ограничения

- **Hierarchy depth**: Рекурсивные запросы ограничены разумной глубиной (10 уровней)
- **Batch size**: Batch operations ограничены разумным размером (1000 records)

## 📝 Дополнительные заметки

### Open Questions

- [ ] Нужна ли поддержка soft delete для всех entities?
- [ ] Требуется ли audit log для tracking изменений?
- [ ] Нужны ли full-text search capabilities для tickets?

### Future Considerations

- **Caching layer**: Redis кеширование для часто запрашиваемых данных
- **Event sourcing**: Возможность перехода на event-driven архитектуру
- **Sharding**: Horizontal scaling стратегии для больших datasets
- **Search engine**: Integration с Elasticsearch для advanced search

### Implementation Notes

- Использовать существующий паттерн с `createFn` и `updateFn` для обеспечения consistency
- Применять тот же подход к error handling что и в UserRepository
- Поддерживать единообразие в naming conventions и code structure
- Все MongoDB коллекции должны использовать одинаковый подход к индексации

---

> 💡 **Важно**: Эта спецификация должна быть реализована поэтапно, с обязательным прохождением `make lint` и `make test`
> после каждого этапа. Каждый репозиторий должен быть полностью протестирован перед переходом к следующему.