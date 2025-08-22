# Current Task - Текущая задача

## 🎯 Активная задача

Добавь репозитории для этих сущностей:

- **Categories**: Система категоризации заявок для организации и фильтрации
- **Organizations**: Организационная структура для группировки пользователей и заявок

## 📋 Детальные требования

### Функциональные требования
- **FR-2**: CategoryRepository с поддержкой иерархической структуры (parent-child)
- **FR-3**: OrganizationRepository с иерархической структурой организаций
- **FR-4**: Связи между сущностями (Ticket-Category, User-Organization)
- **FR-5**: Поддержка фильтрации и сортировки для всех репозиториев

### Нефункциональные требования
- **NFR-1**: Производительность (время отклика < 200ms для простых запросов)
- **NFR-3**: Безопасность (валидация данных на уровне репозиториев)
- **NFR-5**: Совместимость (единый интерфейс для MongoDB реализаций)

## 🏗️ Техническая реализация

### Новые интерфейсы репозиториев

```go
// CategoryRepository - управление категориями
type CategoryRepository interface {
    CreateCategory(ctx context.Context, createFn func() (*categories.Category, error)) (*categories.Category, error)
    UpdateCategory(ctx context.Context, id uuid.UUID, updateFn func(*categories.Category) (bool, error)) (*categories.Category, error)
    GetCategory(ctx context.Context, id uuid.UUID) (*categories.Category, error)
    ListCategories(ctx context.Context, filter CategoryFilter) ([]*categories.Category, error)
    GetCategoryHierarchy(ctx context.Context, rootID uuid.UUID) (*categories.CategoryTree, error)
    DeleteCategory(ctx context.Context, id uuid.UUID) error
}

// OrganizationRepository - управление организациями
type OrganizationRepository interface {
    CreateOrganization(ctx context.Context, createFn func() (*organizations.Organization, error)) (*organizations.Organization, error)
    UpdateOrganization(ctx context.Context, id uuid.UUID, updateFn func(*organizations.Organization) (bool, error)) (*organizations.Organization, error)
    GetOrganization(ctx context.Context, id uuid.UUID) (*organizations.Organization, error)
    ListOrganizations(ctx context.Context, filter OrganizationFilter) ([]*organizations.Organization, error)
    GetOrganizationHierarchy(ctx context.Context, rootID uuid.UUID) (*organizations.OrganizationTree, error)
    DeleteOrganization(ctx context.Context, id uuid.UUID) error
}
```

### Затронутые компоненты
- **Application Layer**: `internal/application/interfaces.go` - новые интерфейсы репозиториев
- **Infrastructure Layer**: 
  - `internal/infrastructure/categories/` - MongoDB реализация
  - `internal/infrastructure/organizations/` - MongoDB реализация
- **Testing**: Моки для всех новых репозиториев в `internal/application/mocks/`

### MongoDB коллекции и индексы
```javascript
// categories collection
db.categories.createIndex({ "parent_id": 1 })
db.categories.createIndex({ "name": 1 }, { unique: true })

// organizations collection  
db.organizations.createIndex({ "parent_id": 1 })
db.organizations.createIndex({ "name": 1 }, { unique: true })
```

## 🔄 User Stories

### История 1: Admin хочет управлять категориями
**Как** администратор системы  
**Я хочу** создавать иерархию категорий  
**Чтобы** организовать классификацию заявок  

**Критерии приемки:**
- [ ] Given родительскую категорию, when создаю подкатегорию, then устанавливается parent-child связь
- [ ] Given категорию с подкategориями, when получаю иерархию, then возвращается полное дерево
- [ ] Given категорию с заявками, when пытаюсь удалить, then операция блокируется

### История 2: Admin хочет управлять организационной структурой
**Как** администратор системы  
**Я хочу** создавать иерархию организаций  
**Чтобы** группировать пользователей и контролировать доступ  

**Критерии приемки:**
- [ ] Given организацию, when создаю подразделение, then устанавливается иерархическая связь
- [ ] Given организацию, when получаю всех пользователей, then возвращаются пользователи всех подразделений
- [ ] Given организацию с пользователями, when пытаюсь удалить, then операция блокируется

## 🚀 План внедрения

### Phase 1: CategoryRepository (Priority: Medium) ✅ COMPLETED
**Срок**: 2-3 дня  
- [x] Интерфейс CategoryRepository с поддержкой иерархий
- [x] MongoDB реализация с рекурсивными запросами
- [x] Специализированные тесты для иерархических операций
- [x] Mock для CategoryRepository

### Phase 2: OrganizationRepository (Priority: Medium) ✅ COMPLETED
**Срок**: 2-3 дня  
- [x] Интерфейс OrganizationRepository
- [x] MongoDB реализация с иерархической логикой
- [x] Тесты для организационной структуры
- [x] Mock для OrganizationRepository

### Definition of Done для каждой фазы ✅ COMPLETED
- [x] Код следует установленным conventions (English comments, clean architecture)
- [x] `make lint` проходит без ошибок
- [x] `make test` показывает > 85% coverage для новых компонентов
- [x] Integration tests с MongoDB проходят
- [x] Моки генерируются и используются в тестах
- [x] Интерфейсы добавлены в interfaces.go
- [x] Performance requirements выполнены

## 🧪 План тестирования

### Unit Tests
- [ ] **Domain logic**: Валидация бизнес-правил для каждой сущности
- [ ] **Repository methods**: Тестирование CRUD операций с моками
- [ ] **Filter logic**: Проверка корректности фильтрации и сортировки
- [ ] **Error handling**: Обработка всех возможных ошибочных ситуаций

### Integration Tests
- [ ] **MongoDB repositories**: Тестирование с реальной БД через testcontainers
- [ ] **Cross-repository operations**: Операции затрагивающие несколько репозиториев
- [ ] **Performance**: Load testing для критичных операций

### Test Coverage Goals
- **Unit tests**: > 90% для репозиториев
- **Integration tests**: 100% покрытие API endpoints
- **Error scenarios**: Все error paths покрыты тестами

## ⚠️ Важные заметки

### Технические ограничения
- **Hierarchy depth**: Рекурсивные запросы ограничены разумной глубиной (10 уровней)

### Implementation Notes
- Использовать существующий паттерн с `createFn` и `updateFn` для обеспечения consistency
- Применять тот же подход к error handling что и в UserRepository
- Поддерживать единообразие в naming conventions и code structure
- Все MongoDB коллекции должны использовать одинаковый подход к индексации
