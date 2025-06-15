# Банковский REST API на Golang

Этот проект представляет собой REST API для банковского сервиса, разработанный на Golang с использованием слоистой архитектуры. API предоставляет функциональность для управления банковскими счетами, картами, транзакциями, кредитами и включает аналитические возможности.

## Функциональные возможности

- Регистрация и аутентификация пользователей
- Управление банковскими счетами (создание, пополнение, снятие)
- Операции с картами (генерация, просмотр, оплата)
- Переводы между счетами
- Кредитные операции (оформление, график платежей)
- Аналитика финансовых операций
- Интеграция с ЦБ РФ для получения ключевой ставки
- Отправка email-уведомлений

## Технологии

- Go 1.23+
- PostgreSQL 17 с расширением pgcrypto
- Gorilla Mux для маршрутизации
- JWT для аутентификации
- Logrus для логирования
- Bcrypt, HMAC, PGP для шифрования
- Gomail для отправки email
- Etree для парсинга XML

## Установка и запуск

### Предварительные требования

- Go 1.23+
- Docker
- pgAdmin
- Git
- Visual Studio Code

### Шаги по установке

1. Клонируйте репозиторий:
```bash
git clone https://github.com/euchekavelo/bank-service.git
cd bank-service
```

2. Запустите контейнер с БД PostgreSQL:
```bash
docker compose -f ./docker-compose.yml up -d
```

3. Установите зависимости:
```bash
go mod download
```

4. Создайте базу данных, запустив скрипт из файла **migrations/001_init_schema.sql**.

4. Создайте или отредактируйте файл .env в корне проекта:
```
SERVER_PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=bank_service
DB_SSLMODE=disable

JWT_SECRET=your-secret-key
PGP_KEY=your-pgp-key
HMAC_KEY=your-hmac-key

SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=your-username
SMTP_PASSWORD=your-password
SMTP_FROM=test_bank@mail.ru

LOG_LEVEL=info
```

5. Соберите и запустите проект:
```bash
go build -o bank-service ./cmd/api
./bank-service
```

## API Endpoints

### Публичные эндпоинты

- `POST /register` - Регистрация нового пользователя
- `POST /login` - Аутентификация пользователя

### Защищенные эндпоинты (требуют JWT токен)

#### Счета
- `POST /accounts` - Создать новый счет
- `GET /accounts` - Получить все счета пользователя
- `GET /accounts/{id}` - Получить информацию о счете
- `POST /accounts/deposit` - Пополнить счет
- `POST /accounts/withdraw` - Снять средства со счета
- `GET /accounts/{id}/predict` - Прогноз баланса

#### Переводы
- `POST /transfer` - Перевод между счетами

#### Карты
- `POST /cards` - Выпустить новую карту
- `GET /cards` - Получить все карты пользователя
- `GET /cards/{id}` - Получить информацию о карте
- `PUT /cards/{id}/status` - Изменить статус карты
- `POST /cards/payment` - Оплата картой

#### Кредиты
- `POST /credits` - Оформить кредит
- `GET /credits` - Получить все кредиты пользователя
- `GET /credits/{id}` - Получить информацию о кредите
- `GET /credits/{id}/schedule` - Получить график платежей

#### Транзакции
- `GET /transactions` - Получить все транзакции пользователя
- `GET /accounts/{id}/transactions` - Получить транзакции по счету

#### Аналитика
- `GET /analytics/transactions` - Аналитика транзакций
- `GET /analytics/credits` - Аналитика кредитов

## Примеры использования

### Регистрация пользователя
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "Password123",
    "full_name": "Test User"
  }'
```

### Аутентификация
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Password123"
  }'
```

### Создание счета (требует JWT токен)
```bash
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "type": "DEBIT"
  }'
```

### Пополнение счета (требует JWT токен)
```bash
curl -X POST http://localhost:8080/accounts/deposit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "account_id": 1,
    "amount": 1000
  }'
```

## Тестирование

Для тестирования API рекомендуется использовать Postman или аналогичные инструменты.

## Структура проекта

```
bank-service/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── config/
│   ├── models/
│   ├── repository/
│   ├── service/
│   ├── handler/
│   ├── middleware/
│   └── scheduler/
├── pkg/
│   ├── logger/
│   ├── validator/
│   ├── encryption/
│   └── utils/
├── migrations/
├── .env
├── go.mod
├── go.sum
└── README.md
```
