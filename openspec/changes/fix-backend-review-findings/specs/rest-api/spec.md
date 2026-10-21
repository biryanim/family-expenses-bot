# Spec Delta

## MODIFIED Requirements

### Requirement: Список трат с фильтрами
`GET /expenses` SHALL возвращать массив трат всех членов семьи. Необязательные query-параметры: `user_id`, `from` (date-time, включительно), `to` (date-time, не включительно) по `spent_at`, `category`, `limit` (целое 1–500, по умолчанию 100), `offset` (целое от 0, по умолчанию 0). Фильтры комбинируются по «И», незаданный не ограничивает выборку. Порядок — от новых `spent_at` к старым. Ответ 200 SHALL содержать заголовок `X-Total-Count` с числом трат, подходящих под фильтры, без учёта `limit` и `offset`. Пустой результат — пустой массив, ответ 200.

#### Scenario: Фильтр по члену семьи и периоду
- **WHEN** запрашивается `GET /expenses?user_id=123&from=2026-10-01T00:00:00%2B03:00&to=2026-10-08T00:00:00%2B03:00`
- **THEN** ответ 200 содержит только траты члена семьи `123` с `spent_at` не раньше 1 октября и раньше 8 октября

#### Scenario: Лимит по умолчанию
- **WHEN** в базе 250 трат и запрашивается `GET /expenses` без `limit`
- **THEN** в теле 100 самых новых трат, `X-Total-Count` равен `250`

#### Scenario: Вторая страница
- **WHEN** в базе 250 трат и запрашивается `GET /expenses?limit=100&offset=100`
- **THEN** в теле траты с 101-й по 200-ю в порядке убывания `spent_at`, `X-Total-Count` равен `250`

#### Scenario: Смещение за пределами выборки
- **WHEN** в базе 50 трат и запрашивается `GET /expenses?offset=100`
- **THEN** ответ 200 с пустым массивом, `X-Total-Count` равен `50`

#### Scenario: Некорректный параметр
- **WHEN** `from` не является date-time или `user_id` не является целым числом
- **THEN** ответ — 400 в едином формате ошибки

#### Scenario: Нет подходящих трат
- **WHEN** под фильтры не попадает ни одна трата
- **THEN** ответ 200 с пустым массивом и `X-Total-Count` равным `0`

### Requirement: CORS для фронтенда
API SHALL отвечать на запросы и preflight-запросы (`OPTIONS`) из браузера с разрешённых источников (origin) из конфигурации, разрешая методы `GET`, `POST`, `PUT`, `DELETE` и заголовок `Content-Type`, и SHALL раскрывать браузеру заголовок `X-Total-Count` (`Access-Control-Expose-Headers`). Для источников вне списка заголовки CORS SHALL NOT выставляться.

#### Scenario: Preflight с разрешённого источника
- **WHEN** приходит `OPTIONS /api/expenses/1` с `Origin: http://localhost:5173`, `Access-Control-Request-Method: PUT`, и этот источник разрешён
- **THEN** ответ содержит `Access-Control-Allow-Origin: http://localhost:5173` и разрешает метод `PUT`

#### Scenario: Запрос с чужого источника
- **WHEN** приходит `GET /api/expenses` с `Origin: https://evil.example`, которого нет в списке
- **THEN** в ответе нет заголовка `Access-Control-Allow-Origin`

#### Scenario: Заголовок количества доступен браузеру
- **WHEN** приходит `GET /api/expenses` с разрешённого источника
- **THEN** ответ содержит `Access-Control-Expose-Headers` со значением `X-Total-Count`

## ADDED Requirements

### Requirement: Ошибки пагинации и комментария
Если `limit` вне диапазона 1–500 или `offset` отрицателен, API SHALL отвечать 400 с кодом `invalid_pagination`. Если `comment` длиннее 200 символов (рун) в `POST /expenses` или `PUT /expenses/{id}`, API SHALL отвечать 400 с кодом `invalid_comment`. Оба кода SHALL входить в перечень `ErrorCode` в `api/openapi.yaml`.

#### Scenario: Лимит больше максимума
- **WHEN** запрашивается `GET /expenses?limit=501`
- **THEN** ответ — 400, `code` = `invalid_pagination`

#### Scenario: Нулевой лимит
- **WHEN** запрашивается `GET /expenses?limit=0`
- **THEN** ответ — 400, `code` = `invalid_pagination`

#### Scenario: Отрицательное смещение
- **WHEN** запрашивается `GET /expenses?offset=-1`
- **THEN** ответ — 400, `code` = `invalid_pagination`

#### Scenario: Слишком длинный комментарий
- **WHEN** создаётся трата с комментарием в 201 символ
- **THEN** ответ — 400, `code` = `invalid_comment`, трата не создаётся

#### Scenario: Комментарий ровно 200 символов
- **WHEN** создаётся трата с комментарием в 200 кириллических символов
- **THEN** ответ — 201, комментарий сохранён целиком

### Requirement: Ограничение размера тела запроса
API SHALL отвергать запросы с телом больше 1 МиБ ответом 400 в едином формате ошибки и SHALL NOT читать избыточную часть тела в память.

#### Scenario: Огромное тело
- **WHEN** на `POST /expenses` приходит тело размером 5 МиБ
- **THEN** ответ — 400 с `code` = `invalid_request`, трата не создаётся
