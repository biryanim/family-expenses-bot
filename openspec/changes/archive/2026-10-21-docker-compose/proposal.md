# Proposal

## Why

Бэкенд, миграции и PostgreSQL уже запускаются через `docker compose`, но панель (задача 13) в compose не подключена: её образ собирается отдельно командой `docker build`. Нужен единый запуск всего сервиса одной командой `docker compose up --build`, как требует цель проекта, и проверка, что он работает с нуля.

## What Changes

- В `docker-compose.yml` добавляется сервис `web` (образ из `web/Dockerfile`, nginx), порт на хосте задаётся `WEB_PORT` (по умолчанию 3000). Он стартует после `app`, проксирует `/api/` на `app` по внутренней сети.
- `API_UPSTREAM` для `web` собирается из `HTTP_PORT`, чтобы смена порта API не ломала панель.
- В `.env.example` добавляется `WEB_PORT`; фронтенд-переменные не нужны (панель ходит в `/api` на своём origin).
- README: единый сценарий запуска (`cp .env.example .env` → `docker compose up --build`), адреса панели и API, как сбросить данные (`docker compose down -v`); удаляется заметка «подключение появится в задаче 14».
- Сквозная проверка на чистом томе: подъём всех сервисов, перезапуск с сохранением данных, бот и панель на одной БД.
- Уже существующие `Dockerfile` (multi-stage: `migrate`, `build`, финальный образ), `web/Dockerfile`, сервисы `postgres` (healthcheck, volume `pgdata`) и `migrate` не переписываются; правки в них допускаются только если проверка вскроет дефект.

## Capabilities

### New Capabilities
- `docker-deployment`: запуск всего сервиса (БД, миграции, бот с API, панель) одной командой `docker compose up`, конфигурация через `.env`, сохранность данных БД между перезапусками.

### Modified Capabilities

Нет. Поведение `app-bootstrap` («Миграции до старта приложения») не меняется: оно уже требует миграции до старта приложения, новый spec лишь добавляет к этому панель и тома.

## Impact

- Файлы: `docker-compose.yml`, `.env.example`, `README.md`; при необходимости `web/Dockerfile`, `web/nginx/default.conf.template`, `.dockerignore`.
- Код Go и фронтенда, REST API и контракт не меняются.
- Не входит в изменение: healthcheck приложения (distroless-образ без shell), HTTPS/внешний reverse proxy, авторизация панели, сборка образов в CI, публикация в registry.
