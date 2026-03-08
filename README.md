# selectel-testcase-linter

Линтер для Go, проверяющий лог-сообщения в `log/slog` и `go.uber.org/zap`.

## Что проверяет

Линтер `logslinter` проверяет 4 правила:

1. сообщение начинается со строчной буквы;
2. сообщение только на английском языке;
3. сообщение не содержит спецсимволы и эмодзи;
4. сообщение не содержит признаков потенциально чувствительных данных.

Поддерживаемые логгеры:

- `log/slog`
- `go.uber.org/zap` (`Logger` и часть API `SugaredLogger`)

## Структура проекта

- `analyzer` — точка входа анализатора на `go/analysis`;
- `internal/loglint/extract` — извлечение message-аргумента из вызовов `slog`/`zap`;
- `internal/loglint/rules` — все правила (`lowercase`, `english`, `no_specials`, `sensitive`) и общий интерфейс;
- `internal/loglint/config` — загрузка и применение настроек плагина;
- `cmd/loglint` — standalone CLI-раннер (`singlechecker`);
- `plugin` — экспорт анализатора для интеграции как module plugin;
- `testdata` — единая директория тестовых Go-пакетов (`analysistest` + integration).

## Требования

- Go `1.25+`

## Установка зависимостей

```bash
go mod tidy
```

## Локальный запуск линтера

```bash
go run ./cmd/loglint ./...
```

## Запуск тестов

```bash
go test ./...
```

## Интеграционный тест

Тест прогоняет `cmd/loglint` на пакете `testdata/src/p` с настоящими импортами `log/slog` и `go.uber.org/zap`:

```bash
go test ./integration -run TestRealRun -v
```

## Примеры нарушений

```text
slog.Info("Starting server")        // uppercase first letter
slog.Error("ошибка подключения")    // non-English
logger.Warn("connection failed!!!") // special symbols
logger.Info("token: " + token)      // potentially sensitive data
```

## Интеграция с golangci-lint (module plugin)

Проект предоставляет пакет `plugin` с регистрацией через `plugin-module-register`.
Для подключения используйте module plugin system из документации golangci-lint:

- [New linters](https://golangci-lint.run/docs/contributing/new-linters/)

Ниже пример конфигурации для кастомной сборки (адаптируйте модульный путь под ваш репозиторий):

```yaml
# .custom-gcl.yml
version: v2.4.0

plugins:
  - module: github.com/S1FFFkA/selectel-testcase-linter
    import: github.com/S1FFFkA/selectel-testcase-linter/plugin

linters:
  enable:
    - logslinter
```

Пример plugin settings (кастомные правила и свои sensitive-слова):

```yaml
plugins:
  - module: github.com/S1FFFkA/selectel-testcase-linter
    import: github.com/S1FFFkA/selectel-testcase-linter/plugin
    settings:
      rules:
        starts_with_lower: true
        english_only: true
        no_emoji_or_special: true
        sensitive_data:
          state: true
          words:
            - "password:"
            - "password="
            - "token:"
            - "merchant_pin"
      autofix:
        starts_with_lower: true
        english_only: true
        no_emoji_or_special: true
        sensitive_data: true
```

## Нюансы текущей реализации

- Правила 1-3 применяются к строковым сообщениям, вычислимым как compile-time string constant.
- Проверка чувствительных данных работает и для динамических выражений (`+`, `fmt.Sprintf`, идентификаторы с именами вроде `password`, `token`, `api_key`).
- Добавлены `SuggestedFixes`: понижение регистра первой буквы, перевод неанглийского текста на английский (через `go_translate`), удаление спецсимволов в литералах, замена статического чувствительного сообщения на нейтральное.
- Чтобы добавить своё "правило-слово" для sensitive, добавьте его в `rules.sensitive_data.words`.