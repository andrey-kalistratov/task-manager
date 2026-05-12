# task-manager

Распределённая система запуска задач в Docker-контейнерах.

## Архитектура

Два сервиса общаются через Kafka:

```
    CLI (tm run)                                                                                                                                                                                                 
         │                                                                                                                                                                                                       
         │  Unix socket
         ▼                                                                                                                                                                                                       
   ┌───────────────┐    Kafka: tasks    ┌──────────────┐                                                                                                                                                       
   │    Planner    │ ─────────────────▶ │    Worker    │
   │    daemon     │ ◀───────────────── │              │                                                                                                             
   └───────────────┘    Kafka: results  └──────────────┘
          │    │                               │                                                                                                                                                                 
    state │    │ read results    write results │                                                                                                                                                                 
          ▼    └──────────┐  ┌─────────────────┘
     ┌─────────┐          ▼  ▼                                                                                                                                                                                   
     │ SQLite  │        ┌──────────┐                                                                                                                                                                             
     └─────────┘        │    S3    │                                                                                                                                                                             
                        └──────────┘    
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
| `--in key=path` | Входной файл: ключ = имя в контейнере, path = путь на хосте planner'а |
| `--out key=path` | Выходной файл: ключ = имя в контейнере, path = куда сохранить на хосте |

## Примеры

Все примеры запускаются через `docker exec` в контейнере `planner`. Входные файлы лежат в `testdata/`, который примонтирован в `/app/testdata` внутри контейнера.

### Запуск без файлов

```bash
docker exec planner tm run "echo 'ok' && date"
```

Минимальная проверка что система работает. Результат виден в логах worker'а.

### Сортировка и дедупликация

```bash
docker exec planner tm run "sort -u < words.txt > sorted.txt" \
  --name sort-words \
  --in words.txt=/app/testdata/words.txt \
  --out sorted.txt=/app/testdata/sorted.txt
```

`testdata/words.txt` содержит слова с повторами. После выполнения `testdata/sorted.txt` будет содержать уникальные слова в алфавитном порядке.

### Сборка документа из двух частей

```bash
docker exec planner tm run "cat preamble.txt body.txt > message.txt" \
  --name assemble \
  --in preamble.txt=/app/testdata/preamble.txt \
  --in body.txt=/app/testdata/helloworld.txt \
  --out message.txt=/app/testdata/message.txt
```

Склеивает два файла в один. Результат в `testdata/message.txt`.

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

- **Go**
- **Kafka** — `segmentio/kafka-go`
- **Docker SDK** — `moby/moby`
- **AWS SDK v2** + MinIO
- **SQLite** — `mattn/go-sqlite3`
- **Cobra** (CLI)
- `log/slog`
