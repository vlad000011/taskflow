# TaskFlow — трекер задач для студенческих команд

Учебный fullstack MVP по ТЗ практики: React + TypeScript, Go + Gin, PostgreSQL, REST API, JWT, RBAC, Docker Compose, миграции, тесты и GitHub Actions.


## Стек
- Frontend: React, TypeScript strict, Vite, React Router, TanStack Query, Zustand, Tailwind CSS, dnd-kit, Recharts, Vitest, Testing Library.
- Backend: Go, Gin, GORM, PostgreSQL, JWT, Argon2id, Swagger/OpenAPI.
- Infra: Docker, Docker Compose, Nginx, GitHub Actions.

## Быстрый запуск
1. Установите Docker Engine/Desktop и Docker Compose.
2. Скопируйте `.env.example` в `.env` и замените секреты.
3. Выполните `docker compose up --build`.
4. Frontend: http://localhost:8080; API: http://localhost:8080/api/v1; Swagger: http://localhost:8080/swagger/index.html.

Для локального запуска backend без Docker: `cd backend && go run ./cmd/api`. Для frontend: `cd frontend && npm install && npm run dev`.

## Демонстрационные учётные записи
Seed-команда создаёт учётные записи при `SEED_DEMO=true`:
- admin@taskflow.local / `ChangeMe123!` — администратор
- lead@taskflow.local / `ChangeMe123!` — тимлид
- student@taskflow.local / `ChangeMe123!` — участник
- teacher@taskflow.local / `ChangeMe123!` — преподаватель

Смените пароли перед публичным развёртыванием. В демо-режиме seed включён только для разработки.

## Команды разработки
```bash
cd backend && go test ./... && gofmt -w .
cd frontend && npm install && npm run lint && npm test && npm run build
```

## Структура
- `backend/` — REST API, слои handler/service/repository, миграции и тесты.
- `frontend/` — SPA.
- `docs/` — требования, user stories, архитектура, тест-план и OpenAPI.
- `deploy/` — Nginx.
- `.github/workflows/ci.yml` — CI.

## Реализованный MVP
- Регистрация/вход, access + refresh JWT, logout, запрос сброса пароля с ссылкой в лог.
- Роли: `admin`, `participant`, `lead`, `teacher`; права проверяются на сервере.
- Проекты, задачи, статусы, спринты, комментарии, журнал изменений, участники.
- Серверные поиск/фильтрация/сортировка/пагинация задач.
- Загрузка файлов с ограничением размера и типа.
- Профиль и смена пароля; администрирование пользователей и ролей.
- Сводная статистика для дашборда.
- OpenAPI, Docker Compose, CI и базовые тесты.


