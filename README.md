# Http API Service

Микросервис для управления длительными I/O bound операциями через HTTP API с хранением данных в памяти.

## 📋 Основные возможности

- Полный цикл управления задачами (создание/отслеживание/удаление)
- Фоновое выполнение длительных операций (3-5 минут)
- Потокобезопасное in-memory хранилище
- Детальная информация о статусе выполнения задач

## 🛠️ HTTP обработчики

### Основные endpoint'ы:

`POST /tasks`  
Создание новой задачи  
*Параметры:* description (описание задачи)  
*Возвращает:* ID и начальный статус задачи

`GET /tasks`  
Получение списка всех задач  
*Поддержка:* пагинация (page, page_size)  
*Возвращает:* массив задач с метаданными

`GET /tasks/{id}`  
Получение информации о конкретной задаче  
*Возвращает:* полный статус + результаты выполнения

`PUT /tasks/{id}`  
Обновление описания задачи  
*Параметры:* новое описание  
*Возвращает:* обновленную задачу

`POST /tasks/{id}/cancel`  
Отмена выполнения задачи  
*Работает только для задач в статусе pending/processing*

`DELETE /tasks/{id}`  
Удаление задачи из системы  
*Возвращает:* 204 No Content при успехе

## 🚀 Запуск сервиса

```bash
Инструкции:
1. Клонировать репозиторий
2. Перейти в директорию проекта
3. Выполнить `go run main.go`

Сервис будет доступен на `http://localhost:8080`

## Структура проекта
http-api/
├── internal/
│   ├── handlers/      # HTTP обработчики
│   ├── models/        # Модели данных
│   ├── services/      # Бизнес-логика
│   └── storage/       # In-memory хранилище
├── main.go            # Точка входа
├── go.mod             # Модули Go
└── README.md          # Этот файл
````
### 🧪 Тестирование

Проект включает полный набор автоматических тестов, обеспечивающих надежность работы API:

#### Ключевые тестовые сценарии:
- **Позитивные тесты**:
  - Создание задачи с валидными данными (201 Created)
  - Получение списка задач (200 OK)
  - Получение конкретной задачи (200 OK)
  - Обновление описания задачи (200 OK)
  - Корректная отмена задачи (200 OK)
  - Удаление задачи (204 No Content)

- **Негативные тесты**:
  - Создание задачи с пустым описанием (400 Bad Request)
  - Невалидный JSON в запросе (400 Bad Request)
  - Запрос несуществующей задачи (404 Not Found)
  - Попытка отмены уже завершенной задачи (400 Bad Request)
  - Использование неподдерживаемых методов (405 Method Not Allowed)

- **Граничные случаи**:
  - Работа с пустым списком задач
  - Многократное обновление задачи
  - Попытка отмены уже отмененной задачи

#### Как запустить:
```bash
# Все тесты с подробным выводом
go test -v ./...
```