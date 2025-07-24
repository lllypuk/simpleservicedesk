# Настройка правил защиты веток для GitHub

Для полной настройки автоматических проверок в пулл реквестах, необходимо также настроить правила защиты веток в GitHub репозитории.

## Как настроить правила защиты веток:

1. Перейдите в настройки репозитория: Settings → Branches
2. Нажмите "Add rule" для создания нового правила защиты ветки
3. Укажите pattern: `main` (или `master`, в зависимости от вашей основной ветки)

## Рекомендуемые настройки:

### Require a pull request before merging
- ✅ Require a pull request before merging
- ✅ Require approvals (минимум 1)
- ✅ Dismiss stale pull request approvals when new commits are pushed
- ✅ Require review from code owners (если есть CODEOWNERS файл)

### Require status checks to pass before merging
- ✅ Require status checks to pass before merging
- ✅ Require branches to be up to date before merging

Обязательные проверки (выберите из списка):
- `Lint Code`
- `Run Tests`
- `Check Code Generation`
- `Coverage Report`
- `Security Scan`
- `Dependency Review` (для PR)
- `Vulnerability Check`

### Дополнительные настройки
- ✅ Restrict pushes that create files larger than 100 MB
- ✅ Require linear history (опционально)
- ✅ Include administrators (применять правила к администраторам)

## Создание CODEOWNERS файла (опционально)

Создайте файл `.github/CODEOWNERS` для автоматического назначения ревьюеров:

```
# Все изменения требуют ревью от владельцев
* @yourusername

# API изменения требуют дополнительного ревью
/api/ @yourusername @apiteam

# Инфраструктурные изменения
/infrastructure/ @yourusername @devops-team
/.github/ @yourusername @devops-team
/docker-compose.yml @yourusername @devops-team
/Dockerfile @yourusername @devops-team
```

## Настройка уведомлений

В Settings → Notifications настройте уведомления для:
- Failed workflow runs
- Pull request reviews
- Security alerts
