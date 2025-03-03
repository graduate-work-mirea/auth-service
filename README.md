# Auth Service MVP

Микросервис аутентификации и авторизации для проекта "Интеллектуальная система оценки спроса на продукт". Этот сервис обеспечивает базовую аутентификацию и авторизацию через HTTP и gRPC API.

## Функциональность

- **Регистрация пользователя** (`POST /auth/register`)
- **Аутентификация** (`POST /auth/login`)
- **Валидация токена** (gRPC метод `ValidateToken`)

## Стек технологий

- **Язык**: Go 1.21+
- **HTTP-фреймворк**: Gin
- **gRPC**: google.golang.org/grpc
- **База данных**: PostgreSQL с sqlx
- **JWT**: github.com/golang-jwt/jwt/v5
- **Хеширование паролей**: golang.org/x/crypto/bcrypt

## Запуск сервиса

### С использованием Docker Compose

```bash
# Клонировать репозиторий
git clone https://github.com/your-username/auth-service.git
cd auth-service

# Запустить сервис и базу данных
docker-compose up --build
```

Сервис будет доступен по адресам:
- HTTP API: http://localhost:8080
- gRPC API: localhost:50051

## Тестирование API

### Регистрация пользователя

```bash
curl -X POST http://localhost:8080/auth/register -d '{"email": "test@example.com", "password": "secret"}'
```

Успешный ответ:
```json
{"user_id": "uuid-value"}
```

### Вход пользователя

```bash
curl -X POST http://localhost:8080/auth/login -d '{"email": "test@example.com", "password": "secret"}'
```

Успешный ответ:
```json
{"access_token": "jwt-token-value"}
```

### Проверка токена (gRPC)

Можно использовать инструменты вроде [grpcurl](https://github.com/fullstorydev/grpcurl) или [BloomRPC](https://github.com/uw-labs/bloomrpc):

```bash
grpcurl -plaintext -d '{"token": "your-jwt-token"}' localhost:50051 auth.AuthService/ValidateToken
```

## Переменные окружения

- `POSTGRES_DSN`: строка подключения к PostgreSQL (по умолчанию: "postgres://user:password@postgres:5432/auth_db?sslmode=disable")
- `JWT_SECRET`: секретный ключ для подписи JWT-токенов (по умолчанию: "your_secret_key_here", для продакшена обязательно изменить)

## Структура проекта

```
/auth-service
├── cmd/
│   └── main.go              # Точка входа
├── internal/
│   ├── handlers/            # HTTP-обработчики
│   │   └── auth.go
│   ├── grpc/                # gRPC-сервер
│   │   └── server.go
│   ├── models/              # Модели данных
│   │   └── user.go
│   ├── repository/          # Работа с БД
│   │   └── user_repository.go
│   └── utils/               # Утилиты (JWT)
│       └── jwt.go
├── proto/
│   ├── auth.proto           # Protobuf-схема
│   ├── auth.pb.go           # Сгенерированный код
│   └── auth_grpc.pb.go      # Сгенерированный gRPC код
├── Dockerfile               # Контейнеризация
└── docker-compose.yml       # PostgreSQL + сервис