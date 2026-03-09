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


## Требования

- Go `1.25+`

## Установка зависимостей

```bash
go mod tidy
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
    version: v1.1.0
```

Пример plugin settings (кастомные правила и свои sensitive-слова):

```yaml
version: "2"

issues:
  max-same-issues: 0
  max-issues-per-linter: 0

linters:
  default: none
  enable:
    - logslinter
  settings:
    custom:
      logslinter:
        type: module
        description: "Linter for log messages."
        settings:
          rules:
            starts_with_lower: true
            english_only: true
            no_emoji_or_special: true
            sensitive_data:
              state: true
              words:
                - password
                - token
                - api_key
                - secret
          autofix:
            starts_with_lower: true
            english_only: true
            no_emoji_or_special: true
            sensitive_data: true
```

## Нюансы текущей реализации

logslinter сейчас работает по таким правилам:

starts_with_lower
Лог должен начинаться с маленькой буквы.

Автофикс: делает первую букву строчной.

english_only
В сообщении не должно быть не-латинских букв (кириллица и т.п.).

Автофикс: переводит текст на английский (если переводчик доступен).

no_emoji_or_special
Запрещены эмодзи и спецсимволы.

Разница:
static (чистый текст): : и = запрещены;

dynamic (есть переменная/выражение): : и = разрешены.

Автофикс: удаляет запрещенные символы.

sensitive_data
Срабатывает только для dynamic сообщений (где есть переменная/поле) и если найден sensitive keyword/regex.

Для static строк не срабатывает.

Автофикса нет (только репорт).

Дополнительно:
Sensitive keywords берутся из конфига (rules.sensitive_data.words), если не заданы — используются дефолтные.

Автофикс не трогает sensitive-значения и не подменяет их на sensitive data redacted.

Для одной проблемной строки в одном прогоне применяется один согласованный фикс (без конфликтов правок).

# Быстрый старт

В корне проекта где хотите подключить линтр для логов создайте два файла 

## .custom-gcl.yml
Его можно взять выше в документации



## .golangci.yml

В нем можно прописать свои треггер-слова 

Отключить/Выключить какие-либо проверки

Отключить/Выключить какие-либо автоисправления 

Пример данного файла можно найти также в документации.

## В командой строке в корне файла пропишите команду
```golangci-lint custom  ```

Требуется немного подождать пока все зависимости подтянуться и после создания exe файла у вас в директории вы сможете использовать две команды 

```.\custom-gcl.exe run -c .golangci.yml ./...      ``` - проверка всех логов на наши правила

```.\custom-gcl.exe run -c .golangci.yml --fix ./...```- также проверяет все логи на наши правила + автоисправляет их , если включены соответсвующие поля в yml файле (рекомендуется прогонять несколько раз , т.к линтер может испраивть лишь одну ошибку за раз)

