# Spec Delta

## MODIFIED Requirements

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

## ADDED Requirements

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
