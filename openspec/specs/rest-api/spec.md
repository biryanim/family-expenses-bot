# rest-api Specification

## Purpose

Определяет контракт REST API веб-панели главы семьи: какие ресурсы и операции доступны, в каком формате передаются траты, члены семьи и ошибки. Источник истины — `api/openapi.yaml`.

## Requirements

### Requirement: Контракт описан в спецификации OpenAPI
REST API SHALL быть описан в `api/openapi.yaml` в формате OpenAPI 3.0.3. Каждая операция, параметр, тело запроса, ответ и схема API SHALL присутствовать в этом файле. Операции, которых нет в спецификации, сервер не предоставляет.

#### Scenario: Состав операций
- **WHEN** читается `api/openapi.yaml`
- **THEN** в нём описаны ровно операции `GET /expenses`, `POST /expenses`, `GET /expenses/{id}`, `PUT /expenses/{id}`, `DELETE /expenses/{id}` и `GET /members`, у каждой есть уникальный `operationId`

### Requirement: Схема траты
Трата в ответах API SHALL передаваться объектом с полями: `id` (int64), `user_id` (int64, Telegram ID члена семьи), `amount` (int64, копейки), `category`, `spent_at` (date-time), `comment` (строка, может быть пустой), `created_at` (date-time), `updated_at` (date-time или `null`, если трата не изменялась). Все поля, кроме `updated_at`, всегда присутствуют.

#### Scenario: Трата без изменений
- **WHEN** API возвращает трату, которая ни разу не изменялась
- **THEN** `updated_at` равно `null`, а `comment` — строка, пустая, если комментария нет

### Requirement: Сумма в копейках
Поле `amount` SHALL быть целым 64-битным числом копеек (`integer`, `format: int64`) в запросах и ответах. Минимальное допустимое значение в запросах — `1`. Дробные значения не допускаются.

#### Scenario: Сумма 1200 рублей
- **WHEN** клиент создаёт трату на 1200 ₽
- **THEN** он передаёт `amount` = `120000`, и в ответе `amount` = `120000`

#### Scenario: Дробная или неположительная сумма
- **WHEN** клиент передаёт `amount` = `350.5`, `0` или `-100`
- **THEN** запрос не соответствует контракту, ответ — 400

### Requirement: Фиксированный набор категорий
Категория SHALL быть перечислением ровно из кодов `food` (еда), `transport` (транспорт), `housing` (жильё), `entertainment` (развлечения), `health` (здоровье), `shopping` (покупки), `other` (другое). Одна и та же схема используется в теле запросов, ответах и фильтре списка.

#### Scenario: Неизвестная категория
- **WHEN** клиент передаёт категорию `travel` или `еда` в теле запроса или в фильтре
- **THEN** запрос не соответствует контракту, ответ — 400

### Requirement: Список трат с фильтрами
`GET /expenses` SHALL возвращать массив трат всех членов семьи. Необязательные query-параметры: `user_id`, `from` (date-time, включительно), `to` (date-time, не включительно) по `spent_at`, `category`. Параметры комбинируются по «И», незаданный не ограничивает выборку. Порядок — от новых `spent_at` к старым. Пустой результат — пустой массив, ответ 200.

#### Scenario: Фильтр по члену семьи и периоду
- **WHEN** запрашивается `GET /expenses?user_id=123&from=2026-10-01T00:00:00%2B03:00&to=2026-10-08T00:00:00%2B03:00`
- **THEN** ответ 200 содержит только траты члена семьи `123` с `spent_at` не раньше 1 октября и раньше 8 октября

#### Scenario: Некорректный параметр
- **WHEN** `from` не является date-time или `user_id` не является целым числом
- **THEN** ответ — 400 в едином формате ошибки

#### Scenario: Нет подходящих трат
- **WHEN** под фильтры не попадает ни одна трата
- **THEN** ответ 200 с пустым массивом

### Requirement: Создание траты
`POST /expenses` SHALL принимать `user_id`, `amount`, `category` (обязательные), `spent_at` (date-time) и `comment` (необязательные) и отвечать 201 с созданной тратой. Если `spent_at` не передан, трата датируется текущим моментом. Неизвестный `user_id` — ответ 400.

#### Scenario: Успешное создание
- **WHEN** отправляется `{"user_id": 123, "amount": 35000, "category": "food", "comment": "кофе"}`
- **THEN** ответ 201 с тратой, у которой назначен `id`, `spent_at` — момент создания, `updated_at` = `null`

#### Scenario: Неизвестный член семьи
- **WHEN** `user_id` не принадлежит ни одному члену семьи
- **THEN** ответ 400 в едином формате ошибки

### Requirement: Получение траты
`GET /expenses/{id}` SHALL отвечать 200 с тратой. Если траты нет — 404.

#### Scenario: Трата не существует
- **WHEN** запрашивается `GET /expenses/999` для несуществующего `id`
- **THEN** ответ 404 в едином формате ошибки

### Requirement: Изменение траты
`PUT /expenses/{id}` SHALL полностью заменять изменяемые поля траты: `amount`, `category`, `spent_at` (обязательные) и `comment` (необязательный; не передан — становится пустым). Владелец траты (`user_id`) через API не меняется. Ответ — 200 с изменённой тратой, 404 если траты нет.

#### Scenario: Успешное изменение
- **WHEN** отправляется `PUT /expenses/1` с `{"amount": 150000, "category": "transport", "spent_at": "2026-10-17T09:00:00+03:00"}`
- **THEN** ответ 200 с тратой, у которой новые значения, пустой `comment` и заполненный `updated_at`

#### Scenario: Изменение несуществующей траты
- **WHEN** отправляется `PUT` для несуществующего `id`
- **THEN** ответ 404

### Requirement: Удаление траты
`DELETE /expenses/{id}` SHALL удалять трату и отвечать 204 без тела. Если траты нет — 404.

#### Scenario: Повторное удаление
- **WHEN** одна и та же трата удаляется второй раз
- **THEN** ответ 404

### Requirement: Список членов семьи
`GET /members` SHALL возвращать массив всех членов семьи, включая тех, у кого нет трат, по возрастанию Telegram ID. У каждого: `telegram_id` (int64), `username`, `first_name`, `last_name` — строки, пустые, если в Telegram значения нет.

#### Scenario: Член семьи без username
- **WHEN** у члена семьи в Telegram нет username и фамилии
- **THEN** он присутствует в ответе с `username` = `""` и `last_name` = `""`

### Requirement: Единый формат ошибок
Каждый ответ с кодом 400, 404 или 500 SHALL иметь тело `{"code": <строка>, "message": <строка>}`, где `code` — машиночитаемый код из перечисления `invalid_request`, `invalid_amount`, `invalid_category`, `invalid_date`, `unknown_member`, `not_found`, `internal`, а `message` — описание для человека. Ответ 500 не раскрывает внутренние детали ошибки. Формат действует и для ошибок разбора запроса: некорректных параметров пути и query, невалидного JSON в теле.

#### Scenario: Внутренняя ошибка
- **WHEN** при обработке запроса падает база данных
- **THEN** ответ 500 с `code` = `internal` и общим сообщением без текста исходной ошибки

#### Scenario: Ошибка «не найдено»
- **WHEN** запрашивается несуществующая трата
- **THEN** ответ 404 с `code` = `not_found`

#### Scenario: Некорректный id в пути
- **WHEN** запрашивается `GET`, `PUT` или `DELETE /expenses/abc`
- **THEN** ответ 400 с `code` = `invalid_request`

#### Scenario: Невалидный JSON в теле
- **WHEN** `POST /expenses` или `PUT /expenses/{id}` отправляется с телом, которое не является JSON, или с `amount` = `350.5`
- **THEN** ответ 400 с `Content-Type: application/json` и `code` = `invalid_request`

### Requirement: Доменные ошибки в кодах ответа
Нарушения бизнес-правил SHALL отдаваться статусом 400 с кодом: недопустимая сумма — `invalid_amount`, недопустимая категория — `invalid_category`, некорректная дата — `invalid_date`, неизвестный член семьи — `unknown_member`. Отсутствующая трата SHALL отдаваться статусом 404 с `not_found`. Любая другая ошибка — 500 с `internal`.

#### Scenario: Нулевая сумма при создании
- **WHEN** отправляется `POST /expenses` с `amount` = `0`
- **THEN** ответ 400 с `code` = `invalid_amount`

#### Scenario: Неизвестная категория в теле
- **WHEN** отправляется `PUT /expenses/1` с `category` = `travel`
- **THEN** ответ 400 с `code` = `invalid_category`

#### Scenario: Неизвестная категория в фильтре
- **WHEN** запрашивается `GET /expenses?category=travel`
- **THEN** ответ 400 с `code` = `invalid_category`

#### Scenario: Изменение без даты
- **WHEN** отправляется `PUT /expenses/1` без `spent_at`
- **THEN** ответ 400 с `code` = `invalid_date`

#### Scenario: Начало периода позже конца
- **WHEN** запрашивается `GET /expenses` с `from` позже `to`
- **THEN** ответ 400 с `code` = `invalid_date`

#### Scenario: Неизвестный член семьи
- **WHEN** отправляется `POST /expenses` с `user_id`, которого нет среди членов семьи
- **THEN** ответ 400 с `code` = `unknown_member`

#### Scenario: Удаление несуществующей траты
- **WHEN** отправляется `DELETE /expenses/999` для несуществующего `id`
- **THEN** ответ 404 с `code` = `not_found`

### Requirement: Префикс пути API
Все операции API SHALL быть доступны по путям из спецификации с префиксом `/api` (значение `servers.url`). Запрос на путь, которого нет в спецификации, SHALL получать 404.

#### Scenario: Список трат по префиксу
- **WHEN** запрашивается `GET /api/expenses`
- **THEN** ответ 200 со списком трат

#### Scenario: Путь без префикса
- **WHEN** запрашивается `GET /expenses`
- **THEN** ответ 404

### Requirement: CORS для фронтенда
API SHALL отвечать на запросы и preflight-запросы (`OPTIONS`) из браузера с разрешённых источников (origin) из конфигурации, разрешая методы `GET`, `POST`, `PUT`, `DELETE` и заголовок `Content-Type`. Для источников вне списка заголовки CORS SHALL NOT выставляться.

#### Scenario: Preflight с разрешённого источника
- **WHEN** приходит `OPTIONS /api/expenses/1` с `Origin: http://localhost:5173`, `Access-Control-Request-Method: PUT`, и этот источник разрешён
- **THEN** ответ содержит `Access-Control-Allow-Origin: http://localhost:5173` и разрешает метод `PUT`

#### Scenario: Запрос с чужого источника
- **WHEN** приходит `GET /api/expenses` с `Origin: https://evil.example`, которого нет в списке
- **THEN** в ответе нет заголовка `Access-Control-Allow-Origin`
