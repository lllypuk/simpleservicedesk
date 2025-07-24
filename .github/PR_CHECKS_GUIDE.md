# 🚀 Руководство по автоматическим проверкам PR

Этот документ описывает настроенную систему автоматических проверок для пулл реквестов.

## 📋 Автоматические проверки

При создании или обновлении пулл реквеста автоматически запускаются следующие проверки:

### ✅ Основные проверки (`pr-checks.yml`)

1. **Lint Code** - Проверка стиля и форматирования кода
   - `go fmt` - форматирование кода
   - `goimports` - организация импортов
   - `golangci-lint` - 30+ статических анализаторов

2. **Run Tests** - Запуск всех тестов
   - Unit тесты: `make unit_test`
   - Integration тесты: `make integration_test` (с MongoDB в Docker)

3. **Check Code Generation** - Проверка актуальности сгенерированного кода
   - Выполняет `make generate`
   - Проверяет, что нет незакоммиченных изменений

4. **Coverage Report** - Отчет о покрытии тестами
   - Генерирует отчет о покрытии
   - Отправляет в Codecov (опционально)

5. **Security Scan** - Сканирование безопасности
   - Использует Gosec для поиска уязвимостей

### 🔒 Проверки безопасности (`security.yml`)

1. **Dependency Review** - Анализ новых зависимостей (только для PR)
2. **Vulnerability Check** - Сканирование уязвимостей с `govulncheck`
3. **License Check** - Проверка лицензий зависимостей

## 🛠 Локальная разработка

Перед созданием PR рекомендуется запустить проверки локально:

```bash
# Запуск всех тестов
make test

# Проверка линтинга
make lint

# Генерация кода
make generate

# Отчет о покрытии
make coverage_report
```

## 🔧 Настройка окружения

### Для разработчиков

1. Установите необходимые инструменты:
```bash
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

2. Создайте `.env` файл:
```bash
MONGO_URI=mongodb://localhost:27017/simpleservicedesk
ENVIRONMENT=development
```

### Для администраторов репозитория

1. **Замените `@yourusername`** в `.github/CODEOWNERS` на реальные GitHub username
2. **Настройте правила защиты веток** (см. `.github/BRANCH_PROTECTION_SETUP.md`)
3. **Настройте Codecov** для отчетов о покрытии (опционально)

## 🚨 Что делать если проверки не проходят

### Ошибки линтинга
```bash
make lint
# Исправьте ошибки и повторите
```

### Тесты не проходят
```bash
make test
# Исправьте тесты или код
```

### Сгенерированный код устарел
```bash
make generate
git add .
git commit -m "Update generated code"
```

### Проблемы с безопасностью
```bash
# Проверьте уязвимости
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# Обновите зависимости
go get -u ./...
go mod tidy
```

## 📊 Статус проверок

Все проверки должны пройти успешно перед мержем PR. Статус проверок отображается:
- В интерфейсе GitHub в пулл реквесте
- В уведомлениях по email/Slack
- На странице Actions репозитория

## 🔄 Непрерывная интеграция

Проверки запускаются:
- ✅ При создании нового PR
- ✅ При push в ветку PR
- ✅ При push в основную ветку
- ✅ Еженедельно (проверки безопасности)

## 📝 Дополнительные ресурсы

- [Конфигурация golangci-lint](.golangci.yml)
- [Правила CODEOWNERS](.github/CODEOWNERS)
- [Настройка защиты веток](.github/BRANCH_PROTECTION_SETUP.md)
