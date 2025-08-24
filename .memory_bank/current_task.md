# Current Task - Phase 5: Extended Users API

## 🎯 Описание Phase 5: Extended Users API

Реализовать недостающие REST API endpoints для управления пользователями системы. Текущая система имеет только базовые операции создания и получения пользователей, необходимо добавить полный набор CRUD операций.

## 📋 Требования

### Функциональные требования
- **FR-U1**: `GET /users` - получение списка пользователей с фильтрацией
- **FR-U2**: `PUT /users/{id}` - обновление данных пользователя
- **FR-U3**: `DELETE /users/{id}` - удаление пользователя (мягкое удаление)
- **FR-U4**: `PATCH /users/{id}/role` - изменение роли пользователя

### API Endpoints

#### GET /users - Список пользователей
```yaml
/users:
  get:
    summary: List users with filtering
    parameters:
      - name: organization_id
        in: query
        schema:
          type: string
          format: uuid
      - name: role
        in: query
        schema:
          $ref: '#/components/schemas/UserRole'
      - name: is_active
        in: query
        schema:
          type: boolean
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
        description: List of users
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ListUsersResponse'
```

#### PUT /users/{id} - Обновление пользователя
```yaml
/users/{id}:
  put:
    summary: Update user
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
            $ref: '#/components/schemas/UpdateUserRequest'
    responses:
      '200':
        description: User updated successfully
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/GetUserResponse'
      '400':
        description: Invalid input data
      '404':
        description: User not found
      '409':
        description: Email already exists
```

#### DELETE /users/{id} - Удаление пользователя
```yaml
/users/{id}:
  delete:
    summary: Delete user (soft delete)
    parameters:
      - in: path
        name: id
        required: true
        schema:
          type: string
          format: uuid
    responses:
      '204':
        description: User deleted successfully
      '404':
        description: User not found
      '409':
        description: Cannot delete user with active tickets
```

#### PATCH /users/{id}/role - Изменение роли
```yaml
/users/{id}/role:
  patch:
    summary: Update user role
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
            $ref: '#/components/schemas/UpdateUserRoleRequest'
    responses:
      '200':
        description: User role updated successfully
      '400':
        description: Invalid role or insufficient permissions
      '404':
        description: User not found
```

### Схемы данных

```yaml
components:
  schemas:
    UpdateUserRequest:
      type: object
      properties:
        email:
          type: string
          format: email
          minLength: 1
          maxLength: 255
        full_name:
          type: string
          minLength: 1
          maxLength: 255
        organization_id:
          type: string
          format: uuid
        is_active:
          type: boolean

    UpdateUserRoleRequest:
      type: object
      required:
        - role
      properties:
        role:
          $ref: '#/components/schemas/UserRole'

    ListUsersResponse:
      type: object
      properties:
        users:
          type: array
          items:
            $ref: '#/components/schemas/GetUserResponse'
        pagination:
          $ref: '#/components/schemas/PaginationInfo'

    PaginationInfo:
      type: object
      properties:
        page:
          type: integer
        limit:
          type: integer
        total:
          type: integer
        total_pages:
          type: integer
```

## 🏗️ Техническая реализация

### Новые файлы для создания
- `internal/application/users/list.go` - хэндлер для получения списка пользователей
- `internal/application/users/update.go` - хэндлер для обновления пользователя
- `internal/application/users/delete.go` - хэндлер для удаления пользователя
- `internal/application/users/role.go` - хэндлер для изменения роли
- `internal/application/users/list_test.go` - тесты для списка пользователей
- `internal/application/users/update_test.go` - тесты для обновления
- `internal/application/users/delete_test.go` - тесты для удаления
- `internal/application/users/role_test.go` - тесты для изменения роли

### Обновления существующих файлов
- `api/openapi.yaml` - добавить новые endpoints
- `internal/application/users/handlers.go` - добавить новые методы в интерфейс
- `internal/run.go` - обновить регистрацию новых маршрутов

### Требования к реализации
1. **Фильтрация**: Поддержка фильтрации по organization_id, role, is_active
2. **Пагинация**: Использование offset/limit подхода
3. **Валидация**: Проверка входных данных согласно domain rules
4. **Безопасность**: Проверка прав доступа для изменения ролей
5. **Мягкое удаление**: Пользователи помечаются как неактивные, но не удаляются физически

## 🧪 План тестирования

### Unit Tests
- [ ] Список пользователей с различными фильтрами
- [ ] Обновление данных пользователя
- [ ] Валидация при обновлении (дублирующий email)
- [ ] Мягкое удаление пользователя
- [ ] Изменение роли пользователя
- [ ] Проверка прав доступа для операций

### Integration Tests
- [ ] GET /users с фильтрами и пагинацией
- [ ] PUT /users/{id} с валидными и невалидными данными
- [ ] DELETE /users/{id} для существующих и несуществующих пользователей
- [ ] PATCH /users/{id}/role с различными ролями
- [ ] Проверка HTTP статус кодов для всех endpoint'ов

## 🎯 Критерии готовности

- [ ] Все 4 новых endpoint'а реализованы
- [ ] OpenAPI спецификация обновлена
- [ ] Unit и integration тесты покрывают новую функциональность
- [ ] `make lint` проходит без ошибок
- [ ] `make test-all` проходит успешно
- [ ] Все endpoint'ы возвращают корректные HTTP статус коды
- [ ] Фильтрация и пагинация работают корректно
- [ ] Валидация входных данных реализована

## 📝 Примеры использования API

### Получение списка активных администраторов
```bash
GET /users?role=Admin&is_active=true&page=1&limit=10
```

### Обновление email пользователя
```bash
PUT /users/123e4567-e89b-12d3-a456-426614174000
{
  "email": "newemail@example.com",
  "full_name": "Updated Name"
}
```

### Изменение роли пользователя на агента
```bash
PATCH /users/123e4567-e89b-12d3-a456-426614174000/role
{
  "role": "Agent"
}
```

### Удаление пользователя
```bash
DELETE /users/123e4567-e89b-12d3-a456-426614174000
```
