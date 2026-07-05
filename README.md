# 🚗 BookingGo

[![Go Report Card](https://goreportcard.com/badge/github.com/твой-username/BookingGo)](https://goreportcard.com/report/github.com/твой-username/BookingGo)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8.svg)](https://go.dev/)
[![Docker](https://img.shields.io/badge/docker-ready-blue.svg)](https://www.docker.com/)
[![Coverage](https://img.shields.io/badge/coverage-85%25-brightgreen.svg)](./coverage.html)

> **Сервис бронирований с уведомлениями, 2FA и real-time обновлениями**

BookingGo — это REST API для управления бронированиями с поддержкой многоканальных уведомлений (email, SMS, in-app), двухфакторной аутентификации (TOTP), с гарантированной доставкой событий через Outbox-паттерн и поддержкой стороннего брокера сообщений.

---

## ✨ Возможности

### 🔐 Аутентификация и безопасность
- JWT-токены с refresh-механизмом
- Двухфакторная аутентификация (TOTP)
- Подтверждение операций через пароль/2FA
- Rate limiting
- JWT Blacklist для logout

### 📅 Бронирования
- Создание, отмена, завершение бронирований
- Проверка доступности слотов с кэшированием
- Валидация пересечений
- Разделение ролей (client, staff, admin)

### 🔔 Уведомления
- Многоканальные уведомления (email, SMS, in-app)
- Настраиваемые предпочтения пользователя
- Гарантированная доставка через Outbox-паттерн
- Фоновая обработка через воркеры

### ️ Архитектура
- Clean Architecture (контроллеры → usecase → repository)
- Outbox-паттерн для надёжности
- Кэширование через Redis
- Асинхронная обработка через NATS JetStream
- Структурированное логирование (slog)

---

## ️ Технологический стек

| Компонент | Технология |
|-----------|-----------|
| **Язык** | Go 1.22+ |
| **Фреймворк** | Gin |
| **База данных** | PostgreSQL 18 |
| **Кэш** | Redis 7 |
| **Очередь сообщений** | NATS JetStream |
| **ORM** | GORM |
| **Контейнеризация** | Docker + Docker Compose |
| **Тестирование** | go.uber.org/mock, testify |
| **Логирование** | slog (standard library) |

---

### Требования

| Компонент | Минимальная версия | Проверить |
|-----------|-------------------|-----------|
| Go | 1.22+ | `go version` |
| Docker | 20.10+ | `docker --version` |
| Docker Compose | 2.0+ | `docker-compose --version` |
| Make | 3.81+ | `make --version` |
