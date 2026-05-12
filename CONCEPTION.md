# Planner + Worker: Архитектура и контракты

## Общая идея

Два микросервиса общаются через Kafka.

**Planner** — знает про типы задач, пайплайны, хранит состояние в SQLite.  
**Worker** — универсальный исполнитель shell-команд в Docker-контейнерах.

---

## Сервисы

### Planner

- CLI (`tm`) + daemon (`tmd`) на одной машине; CLI передаёт daemon'у команду и `cwd`
- Daemon резолвит относительные пути → загружает файлы в S3
- Парсит `@params`, резолвит зависимости pipeline, формирует Task'и
- Хранит задачи и историю запусков в SQLite
- Подписан на результаты из Kafka, обновляет статусы и скачивает выходные файлы из S3

### Worker

- Получает Task из Kafka, скачивает входные файлы из S3 в рабочую директорию задачи
- Запускает команду в Docker-контейнере; можно указать произвольный образ
- После выполнения загружает выходные файлы в S3, публикует TaskResult в Kafka

---

## CLI

Везде где принимается `<id|name>` — можно передавать UUID или имя задачи/группы.  
Основной способ наблюдения за задачами — polling.

```
tm run "cmd" [--image <image>] [--name <name>] [--in key=value,...] [--out key=value,...]
tm show <id|name>
tm logs <id|name>
tm cancel <id|name>
tm list

tm group create [--name <name>]
tm group add <id|name> "cmd" [--image <image>] [--name <name>]
tm group run <id|name> [--in key=value,...] [--out key=value,...]
tm group show <id|name>
tm group cancel <id|name>
tm group list
```

### tm list

```
ID       NAME         STATUS    CREATED    DURATION  COMMAND
abc123   -            done      2m ago     1.2s      curl https://...
def456   transcode    running   10s ago    -         ffmpeg -i @input...
ghi789   -            failed    1h ago     0.3s      whisper @audio...
```

### tm show

```
Task:     ghi789
Name:     -
Status:   failed
Created:  2026-04-02 14:23:11
Duration: 0.3s
Command:  whisper @audio

stderr:   Error: model file not found
          Run 'tm logs ghi789' for full output
```

Первые строки stderr выводятся сразу — не надо идти за логами если ошибка очевидная.

---

## @param синтаксис

Пользователь пишет команды с `@param` — именованными ссылками на файлы. Planner резолвит их в конкретные пути до отправки Task в Kafka. Worker про `@params` не знает.

Формат: `@name` или `@name.ext` — только строчные латинские буквы, опциональное расширение.

| Пишет пользователь | Роль          | Примечание                         |
|--------------------|---------------|------------------------------------|
| `@input`           | входной файл  | имя файла берётся из переменной    |
| `@input.mp4`       | входной файл  | явное расширение переопределяет    |
| `@output.mp3`      | выходной файл |                                    |
| `@tmp`             | промежуточный | не загружается в S3, не сохраняется|

---

## Pipeline

Несколько команд передаются как цепочка:

```
tm group add <id> "yt-dlp @url -o @video" \
             "ffmpeg -i @video -o @audio" \
             "whisper @audio -o @transcript"
```

Planner резолвит зависимости по совпадению `@params` между командами:

```
@video:       out Task1 → keep,   in Task2
@audio:       out Task2 → keep,   in Task3
@transcript:  out Task3 → финальный результат (загружается в S3)
```

Промежуточные файлы (`keep`) остаются в рабочей директории worker'а между шагами и не гоняются через S3. Planner заранее вычисляет, какая Task должна сделать `cleanup` чужих keep-файлов, и прописывает это в контракт.

Все Task одного pipeline получают одинаковый `group_id` как Kafka partition key — гарантирует обработку одним worker'ом. Каждый шаг может использовать разный Docker-образ.

---

## Контракты Kafka

Сообщения — JSON, типизированные Go-структуры.

### Task

```
Task {
  id:          UUID
  group_id:    UUID?
  depends_on:  [UUID]

  command:     string
  image:       string?

  inputs:      Map<string, File>
  outputs:     Map<string, File>
  keep:        [string]
  cleanup:     [string]

  env:         Map<string, string>?
  deadline:    timestamp?
  retry:       RetryPolicy?
}

File {
  path:     string
  provider: "s3" | "fs"
}
```

### TaskResult

```
TaskResult {
  task_id:  UUID
  group_id: UUID?
  status:   "succeeded" | "failed"
}
```

---

## Сценарии

- **Одиночные задачи** — бэкап, очистка, генерация отчётов
- **Uptime monitor** — периодический `curl` + проверка exit code
- **Watch + alert** — `curl` + `jq`, логика алерта в команде, exit code сигнализирует о пороге
- **Pipeline на одном образе** — несколько шагов с одним инструментом
- **Pipeline с разными образами** — `yt-dlp` → `ffmpeg` → `whisper`, каждый шаг свой контейнер

---

## Планы развития

- **Groups / pipeline** — `depends_on`, `keep`, `cleanup`, команды `tm group *`
- **Наблюдение** — `tm show`, `tm logs`, `tm cancel`, `tm list`
- **Retry** — поле есть в контракте, логика повторов на стороне planner'а
- **Расписание** — слой поверх Task внутри planner daemon
- **Секреты в `env`** — в проде передавать по ссылке (Vault, k8s secrets), не значением
