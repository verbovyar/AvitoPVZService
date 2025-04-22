# AvitoPVZService

## Проблема модели данных:
Исходя из нефункциональных требований к сервису (1000 RPS, 100ms latency), мною был сделан вывод, что выбор базы данных и их правильная организация будет ключевым моментом для того, чтобы достигнуть нужного уровня сервисного обслуживания. Мною анализировались как SQL решения (PostgresSQL), так и NoSQL решения(документоориентированная mongo, коробочное распределенное решение cassandra, scylla).

Было принято идти с PostgresSQL. Мною была выдвинута гипотеза, что классическая нормализованная модель данных, раскиданная по разным таблицам не даст необходимых результатов по мере роста записей в таблицах и нагрузок, ввиду большого количество джоинов и группировок на read нагрузке. Проверка этой гипотезы на локальной машине при минимальной нагрузке показала ее верность.

Вследствие чего я решил прибегнуть к денормализации насколько это возможно и перейти к максимально к плоской структуре данных. Было принято решение использовать всю мощь JSONB типов для того, чтобы хранить данные по приемке и товарам непосредственно в таблице ПВЗ, чтобы оптимизировать чтение. Такой необычный способ хранения позволил выдержать заданные нагрузки и максимально уменьшить latency.

![Схема базы данных](https://raw.githubusercontent.com/verbovyar/AvitoPVZService/main/images/dataBase.jpg)


## Сервис на Go для управления пунктами выдачи (ПВЗ), приёмом и передачей товаров через HTTP REST API и gRPC.

## Описание

AvitoPVZService — это микросервис, реализующий:
- Регистрацию и аутентификацию пользователей (с ролями `moderator` и `employee`) через JWT.
- Управление ПВЗ (создание, получение списка).
- Учёт приёмов (`receptions`) и товаров (`products`) внутри ПВЗ.
- Удаление последнего товара и закрытие последнего приёма.
- Экспорт списка ПВЗ через gRPC.
- Метрики Prometheus.


## Технологии

- Go 1.20+
- PostgreSQL + JSONB
- gRPC / Protocol Buffers (proto3)
- Viper (конфигурация)
- pgx (Postgres pool)
- jwt-go (JWT)
- Prometheus client

## Быстрый старт

### 1. Клонировать репозиторий
```bash
git clone https://github.com/verbovyar/AvitoPVZService.git
cd AvitoPVZService/Service
```

### 2. Устновка зависимости
```bash
go mod download
```

### 3. Настроить окружение

Создайте файл config/conf.env (пример уже есть в репозитории):

```env
PORT=:9000
CONNECTING_STRING=postgres://user:password@localhost:5432/AvitoDb
NETWORK_TYPE=tcp
GRPC_PORT=:3000
```

### 4. Применить миграции
```bash
goose -dir internal/migration postgres "$CONNECTING_STRING" up
```

### 5. Запустить сервис
```bash
go run cmd/app/main.go
```

- HTTP API на ${PORT} (по умолчанию :9000)

- gRPC API на ${GRPC_PORT} (по умолчанию :3000)

- Метрики Prometheus на :9000/metrics

## HTTP API

Все тела запросов и ответов — JSON

### Авторизация и пользователи
- POST /register

Регистрация:
```json
{
  "email": "user@example.com",
  "password": "secret",
  "role": "employee"
}
```

- POST /login

Логин:
```json
{
  "email": "user@example.com",
  "password": "secret"
}
```
— возвращает JWT токен.

- POST /dummyLogin

Генерация токена без БД (только для тестов):
```json
{ "role": "moderator" }
```
Все защищённые эндпоинты требуют заголовок
Authorization: Bearer <JWT_TOKEN>

### ПВЗ

- POST /pvz

  Создать ПВЗ (только для moderator):
```json
{ "city": "Москва" }
```
Разрешённые города: Москва, Санкт-Петербург, Казань

- GET /pvz?startDate=<RFC3339>&endDate=<RFC3339>&limit=10&offset=0

Получить список ПВЗ с приёмами и товарами (ролевая проверка: moderator, employee).

## Приёмы и товары

### POST /receptions

Открыть приём (только employee):
```json
{ "pvzId": "<UUID ПВЗ>" }
```

### POST /products

Добавить товар (только employee):

```json
{
  "pvzId": "<UUID ПВЗ>",
  "type": "электроника"
}
```

### POST /pvz/{pvzId}/delete_last_product

Удалить последний товар из открытого приёма (employee)

### POST /pvz/{pvzId}/close_last_reception

Закрыть последний приём (employee)

## gRPC API

Файл описания: Service/api/service.proto

- Сервис: pvz.v1.PVZService
- Метод 
```proto
rpc GetPVZList(GetPVZListRequest) returns (GetPVZListResponse);
```
Возвращает список ПВЗ за заданный период (по умолчанию — последние 24 ч)

## Структура проекта
```pgsql
Service/
├── cmd/app           – точка входа (main.go)
├── config            – конфигурация (Viper + `.env`)
├── internal/
│   ├── domain        – бизнес-модели
│   ├── handlers      – HTTP и gRPC хендлеры + middleware
│   ├── repositories/  
│   │   ├── interfaces – интерфейсы репозиториев
│   │   └── db         – реализация Postgres (JSONB)
│   ├── migration     – SQL-миграции (goose)
│   └── tokens        – JWT middleware
├── pkg               – обёртка над pgx pool
└── api               – proto файлы и автосгенерённый код
```

## Тестирование
```bash
go test ./internal/handlers
go test ./internal/repositories/db
go test ./internal/tokens
```
