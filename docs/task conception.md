# Planner + Worker: Архитектура и контракты

> Живой документ — дорабатывается по ходу проекта

---

## Общая идея

Два микросервиса общаются через Kafka.

**Planner** — знает про типы задач, расписание, пайплайны, хранит состояние.  
**Worker** — универсальный обработчик, обеспечивает эффективное исполнение.

---

## Сервисы

### Planner

- CLI + daemon на одной машине
- CLI передаёт daemon'у команду и `cwd` откуда был вызван
- Daemon резолвит относительные пути файлов → загружает на file server
- Парсит `@params`, резолвит зависимости pipeline, формирует Task'и с `{{task:id/path}}`-ссылками
- Хранит задачи и историю запусков в SQLite
- Поднимает простой HTTP file server для обмена файлами с worker'ом
- Знает про типы задач, шаблоны команд, расписание — всё внутри planner'а, в контракт не выходит

### Worker

- Универсальный исполнитель shell-команд в Docker-контейнере
- Держит пул горячих контейнеров для быстрого старта
- Базовый образ со стандартными утилитами; можно указать свой образ (но холодный старт)
- Один постоянный volume примонтирован при старте — общий для всех контейнеров на хосте
- Перед запуском: резолвит все `{{task:id/path}}`-ссылки в реальные пути, скачивает `fetch`-файлы
- После запуска: загружает `upload`-файлы на planner file server, выполняет `cleanup`
- Логи (stdout/stderr) пишет в предсказуемое место по соглашению — planner знает где забрать

---

## Ссылки на файлы: {{task:id/path}}

Все ссылки на файлы в контракте используют формат `{{task:id/path}}` — абстракция над реальной файловой системой worker'а. Planner не знает как worker организует хранилище, worker сам резолвит ссылки в конкретные пути.

**Абстракция:** `{{task:abc123/outputs/video.mp4}}`  
**Реализация:** worker резолвит в `/tasks/abc123/outputs/video.mp4` на своём volume

Синтаксис `{{...}}` выбран как достаточно уникальный — двойные фигурные скобки практически не встречаются в реальных shell-командах. Worker парсит и резолвит ссылки до передачи команды в shell, поэтому конфликт с shell-синтаксисом исключён.

---

## @param синтаксис (только в CLI и planner)

Пользователь пишет команды с `@param` — именованными ссылками на файлы. Planner резолвит их в `{{task:id/path}}`-ссылки до отправки в Kafka. Worker про `@params` не знает.

Regex: `@([a-z]+)(?:\.([a-z0-9]+))?` — имя переменной только строчные латинские буквы, опциональное расширение после точки.

| Пишет пользователь | Тип    | Резолвится в `{{task:id/...}}`  | Примечание                      |
|--------------------|--------|---------------------------------|---------------------------------|
| `@input`           | in     | `inputs/video.mp4`              | имя файла из переменной         |
| `@input.mp4`       | in     | `inputs/video.mp4`              | явное расширение переопределяет |
| `@output.mp3`      | out    | `outputs/audio.mp3`             |                                 |
| `@tmp`             | interm | `intermediate/tmp`              | не загружается, не keep         |

---

## Pipeline

Пользователь пишет команды как набор quoted-строк:

```
planner add "yt-dlp @url -o @video" "ffmpeg -i @video -o @audio" "whisper @audio -o @transcript"
```

Planner резолвит зависимости по совпадению `@params` между соседними командами и формирует цепочку Task:

```
@video:       out Task1 → keep,   in Task2 → {{task:task1_id/outputs/video.mp4}}
@audio:       out Task2 → keep,   in Task3 → {{task:task2_id/outputs/audio.mp3}}
@transcript:  out Task3 → upload на file server (финальный результат)
```

Промежуточные файлы не покидают worker-хост. Только финальный результат загружается на file server. Planner заранее вычисляет кому поручить cleanup каждого keep-файла и прописывает это в соответствующую Task.

Все Task одного pipeline получают одинаковый `group_id` как Kafka partition key — гарантирует обработку одним worker-хостом. Внутри хоста каждый шаг может использовать разный Docker образ — volume общий, файлы доступны всем контейнерам.

> Worker в текущей реализации запускается в единственном instance, так что routing зависимых задач на один хост актуален скорее на вырост.

---

## Контракты Kafka

Сообщения в Kafka — JSON, сериализуются в типизированные Go-структуры через `encoding/json` + `github.com/google/uuid`.

### Task

```
Task {
  id:          UUID
  group_id: UUID?         // null если одиночная задача
  depends_on:  [UUID]        // task_id'ы которые должны завершиться до старта

  command:     string        // готовая команда с {{task:id/path}}-ссылками
  image:       string?       // null = базовый образ

  fetch:   [{ file_id: UUID, input: string }]  // скачать с file server; input = {{task:id/path}}
  upload:  [{ input: string, file_id: UUID }]  // загрузить на file server после
  keep:    [string]                          // refs оставить в volume для следующей Task
  cleanup: [string]                          // refs удалить после выполнения (могут быть от других Task)

  env:         Map<string, string>?
  deadline:    timestamp?
  retry:       RetryPolicy?  // не уточняем пока, добавить позже
}
```

### TaskResult

```
TaskResult {
  task_id:     UUID
  group_id: UUID?
  status:      Success | Failure
  exit_code:   i32
  duration_ms: u64
}
```

Логи planner забирает по соглашению — worker всегда пишет stdout/stderr в предсказуемое место относительно task_id. Uploads planner знает заранее из Task — если upload не случился, это failure.

---

## CLI

Основной способ наблюдения за задачами — polling. Никаких уведомлений, пользователь смотрит сам когда нужно.

Везде где принимается `<id|name>` — можно передавать как UUID так и имя задачи/группы (уникальное).

```
planner run "cmd" [--image <image>] [--name <name>] [--in key=value,...] [--out key=value,...]
planner show <id|name>
planner logs <id|name>
planner cancel <id|name>
planner list

planner group create [--name <name>]
planner group add <id|name> "cmd" [--image <image>] [--name <name>]
planner group run <id|name> [--in key=value,...] [--out key=value,...]
planner group show <id|name>
planner group cancel <id|name>
planner group list
```

### planner list

```
ID       NAME         STATUS    CREATED   DURATION  COMMAND
abc123   -            done      2m ago    1.2s      curl https://...
def456   transcode    running   10s ago   -         ffmpeg -i @input...
ghi789   -            failed    1h ago    0.3s      whisper @audio...
```

### planner show <id|name>

```
Task:     ghi789
Name:     -
Status:   failed
Created:  2026-04-02 14:23:11
Started:  2026-04-02 14:23:12
Duration: 0.3s
Command:  whisper {{task:ghi789/inputs/audio.mp3}}
Exit:     1

stderr:   Error: model file not found
          Run 'planner logs ghi789' for full output
```

Первые несколько строк stderr выводятся сразу — не надо идти за логами если ошибка очевидная.

---

## Сценарии

- **Uptime monitor** — периодический `curl` + проверка exit code
- **Watch + alert** — `curl` + `jq`, логика алерта в команде, exit code сигнализирует о пороге
- **Одиночные задачи** — бэкап, очистка, генерация отчётов
- **Pipeline на одном образе** — несколько шагов с одним инструментом
- **Pipeline с разными образами** — `yt-dlp` → `ffmpeg` → `whisper`, каждый шаг свой контейнер, файлы через общий volume

---

## Отложено на потом

| Фича | Статус | Примечание |
|------|--------|------------|
| `RetryPolicy` | Поле есть, значение `null` | Planner при failure не перезапускает |
| Расписание | Не в контракте | Своё, добавляется как слой поверх Task в planner daemon |
| Секреты в `env` | Летят через Kafka открыто | В проде — Vault или k8s secrets по ссылке |
| Удалённый worker | Заложено через file server | Не тестировалось |