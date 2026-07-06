#  BookingGo

[![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8.svg)](https://go.dev/)
[![Docker](https://img.shields.io/badge/docker-ready-blue.svg)](https://www.docker.com/)
[![Coverage](https://img.shields.io/badge/coverage-85%25-brightgreen.svg)](./coverage.html)
[![Code Quality](https://img.shields.io/badge/code%20quality-A+-brightgreen.svg)]()

> **REST API для бронирований: уведомления, 2FA, Outbox-паттерн**

BookingGo — это REST API для управления бронированиями, построенный по принципам чистой архитектуры. 

В проекте реализованы уведомления для пользователя по различным каналам (email, SMS, in-app),
двухфакторная аутентификация (TOTP), кэширование через Redis и гарантированная доставка
событий через Outbox-паттерн с брокером NATS JetStream.
---

## Функционал

### ▸ Аутентификация и безопасность
- JWT-токен
- Двухфакторная аутентификация (TOTP)
- Подтверждение операций через пароль/2FA
- Rate limiting
- JWT Blacklist для logout

### ▸ Бронирования
- Создание, отмена, завершение бронирований
- Проверка доступности слотов с кэшированием
- Валидация пересечений
- Разделение ролей (client, staff, admin)

### ▸ Уведомления
- Многоканальные уведомления (email, SMS, in-app)
- Настраиваемые предпочтения пользователя
- Гарантированная доставка через Outbox-паттерн
- Фоновая обработка через воркеры

### ▸ Архитектура
- Clean Architecture (контроллеры → usecase → repository)
- Outbox-паттерн
- Кэширование через Redis
- Асинхронная обработка через NATS JetStream
- Структурированное логирование (slog)

---

## ️Технологический стек

| Компонент | Технология |
|-----------|----------|
| **Язык** | Go 1.22+ |
| **Фреймворк** | Gin |
| **База данных** | PostgreSQL 18 |
| **Кэш** | Redis 7 |
| **Очередь сообщений** | NATS JetStream |
| **ORM** | GORM |
| **Контейнеризация** | Docker + Docker Compose |
| **Тестирование** | go.uber.org/mock, testify |
| **Логирование** | slog |
| **Документация API** | Swagger/OpenAPI (swaggo) |

---

## Быстрый старт

### Требования

| Компонент | Минимальная версия | Проверить |
|-----------|-------------------|-----------|
| Go | 1.22+ | `go version` |
| Docker | 20.10+ | `docker --version` |
| Docker Compose | 2.0+ | `docker-compose --version` |

### Установка

```bash
# Клонируйте репозиторий
git clone https://github.com/UniversityProgramm/BookingGo.git
cd BookingGo

# Настройте переменные окружения
cp .env.example .env

# Запустите через Docker
docker-compose up --build -d

# Откройте Swagger UI
open http://localhost:8080/swagger/index.html
```