# TaskFlow

**TaskFlow** — веб-приложение для управления задачами учебных команд. Проект объединяет работу с проектами, задачами, участниками и спринтами в одном интерфейсе.

## Возможности

- Регистрация, вход и выход из системы.
- Аутентификация на основе access- и refresh-JWT.
- Ролевая модель и проверка прав доступа на backend.
- Создание и управление проектами и задачами.
- Канбан-доска со статусами задач.
- Поиск, фильтрация, сортировка и серверная пагинация задач.
- Назначение исполнителей, приоритетов и сроков.
- Комментарии к задачам и история изменений.
- Создание спринтов и привязка к ним задач.
- Управление участниками проекта.
- Прикрепление и скачивание файлов.
- Профиль пользователя и смена пароля.
- Администрирование глобальных ролей и блокировок пользователей.
- Дашборд со сводной статистикой проекта.
- Переключение темы интерфейса.

## Технологический стек

| Область | Технологии |
|---|---|
| Frontend | React 18, TypeScript, Vite |
| Маршрутизация | React Router |
| Работа с серверными данными | TanStack Query, Axios |
| Клиентское состояние | Zustand |
| Интерфейс | Tailwind CSS, lucide-react |
| Drag & Drop | dnd-kit |
| Графики | Recharts |
| Frontend-тесты | Vitest, Testing Library, jsdom |
| Backend | Go 1.23, Gin |
| ORM и база данных | GORM, PostgreSQL 16 |
| Аутентификация | JWT, Argon2id |
| API-документация | OpenAPI 3.0.3, Swagger UI |
| Инфраструктура | Docker, Docker Compose, Nginx |
| CI | GitHub Actions |

## Архитектура

Приложение состоит из нескольких основных компонентов:

1. **Frontend** — одностраничное React-приложение. Оно отображает интерфейс, управляет пользовательским состоянием и обращается к REST API.
2. **Backend** — HTTP API на Go и Gin. Обрабатывает запросы, выполняет валидацию, проверяет аутентификацию и права доступа, взаимодействует с базой данных.
3. **PostgreSQL** — хранит пользователей, проекты, участников, задачи, спринты, комментарии, токены и историю изменений.
4. **Nginx** — единая точка входа для frontend, API и Swagger UI.
5. **Docker Compose** — запускает инфраструктурные сервисы приложения.

Упрощённый поток запроса:

```text
Browser
   |
   v
 Nginx :8080
   |----------------------|
   v                      v
Frontend               Backend :8081
                          |
                          v
                      PostgreSQL
```

## Быстрый запуск через Docker Compose

### Требования

- Docker Desktop или Docker Engine.
- Docker Compose v2.
- Git.

### 1. Получение проекта

```bash
git clone https://github.com/vlad000011/taskflow.git
cd taskflow
```

### 2. Настройка окружения

Создайте локальный `.env` на основе примера:

```bash
cp .env.example .env
```

Для локального учебного запуска можно оставить значения из примера. Перед публикацией или развёртыванием в доступной извне среде обязательно замените пароли и секреты на собственные случайные значения.

### 3. Запуск

```bash
docker compose up -d --build
```

Проверить состояние контейнеров:

```bash
docker compose ps
```

Посмотреть логи:

```bash
docker compose logs -f
```

Остановить приложение, сохранив данные PostgreSQL:

```bash
docker compose down
```

> Не используйте `docker compose down -v`, если хотите сохранить данные: флаг `-v` удаляет именованные тома, включая том базы данных и загруженных файлов.

### Адреса после запуска

| Компонент | Адрес |
|---|---|
| Web-приложение | http://localhost:8080 |
| REST API | http://localhost:8080/api/v1 |
| Проверка доступности API | http://localhost:8080/api/v1/health |
| Swagger UI | http://localhost:8080/swagger/index.html |
| OpenAPI YAML | http://localhost:8080/swagger/openapi.yaml |

Swagger UI загружает `backend/docs/openapi.yaml`. Для загрузки интерфейса Swagger UI используются ресурсы `swagger-ui-dist` с CDN, поэтому браузеру требуется доступ к сети.

## Демонстрационные учётные записи

При `SEED_DEMO=true` backend создаёт демонстрационные учётные записи и учебный проект, если таблица пользователей ещё пуста.

| Роль | Email | Пароль |
|---|---|---|
| Администратор | `admin@taskflow.local` | `ChangeMe123!` |
| Тимлид | `lead@taskflow.local` | `ChangeMe123!` |
| Участник | `student@taskflow.local` | `ChangeMe123!` |
| Преподаватель | `teacher@taskflow.local` | `ChangeMe123!` |

Это тестовые данные для локальной демонстрации. Не используйте указанные пароли в публичном развёртывании. Если база уже содержит пользователей, повторный запуск seed не пересоздаёт демонстрационные аккаунты.

## Роли и доступ

В приложении используются глобальные роли и роли участника внутри конкретного проекта.

- **Администратор (`admin`)** — имеет глобальные административные права, включая управление пользователями и доступ к проектам.
- **Тимлид (`lead`)** — управляет проектом, задачами, спринтами и составом команды в рамках проекта.
- **Участник (`participant`)** — работает с доступными ему задачами, может менять статус назначенной ему задачи и оставлять комментарии.
- **Преподаватель (`teacher`)** — имеет доступ на просмотр в рамках проекта.

Права проверяются на backend. Наличие или отсутствие кнопки в интерфейсе не заменяет серверную проверку доступа.

## Конфигурация

Основные переменные окружения из `.env.example`:

| Переменная | Назначение | Значение примера |
|---|---|---|
| `POSTGRES_DB` | Имя базы данных | `taskflow` |
| `POSTGRES_USER` | Пользователь PostgreSQL | `taskflow` |
| `POSTGRES_PASSWORD` | Пароль PostgreSQL | `change-this-db-password` |
| `DATABASE_URL` | Строка подключения backend к PostgreSQL | `postgres://...@db:5432/taskflow?sslmode=disable` |
| `JWT_SECRET` | Секрет подписи JWT, минимум 32 символа | `replace-with-at-least-32-random-bytes` |
| `ACCESS_TOKEN_MINUTES` | Срок действия access-токена в минутах | `15` |
| `REFRESH_TOKEN_DAYS` | Срок действия refresh-токена в днях | `14` |
| `CORS_ORIGIN` | Разрешённый origin frontend | `http://localhost:8080` |
| `API_PORT` | Порт backend внутри контейнера | `8081` |
| `SEED_DEMO` | Создание демонстрационных данных | `true` |
| `UPLOAD_DIR` | Каталог хранения вложений | `/app/uploads` |
| `MAX_UPLOAD_MB` | Максимальный размер загружаемого файла в МБ | `5` |

В Docker Compose PostgreSQL доступен другим контейнерам по имени сервиса `db` и порту `5432`. Порт `5433` опубликован на localhost для подключения с хост-машины.

## Локальный запуск без Docker

Для локального запуска нужны Go 1.23+, Node.js и npm, а также доступный экземпляр PostgreSQL.

### Backend

1. Создайте базу PostgreSQL.
2. Перейдите в каталог backend.
3. Перед запуском задайте переменные окружения для подключения к своей базе и JWT-секрет.

Пример для macOS/Linux, если PostgreSQL запущен через Compose и опубликован на `localhost:5433`:

```bash
cd backend
export DATABASE_URL='postgres://taskflow:change-this-db-password@localhost:5433/taskflow?sslmode=disable'
export JWT_SECRET='replace-with-a-random-secret-of-at-least-32-characters'
export CORS_ORIGIN='http://localhost:5173'
export SEED_DEMO=true
go run ./cmd/api
```

Backend запускается на порту `8081` по умолчанию. Миграции загружаются из `backend/migrations` относительно текущего рабочего каталога, поэтому команду `go run` следует выполнять из `backend`.

### Frontend

В отдельном терминале:

```bash
cd frontend
npm install
npm run dev
```

Vite запускает frontend на адресе, который выводит терминал. Для локального frontend задайте `VITE_API_URL`, если API находится не по ожидаемому адресу `/api/v1`. Для полноценной работы локального dev-сервера также настройте проксирование `/api` на backend либо используйте единый Docker/Nginx запуск, описанный выше.

## Тестирование и проверки

### Backend

Из каталога `backend`:

```bash
go test ./...
go test -cover ./...
go build ./...
```

Чтобы сформировать файл покрытия и посмотреть детализацию:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### Frontend

Из каталога `frontend`:

```bash
npm install
npm run lint
npm test -- --run
npm run build
```

`npm test -- --run` запускает Vitest в однократном режиме. `npm run build` выполняет проверку TypeScript и production-сборку Vite.

### CI

GitHub Actions workflow находится в `.github/workflows/ci.yml`. Он запускает backend-тесты с PostgreSQL, проверки frontend (lint, tests, build), а также сборку Docker-образов backend и frontend.

## REST API

Базовый префикс API:

```text
/api/v1
```

Полная спецификация находится в `backend/docs/openapi.yaml` и доступна через Swagger UI после запуска приложения.

Основные группы API:

- `/auth` — регистрация, вход, обновление и отзыв токенов.
- `/users` — профиль и изменение пароля.
- `/projects` — проекты, участники, задачи, спринты, история и статистика.
- `/tasks` — задачи, комментарии, история и вложения.
- `/admin` — административное управление пользователями.

Защищённые endpoints требуют заголовок:

```http
Authorization: Bearer <access_token>
```

Успешные ответы API обычно возвращаются в оболочке `data`, ошибки — в формате:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Описание ошибки"
  }
}
```

## Загрузка файлов

К задаче можно прикрепить файл через endpoint `POST /api/v1/tasks/{id}/attachment` с типом содержимого `multipart/form-data` и полем `file`.

Поддерживаемые форматы: PNG, JPG/JPEG, PDF, TXT и DOCX. Максимальный размер задаётся переменной `MAX_UPLOAD_MB` (по умолчанию 5 МБ). Скачивание выполняется через авторизованный endpoint `GET /api/v1/tasks/{id}/attachment`; доступ проверяется на backend.

Файлы хранятся в Docker volume `uploads`, поэтому они сохраняются при обычном пересоздании контейнеров. Удаление томов удаляет и эти файлы.

## Структура репозитория

```text
taskflow/
├── backend/
│   ├── cmd/api/             # Точка входа backend
│   ├── docs/                # OpenAPI-спецификация
│   ├── internal/
│   │   ├── config/          # Конфигурация
│   │   ├── handler/         # HTTP-обработчики
│   │   ├── httpx/           # Формат HTTP-ответов и ошибок
│   │   ├── middleware/      # JWT и проверка ролей
│   │   ├── model/           # Модели данных
│   │   └── service/         # Бизнес-правила и вспомогательная логика
│   └── migrations/          # SQL-миграции
├── frontend/
│   └── src/                 # React-приложение
├── deploy/
│   └── nginx.conf           # Конфигурация Nginx
├── docs/                    # Требования и проектные материалы
├── .github/workflows/       # CI GitHub Actions
├── .env.example             # Пример переменных окружения
└── docker-compose.yml       # Оркестрация сервисов
```
