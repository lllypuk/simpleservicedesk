# Current Task - Phase 4: Categories API

## 🎯 Цель Phase 4

Реализовать REST API endpoints для управления категориями тикетов, включая поддержку иерархической структуры категорий.

## 📋 Требования

### Categories API Endpoints
- [ ] **FR-C1**: `POST /categories` - создание категории
- [ ] **FR-C2**: `GET /categories/{id}` - получение категории по ID
- [ ] **FR-C3**: `GET /categories` - получение дерева категорий
- [ ] **FR-C4**: `PUT /categories/{id}` - обновление категории
- [ ] **FR-C5**: `DELETE /categories/{id}` - удаление категории

## 🏗️ Техническая реализация

### API Endpoints

```yaml
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

```yaml
components:
  schemas:
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

## 🧪 План тестирования

### Unit Tests
- [ ] Domain logic для category хэндлеров
- [ ] Validation входных данных
- [ ] Error handling для edge cases
- [ ] Hierarchical structure logic

### Integration Tests
- [ ] API endpoints с реальной MongoDB
- [ ] CRUD операции для categories
- [ ] Tree structure operations
- [ ] Parent-child relationships

## ✅ Критерии готовности
- [ ] Все endpoints реализованы
- [ ] Unit tests покрывают 85%+ кода
- [ ] Integration tests проходят
- [ ] `make lint` проходит без ошибок
- [ ] `make test-all` проходит без ошибок
