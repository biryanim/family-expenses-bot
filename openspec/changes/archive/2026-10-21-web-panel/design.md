# Design

## Context

- `web/` содержит только `.gitkeep`. В `.gitignore` уже есть `web/node_modules/` и `web/dist/`.
- REST API готов (`openspec/specs/rest-api`): `GET/POST /expenses`, `GET/PUT/DELETE /expenses/{id}`, `GET /members`, префикс `/api` (`servers.url` и `BaseURL: "/api"` в `internal/transport/http/new_router.go`). Ошибки приходят как `{"code", "message"}`. Для `invalid_request` поле `message` — английский текст ошибки разбора (`HandleRequestError`), для 500 — `internal server error`.
- Бэкенд уже умеет CORS (`CORS_ALLOWED_ORIGINS`, по умолчанию `http://localhost:5173`).
- `PUT /expenses/{id}` требует `amount`, `category`, `spent_at`. `comment` при отсутствии становится пустым. `user_id` не меняется.
- В `GET /expenses` нет пагинации, и для семьи она не нужна.
- Бот считает «сегодня/вчера» в часовом поясе контейнера (`TZ` в `.env`, пример `Europe/Moscow`).
- Локально Node 22, npm 10.
- `docker-compose.yml` пока без фронтенда. Сервис панели добавит задача 14.

## Goals / Non-Goals

**Goals:**
- Панель, которая на одном образе работает в любом окружении: адрес бэкенда задаётся при запуске контейнера.
- Типы API только из `api/openapi.yaml`, расхождение кода и контракта ловится `tsc`.
- Деньги без float и на вводе, и на выводе.
- Unit-тесты на чистую логику: деньги, границы периода, тексты ошибок, имя владельца.

**Non-Goals:**
- Подключение в `docker-compose.yml` (задача 14).
- Авторизация, роли, пагинация, сортировка на клиенте, графики и сводки.
- Смена владельца траты: API этого не позволяет, а менять контракт ради этого задача не требует.
- Тесты компонентов в DOM и e2e-тесты браузером. Критерии готовности проверяются руками на живом бэкенде (задача 16 — отдельная e2e-проверка).
- Изменения Go-кода и `api/openapi.yaml`.

## Decisions

### 1. Одно изменение на задачи 11–13

Каркас без списка не показывает данных, а правка без списка не имеет точки входа. Если делать три изменения, первые два пришлось бы архивировать с заглушками. Спецификация одна — `web-panel`. В `tasks.md` группы задач идут в порядке 11 → 12 → 13, так что их можно сдавать по очереди.

### 2. Адрес API: относительный `/api` и прокси, а не абсолютный URL с CORS

- В коде клиента базовый адрес — `import.meta.env.VITE_API_BASE_URL`, по умолчанию `/api`. Значение `/api` совпадает с `servers.url` из спецификации.
- **Dev**: `vite.config.ts` проксирует `/api` на `VITE_API_PROXY_TARGET`. Значение лежит в закоммиченном `web/.env.development` (`http://localhost:8080`) и переопределяется через `web/.env.development.local` или окружение.
- **Образ**: официальный образ nginx при старте подставляет переменные окружения в `/etc/nginx/templates/*.template` (envsubst только для заданных переменных, поэтому `$uri` и другие nginx-переменные не трогаются). `default.conf.template` содержит `location /api/ { proxy_pass ${API_UPSTREAM}; }`. Без URI в `proxy_pass` путь `/api/...` уходит на бэкенд как есть. В Dockerfile задано `ENV API_UPSTREAM=http://app:8080` как значение по умолчанию под имя сервиса в compose, его можно переопределить при запуске.

Альтернативы:
1. **Абсолютный `VITE_API_BASE_URL` при сборке + CORS.** Адрес вшивается в бандл, и на каждое окружение нужна своя сборка. К тому же браузеру нужен адрес, доступный с хоста, а не из сети compose. Остаётся возможным через ту же переменную, если кто-то захочет без прокси.
2. **`config.js` с `window.__CONFIG__`, генерируемый при старте контейнера.** Даёт runtime-конфиг без прокси, но требует CORS, лишнего скрипта и глобальной переменной. Прокси проще и убирает CORS вообще.

### 3. Клиент: `openapi-fetch` поверх сгенерированных типов

`openapi-typescript` генерирует только типы (`paths`, `components`), клиента он не даёт. `openapi-fetch` — клиент того же проекта, лёгкий и без кодогенерации. Он типизирует путь, параметры, тело и ответ прямо из `paths`, так что поле, которого нет в спецификации, не скомпилируется. Это новая зависимость сверх AGENTS.md, и она добавляется осознанно.

Альтернативы:
- **Свой `fetch`-обёртка с дженериками над `paths`.** Это переписывание `openapi-fetch` своими руками, причём типы извлечения параметров и ответов сложные и хрупкие.
- **orval / openapi-generator.** Другой генератор, конфликтует с выбранным в AGENTS.md `openapi-typescript` и генерирует много кода.
- **`openapi-react-query`.** Ещё один слой. Хуков поверх TanStack Query нужно пять, их проще написать явно.

Сгенерированный файл — `web/src/api/schema.gen.ts`. Скрипт: `"generate:api": "openapi-typescript ../api/openapi.yaml -o src/api/schema.gen.ts"`. Сборка (`tsc -b && vite build`) генератор не вызывает, поэтому Docker-образу не нужен `api/`, и контекст сборки — только `web/`. `make generate-web` = `npm --prefix web run generate:api`. `make generate` не трогаем: спецификация `api-codegen` говорит, что он равен `go generate ./...`, и Go-генерации не должна требоваться Node.

### 4. Слой запросов: TanStack Query

- Ключи: `['members']`, `['expenses', filters]`, где `filters` — уже нормализованные query-параметры API (`user_id`, `from`, `to`, `category`). Смена фильтра — новый ключ и новый запрос.
- Мутации `createExpense`, `updateExpense`, `deleteExpense` в `onSuccess` вызывают `invalidateQueries({ queryKey: ['expenses'] })`. Активный запрос с текущими фильтрами перезапрашивается, а фильтры живут вне Query и не сбрасываются.
- При 404 на правке или удалении список тоже инвалидируется: трату удалили в боте, и показывать её дальше нельзя.
- Актуальность изменений из бота обеспечивают `refetchOnWindowFocus` (в TanStack Query включён по умолчанию, оставляем) и кнопка «Обновить». Поллинг не делаем: для одного главы семьи он даёт только лишнюю нагрузку.
- `retry`: для запросов списка — 1, для мутаций — 0, чтобы не создавать дубликаты траты при повторе `POST`.
- `queryFn` бросает `ApiError` (см. решение 7), поэтому компоненты работают с одним типом ошибки.

### 5. Фильтры в URL без роутера

Страница одна, так что `react-router` не нужен. Состояние фильтров читается из `URLSearchParams` при старте и пишется через `history.replaceState`. В URL хранятся значения формы, а не API: `member`, `from` и `to` как даты `YYYY-MM-DD`, `category`. Некорректные значения из URL игнорируются.

### 6. Деньги и даты

- **Вывод** (`formatAmount`): `rub = Math.trunc(k / 100)`, `kop = k % 100`. Рубли группируются через `Intl.NumberFormat('ru-RU')` — это целое число, не float. Копейки дополняются до двух знаков, получается `1 200,00 ₽`. Разделитель разрядов — неразрывный пробел, как отдаёт `Intl`.
- **Ввод** (`parseAmount`): удаляются все пробельные символы, затем проверка `^\d+([.,]\d{1,2})?$`. Результат: `int(целая) * 100 + int(дробная, дополненная до 2)`. Разбор через `Number` только целых частей, без умножения дробного числа. Значение `0` и всё, что больше `Number.MAX_SAFE_INTEGER` копеек, отклоняется. Ошибка возвращается как значение, без исключения.
- **Дата траты в форме** — `<input type="datetime-local">`. В форме правки сохраняется время траты из бота, а не только день. В API уходит RFC 3339 со смещением часового пояса браузера на эту дату (`2026-10-21T14:35:00+03:00`), смещение считается через `getTimezoneOffset()` конкретной даты, поэтому учитывается переход на летнее время.
- **Период**: `from` — локальная полночь даты начала, `to` — локальная полночь дня после даты конца, обе в том же формате со смещением. Сравнение «начало позже конца» — по строкам `YYYY-MM-DD`.
- **Вывод даты**: `Intl.DateTimeFormat('ru-RU', { dateStyle: 'short', timeStyle: 'short' })` в поясе браузера.
- `spent_at` при создании по умолчанию — текущий момент в поле формы. Пустое поле не отправляется, тогда сервер сам ставит текущий момент.

### 7. Ошибки

`ApiError { status, code?, message }`:
- Если тело — объект с `code` и `message`, `code` переводится в русский текст: `invalid_amount` — «Некорректная сумма», `invalid_category` — «Некорректная категория», `invalid_date` — «Некорректная дата», `unknown_member` — «Такого члена семьи нет», `not_found` — «Трата не найдена», `internal` — «Ошибка сервера, попробуйте позже», `invalid_request` — «Некорректный запрос». Неизвестный `code` — `message` из ответа.
- `fetch` упал (сеть, прокси без бэкенда отдал 502 не в нашем формате) — «Сервер недоступен».
- Для `invalid_request` английский `message` сервера не показываем: он технический.

Ошибка мутации показывается внутри формы, и форма не закрывается. Ошибка списка — блок с кнопкой «Повторить».

### 8. UI без библиотек

Формы — состояние React и функции валидации из `lib/`. Без `react-hook-form` и `zod`: полей пять, правила простые, и валидацию всё равно пишем сами. Модалка формы — нативный `<dialog>`. Подтверждение удаления — `window.confirm`: синхронно, доступно с клавиатуры и не требует отдельного компонента. Стили — один `styles.css`, без UI-кита.

### 9. Структура `web/`

```
web/
  package.json, package-lock.json, tsconfig.json, tsconfig.node.json, vite.config.ts, index.html
  .env.development        VITE_API_PROXY_TARGET=http://localhost:8080
  .env.example            описание VITE_API_BASE_URL и VITE_API_PROXY_TARGET
  Dockerfile, .dockerignore
  nginx/default.conf.template
  src/
    main.tsx              QueryClientProvider, рендер App
    App.tsx               страница: фильтры, кнопки, таблица, модалка формы
    styles.css
    api/schema.gen.ts     сгенерировано
    api/client.ts         createClient<paths>({ baseUrl })
    api/api-error.ts      ApiError и разбор ответа
    api/queries.ts        useMembers, useExpenses, useCreateExpense, useUpdateExpense, useDeleteExpense
    lib/money.ts          formatAmount, parseAmount
    lib/dates.ts          toRfc3339Local, periodToRange, formatDateTime, toDateTimeLocalValue
    lib/categories.ts     коды и русские названия
    lib/members.ts        memberDisplayName
    lib/filters.ts        чтение и запись фильтров в URL
    lib/*.test.ts         vitest
    components/Filters.tsx, ExpenseTable.tsx, ExpenseForm.tsx, ErrorMessage.tsx
```

Правило «комментариев в коде нет» распространяется и на TS. Правила Go-раскладки (метод на файл, `deps.go`) к фронтенду не применяются: в TS группировка по модулю (`lib/money.ts`) читается лучше.

### 10. Тесты: vitest на чистую логику

`vitest` добавляется в dev-зависимости, потому что он работает на конфиге Vite и не требует отдельной настройки транспиляции. Тестируются `parseAmount`, `formatAmount`, `periodToRange` (с фиксированным `TZ` в скрипте `test`: `TZ=Europe/Moscow vitest run`), `toRfc3339Local`, разбор `ApiError`, `memberDisplayName`, чтение фильтров из URL. Компоненты не тестируем (см. Non-Goals). `npm test` и `npm run build` — проверки перед коммитом для `web/`.

### 11. Dockerfile

```
FROM node:22-alpine AS build      npm ci → npm run build
FROM nginx:<stable>-alpine        COPY dist → /usr/share/nginx/html, nginx/default.conf.template → /etc/nginx/templates/
ENV API_UPSTREAM=http://app:8080
```

Точный тег stable-nginx фиксируется при реализации по актуальному на тот момент. `default.conf.template`: `listen 80`, `location /api/` → прокси, `location /` → `try_files $uri /index.html`, долгий кеш для `/assets/` (имена с хешем), `no-cache` для `index.html`.

## Risks / Trade-offs

- [Если nginx стартует раньше, чем в Docker DNS появится `app`, он падает с «host not found in upstream»] → В задаче 14 сервис панели получает `depends_on: app`. Альтернатива с `resolver 127.0.0.11` и переменной в `proxy_pass` сложнее, и при одном инстансе бэкенда она не нужна.
- [Часовой пояс браузера и `TZ` бэкенда могут различаться, тогда «вчера» в боте и день в фильтре панели не совпадут на границе суток] → Семья в одном поясе. Решение явное: панель живёт в поясе браузера. В README это указано.
- [int64 копеек больше 2^53 в JSON теряет точность в JS] → Нереалистично для семейных трат (~90 трлн ₽). `parseAmount` не пропускает такие значения.
- [`window.confirm` выглядит по-системному] → Сознательный обмен на простоту и доступность.
- [Глава семьи не видит трату, которую бот добавил, пока не вернётся на вкладку или не нажмёт «Обновить»] → Это поведение из спеки. Поллинг можно добавить позже, не меняя остальное.
- [Сгенерированный `schema.gen.ts` может разойтись со спецификацией, если забыть `make generate-web`] → Генерация детерминирована. Проверка «перегенерировать и `git diff --exit-code`» описана в задачах и выполняется перед коммитом при изменении спецификации.

## Migration Plan

Новый каталог, существующее поведение не меняется. Откат — удалить `web/` и цель `generate-web`.
