# Remaining Entity Handlers - API Implementation

> 📝 **Техническое задание**: Реализация недостающих REST API хэндлеров для сущностей системы

## 🎯 Описание задачи

### Проблема
Текущая OpenAPI спецификация и система содержит только базовые хэндлеры для пользователей (Users). Отсутствуют REST API endpoints для остальных основных сущностей системы:
- **Tickets** - основная бизнес-логика service desk
- **Organizations** - управление организациями
- **Categories** - категоризация тикетов
- Недостающие операции для Users (update, delete, list)

### Решение
Реализовать полный набор REST API endpoints для всех сущностей системы, включая:
- CRUD операции для каждой сущности
- Соответствующие хэндлеры в application layer
- Обновление OpenAPI спецификации
- Генерация кода с помощью oapi-codegen

## 📋 Требования

### Функциональные требования

#### Tickets API
- [ ] **FR-T1**: `POST /tickets` - создание нового тикета
- [ ] **FR-T2**: `GET /tickets/{id}` - получение тикета по ID
- [ ] **FR-T3**: `GET /tickets` - получение списка тикетов с фильтрацией
- [ ] **FR-T4**: `PUT /tickets/{id}` - обновление тикета
- [ ] **FR-T5**: `PATCH /tickets/{id}/status` - изменение статуса тикета
- [ ] **FR-T6**: `DELETE /tickets/{id}` - удаление тикета
- [ ] **FR-T7**: `POST /tickets/{id}/comments` - добавление комментария

#### Organizations API
- [ ] **FR-O1**: `POST /organizations` - создание организации
- [ ] **FR-O2**: `GET /organizations/{id}` - получение организации по ID
- [ ] **FR-O3**: `GET /organizations` - получение списка организаций
- [ ] **FR-O4**: `PUT /organizations/{id}` - обновление организации
- [ ] **FR-O5**: `DELETE /organizations/{id}` - удаление организации
- [ ] **FR-O6**: `GET /organizations/{id}/users` - пользователи организации

#### Categories API
- [ ] **FR-C1**: `POST /categories` - создание категории
- [ ] **FR-C2**: `GET /categories/{id}` - получение категории по ID
- [ ] **FR-C3**: `GET /categories` - получение дерева категорий
- [ ] **FR-C4**: `PUT /categories/{id}` - обновление категории
- [ ] **FR-C5**: `DELETE /categories/{id}` - удаление категории

#### Extended Users API
- [ ] **FR-U1**: `GET /users` - получение списка пользователей
- [ ] **FR-U2**: `PUT /users/{id}` - обновление пользователя
- [ ] **FR-U3**: `DELETE /users/{id}` - удаление пользователя
- [ ] **FR-U4**: `PATCH /users/{id}/role` - изменение роли пользователя

### Нефункциональные требования
- [ ] **NFR-1**: Все endpoints должны возвращать ответ < 200ms
- [ ] **NFR-2**: Поддержка пагинации для list endpoints
- [ ] **NFR-3**: Валидация входных данных согласно domain rules
- [ ] **NFR-4**: Обработка всех возможных HTTP статус кодов
- [ ] **NFR-5**: Consistent error response format

## 🏗️ Техническая архитектура

### Затронутые компоненты
- **Domain Layer**: Существующие модели в `internal/domain/{tickets,organizations,categories,users}/`
- **Application Layer**: Новые хэндлеры в `internal/application/handlers/`
- **Infrastructure Layer**: Repository implementations уже существуют
- **API**: Расширение `api/openapi.yaml`
- **Generated Code**: Полное обновление `generated/` после изменений в OpenAPI

### Новые API endpoints

```yaml
# Tickets endpoints
paths:
  /tickets:
    post:
      summary: Create a new ticket
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateTicketRequest'
      responses:
        '201':
          description: Ticket created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetTicketResponse'
    get:
      summary: List tickets with filtering
      parameters:
        - name: status
          in: query
          schema:
            $ref: '#/components/schemas/TicketStatus'
        - name: priority
          in: query
          schema:
            $ref: '#/components/schemas/TicketPriority'
        - name: category_id
          in: query
          schema:
            type: string
            format: uuid
        - name: assignee_id
          in: query
          schema:
            type: string
            format: uuid
        - name: page
          in: query
          schema:
            type: integer
            minimum: 1
            default: 1
        - name: limit
          in: query
          schema:
            type: integer
            minimum: 1
            maximum: 100
            default: 20
      responses:
        '200':
          description: List of tickets
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListTicketsResponse'

  /tickets/{id}:
    get:
      summary: Get ticket by ID
      parameters:
        - in: path
          name: id
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Ticket details
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetTicketResponse'
    put:
      summary: Update ticket
      parameters:
        - in: path
          name: id
          required: true
          schema:
            type: string
            format: uuid
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateTicketRequest'
      responses:
        '200':
          description: Ticket updated successfully

  /tickets/{id}/status:
    patch:
      summary: Update ticket status
      parameters:
        - in: path
          name: id
          required: true
          schema:
            type: string
            format: uuid
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateTicketStatusRequest'
      responses:
        '200':
          description: Ticket status updated successfully

# Organizations endpoints
  /organizations:
    post:
      summary: Create a new organization
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateOrganizationRequest'
      responses:
        '201':
          description: Organization created successfully
    get:
      summary: List organizations
      parameters:
        - name: page
          in: query
          schema:
            type: integer
            minimum: 1
            default: 1
        - name: limit
          in: query
          schema:
            type: integer
            minimum: 1
            maximum: 100
            default: 20
      responses:
        '200':
          description: List of organizations

  /organizations/{id}:
    get:
      summary: Get organization by ID
      parameters:
        - in: path
          name: id
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Organization details

# Categories endpoints
  /categories:
    post:
      summary: Create a new category
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateCategoryRequest'
      responses:
        '201':
          description: Category created successfully
    get:
      summary: Get categories tree
      responses:
        '200':
          description: Categories tree structure

  /categories/{id}:
    get:
      summary: Get category by ID
      parameters:
        - in: path
          name: id
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Category details
```

### Схемы данных
Необходимо добавить следующие схемы в OpenAPI:

```yaml
components:
  schemas:
    # Ticket schemas
    TicketStatus:
      type: string
      enum: [Open, InProgress, Resolved, Closed]

    TicketPriority:
      type: string
      enum: [Low, Medium, High, Critical]

    CreateTicketRequest:
      type: object
      required:
        - title
        - description
        - priority
        - category_id
        - requester_id
      properties:
        title:
          type: string
          minLength: 1
          maxLength: 200
        description:
          type: string
          minLength: 1
        priority:
          $ref: '#/components/schemas/TicketPriority'
        category_id:
          type: string
          format: uuid
        requester_id:
          type: string
          format: uuid
        assignee_id:
          type: string
          format: uuid
        organization_id:
          type: string
          format: uuid

    GetTicketResponse:
      type: object
      properties:
        id:
          type: string
          format: uuid
        title:
          type: string
        description:
          type: string
        status:
          $ref: '#/components/schemas/TicketStatus'
        priority:
          $ref: '#/components/schemas/TicketPriority'
        category_id:
          type: string
          format: uuid
        requester_id:
          type: string
          format: uuid
        assignee_id:
          type: string
          format: uuid
        organization_id:
          type: string
          format: uuid
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time

    # Organization schemas
    CreateOrganizationRequest:
      type: object
      required:
        - name
      properties:
        name:
          type: string
          minLength: 1
          maxLength: 100
        description:
          type: string
        parent_id:
          type: string
          format: uuid

    # Category schemas
    CreateCategoryRequest:
      type: object
      required:
        - name
      properties:
        name:
          type: string
          minLength: 1
          maxLength: 100
        description:
          type: string
        parent_id:
          type: string
          format: uuid
```

## 🔄 User Stories

### История 1: Администратор хочет создать тикет для пользователя
**Как** администратор системы
**Я хочу** создать тикет от имени пользователя
**Чтобы** быстро регистрировать проблемы, поступающие по телефону или email

**Критерии приемки:**
- [ ] Given авторизованный администратор, when отправляет POST /tickets с валидными данными, then тикет создается со статусом "Open"
- [ ] Given невалидные данные, when отправляет POST /tickets, then возвращается 400 с описанием ошибки
- [ ] Given несуществующий category_id, when отправляет POST /tickets, then возвращается 400 с ошибкой "Category not found"

### История 2: Агент хочет обновить статус тикета
**Как** агент поддержки
**Я хочу** изменить статус тикета на "InProgress"
**Чтобы** показать что работаю над решением проблемы

**Критерии приемки:**
- [ ] Given тикет со статусом "Open", when агент отправляет PATCH /tickets/{id}/status, then статус меняется на "InProgress"
- [ ] Given тикет со статусом "Closed", when агент пытается изменить статус, then возвращается 400 "Invalid status transition"

### История 3: Пользователь хочет просмотреть свои тикеты
**Как** пользователь системы
**Я хочу** получить список своих тикетов с фильтрацией по статусу
**Чтобы** отслеживать прогресс решения проблем

**Критерии приемки:**
- [ ] Given авторизованный пользователь, when отправляет GET /tickets?status=Open, then получает только свои открытые тикеты
- [ ] Given запрос с пагинацией, when указывает page=2&limit=10, then получает правильную страницу результатов

## 🧪 План тестирования

### Unit Tests
- [ ] Domain logic для всех новых хэндлеров
- [ ] Validation входных данных для каждого endpoint
- [ ] Error handling для edge cases
- [ ] Status transition logic для тикетов

### Integration Tests
- [ ] API endpoints с реальной MongoDB
- [ ] CRUD операции для каждой сущности
- [ ] Фильтрация и пагинация для list endpoints
- [ ] Authorization и access control

### Manual Testing Scenarios
1. **Создание тикета через API**:
   - Отправить POST /tickets с валидными данными
   - Проверить что тикет создан в БД
   - Проверить что возвращается корректный response

2. **Обновление статуса тикета**:
   - Создать тикет со статусом "Open"
   - Изменить статус на "InProgress" через PATCH
   - Проверить что статус обновлен

3. **Фильтрация тикетов**:
   - Создать тикеты с разными статусами
   - Запросить тикеты с фильтром status=Open
   - Проверить что возвращаются только открытые тикеты

## 📊 Метрики успеха

### Технические метрики
- [ ] API response time < 200ms для всех endpoints
- [ ] Error rate < 1% в production
- [ ] Test coverage > 85% для новых хэндлеров
- [ ] Zero linting errors после `make lint`

### Продуктовые метрики
- [ ] Successful API calls rate > 99%
- [ ] Complete CRUD functionality для всех сущностей
- [ ] Consistent response format across all endpoints

## 🚀 План внедрения

### Этапы разработки
1. **Phase 1**: OpenAPI спецификация
   - Расширить api/openapi.yaml со всеми endpoints
   - Сгенерировать код с помощью `make generate`
   - Проверить что генерация проходит без ошибок

2. **Phase 2**: Tickets API
   - Реализовать хэндлеры для tickets
   - Unit и integration tests
   - Тестирование через API

3. **Phase 3**: Organizations API
   - Реализовать хэндлеры для organizations
   - Unit и integration tests
   - Тестирование через API

4. **Phase 4**: ✅ Completed - Categories API moved to current_task.md

5. **Phase 5**: Extended Users API
   - Добавить недостающие операции для users
   - Unit и integration tests
   - Полное тестирование всего API

### Definition of Done
- [ ] Все endpoints реализованы и работают корректно
- [ ] OpenAPI спецификация обновлена и корректна
- [ ] Код сгенерирован и соответствует спецификации
- [ ] Unit tests покрывают 85%+ нового кода
- [ ] Integration tests проходят для всех endpoints
- [ ] `make lint` проходит без ошибок
- [ ] `make test-all` проходит без ошибок
- [ ] API документация обновлена
- [ ] Manual testing completed successfully

## 🔗 Зависимости

### Внешние зависимости
- oapi-codegen для генерации кода
- Существующие domain models
- MongoDB repositories (уже реализованы)

### Внутренние зависимости
- Clean Architecture должна быть соблюдена
- Соответствие существующим patterns в коде
- Интеграция с существующими middleware

## ⚠️ Риски и ограничения

### Технические риски
- Большое количество изменений в OpenAPI может вызвать конфликты при генерации
- Breaking changes в существующих endpoints
- Performance impact от новых endpoints

### Митигация рисков
- Поэтапное добавление endpoints с тестированием на каждом этапе
- Сохранение обратной совместимости для существующих endpoints
- Performance testing после каждого этапа

## 📝 Дополнительные заметки

### Open Questions
- [ ] Нужна ли авторизация на уровне endpoints или middleware?
- [ ] Какая стратегия pagination: offset/limit или cursor-based?
- [ ] Нужно ли логирование всех API calls?

### Future Considerations
- Добавление rate limiting для API endpoints
- API versioning стратегия
- OpenAPI documentation UI (Swagger)
- API metrics и monitoring

---

> 💡 **Важно**: Этот документ описывает реализацию полного REST API для всех сущностей системы. Реализация должна строго следовать принципам Clean Architecture и существующим patterns в коде.
