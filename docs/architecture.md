# Архитектура

```text
Browser
  └── Nginx :8080
       ├── /             → React SPA (static files)
       ├── /api/v1/*     → Go + Gin REST API
       └── /uploads/*    → backend uploaded files
                         └── GORM → PostgreSQL
```

Backend организован по слоям: HTTP handlers → business/service helpers → GORM models/database. JWT middleware аутентифицирует пользователя; отдельный role middleware защищает административные endpoints. Для каждого проекта проверяется membership на сервере.

## Данные
`users`, `projects`, `project_members`, `tasks`, `comments`, `sprints`, `sprint_tasks`, `task_history`, `refresh_tokens`, `password_reset_tokens`.

## Ошибки
Единый формат: `{ "error": { "code": "...", "message": "..." } }`. Успешные ответы обёрнуты в `{ "data": ... }`.

## Конфигурация
Секреты и адрес БД задаются переменными окружения. `.env.example` содержит только примеры. В публичной среде нужно задать сильный `JWT_SECRET`, пароль БД и точный `CORS_ORIGIN`.
