# Tasks

## 1. Сервис `web` в compose

- [x] 1.1 Добавить в `docker-compose.yml` сервис `web` (build `./web`, `API_UPSTREAM: http://app:${HTTP_PORT:-8080}`, порт `${WEB_PORT:-3000}:80`, `depends_on: app` с `service_started`, `restart: unless-stopped`); проверить `docker compose config` без ошибок
- [x] 1.2 Добавить `WEB_PORT=3000` с комментарием в `.env.example`; проверить, что все переменные, на которые ссылается compose, описаны в `.env.example`

## 2. Документация

- [x] 2.1 Обновить README: сценарий `cp .env.example .env` → `docker compose up --build`, адреса панели (`http://localhost:3000`) и API, `docker compose down -v` для сброса данных, предупреждение про URL-кодирование пароля в `DATABASE_DSN`; убрать заметку «появится в задаче 14» и проверить, что команды из README выполняются как написано

## 3. Сквозная проверка с нуля

- [x] 3.1 На чистом состоянии (`docker compose down -v`, `docker image prune` для образов проекта) выполнить `docker compose up --build`; убедиться, что `postgres` healthy, `migrate` завершился с кодом 0, `app` и `web` работают
- [x] 3.2 Проверить через `curl http://localhost:3000/api/expenses` (через nginx) и `curl http://localhost:8080/api/expenses` (напрямую): оба отвечают 200; создать трату через `POST /api/expenses` и увидеть её в списке
- [x] 3.3 Проверить сохранность данных: `docker compose down`, `docker compose up`, трата из 3.2 на месте, повторный `migrate` завершается с кодом 0
- [x] 3.4 Проверить `HTTP_PORT=9090` в `.env`: панель отвечает через `/api/`
- [x] 3.5 Проверить отсутствие секретов в образах: `docker history --no-trunc` для `app`, `migrate`, `web` и поиск значений из `.env`
- [x] 3.6 Проверить без `.env`: `docker compose up` падает с понятной ошибкой
- [x] 3.7 Проверить, действует ли `TZ=Europe/Moscow` в контейнере `app` (в distroless нет tzdata); при расхождении исправить отдельной правкой и описать в README
- [x] 3.8 Вручную с реальными токенами: трата из бота видна в панели (`http://localhost:3000`), трата из панели видна боту
