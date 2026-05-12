# task-manager

Распределённая система запуска задач в Docker-контейнерах.

## Архитектура

Два сервиса общаются через Kafka:

```
CLI (tm run)
  └─[Unix socket]─▶ Planner daemon
                      └─[Kafka: tasks]─▶ Worker
                      ◀─[Kafka: results]─┘
```

**Planner** — принимает задачи от CLI, хранит состояние в SQLite, загружает входные файлы в S3, публикует задачи в Kafka, получает результаты.

**Worker** — читает задачи из Kafka, скачивает файлы из S3, запускает команду в Docker-контейнере, загружает выходные файлы обратно в S3, публикует результат.

**Инфраструктура** (docker-compose): Kafka + Zookeeper, MinIO (S3), Kafka UI, SQLite Web UI.

## Запуск

```bash
docker-compose up
```

UI для инспекции:
- `localhost:8080` — Kafka UI
- `localhost:8081` — SQLite Web
- `localhost:9000` — MinIO

## CLI

```bash
tm run "<команда>" [флаги]
```

| Флаг | Описание |
|------|----------|
| `--name` | Имя задачи |
| `--image` | Docker-образ (по умолчанию `bash:latest`) |
| `--in key=path` | Входной файл |
| `--out key=path` | Выходной файл |

Пример:
```bash
tm run "cat input.txt > output.txt" --in input=./data.txt --out result=./output.txt
```

## Жизненный цикл задачи

1. CLI отправляет запрос через Unix socket (`/var/run/task-manager/planner.sock`)
2. Planner создаёт запись в SQLite со статусом `running`, загружает входные файлы в S3
3. Planner публикует задачу в Kafka
4. Worker скачивает файлы из S3, запускает команду в контейнере
5. Worker загружает выходные файлы в S3, публикует результат в Kafka
6. Planner обновляет статус в SQLite (`succeeded` / `failed`), скачивает выходные файлы

## Конфигурация

Единый `config.json` (генерируется из `config.template.json` при старте):

```json
{
  "log_level": "info",
  "storage": {
    "sqlite_file": "/var/lib/task-manager/planner.db",
    "s3": { "endpoint": "...", "bucket": "...", "access_key_id": "...", "secret_access_key": "..." }
  },
  "messaging": {
    "brokers": ["kafka:9092"],
    "topics": { "tasks": "tasks", "results": "results" }
  },
  "shutdown_timeout": "10s"
}
```

Worker дополнительно имеет секцию `execution`:
```json
{
  "execution": {
    "work_dir": "/tasks",
    "default_image": "bash:latest"
  }
}
```

## Что реализовано

- Запуск одиночной задачи через CLI
- Передача файлов: локальная ФС ↔ S3 ↔ Worker
- Выполнение произвольной команды в Docker (pull образа при необходимости)
- Персистентность состояния задач (SQLite)
- Graceful shutdown обоих сервисов
- IPC через Unix socket (HTTP over socket)

## Что не реализовано

- Команды `list`, `show`, `logs`, `cancel`
- Группы задач и пайплайны
- Зависимости между задачами (`depends_on`)
- Повторные попытки (`retry`)
- Переменные окружения для контейнера
- Таймауты задач (`deadline`)
- Политика хранения файлов (`keep` / `cleanup`)

## Стек

Go, Kafka (`segmentio/kafka-go`), Docker SDK (`moby/moby`), AWS SDK v2 (MinIO), SQLite (`mattn/go-sqlite3`), Cobra, `log/slog`
