# Архитектура `internal/recap`

`recap` организован по бизнес-возможностям, а не как один общий API-пакет.
В корне каталога нет Go-файлов: каждый пакет имеет одну ответственность.

```text
internal/recap/
├── model/                 # доменные типы, enum, JSON union и нормализация
├── ruleset/               # версия, пороги, политики, validation и digest правил
├── analytics/             # производные метрики, календарный период, каталог категорий
├── behavior/              # определение пользовательского поведения и evidence
├── achievement/           # каталог, выбор портфеля и тематические достижения
├── nextaction/            # выбор CTA, тексты, правила и типизированные targets
├── presentation/
│   ├── cards/             # сборка персональной истории карточек
│   └── share/             # безопасная публичная проекция
├── validation/
│   └── structural/        # форма моделей и локальные инварианты
├── integrity/             # пересчёт и сверка всего derived data из storage
├── application/           # use cases, порты, генерация, чтение и share
├── testkit/               # общие детерминированные fixtures и storage doubles
└── docs/                  # документация модуля
```

## Направление зависимостей

```text
model
  ├── ruleset
  └── analytics
        ├── behavior
        ├── achievement
        └── nextaction
              └── presentation

model + analytics + ruleset
              └── validation/structural

derived packages + presentation + validation/structural
              └── integrity
                     └── application
```

Практически это означает:

- `model` не знает о правилах, сервисах и хранилищах;
- `ruleset` хранит конфигурацию, но не запускает бизнес-сценарии;
- `behavior`, `achievement` и `nextaction` можно тестировать и менять независимо;
- `presentation` только преобразует готовые доменные результаты в карточки;
- `validation/structural` проверяет форму моделей, локальные ограничения и согласованность полей, не пересчитывая behavior, achievements, CTA и cards;
- `integrity` считает storage недоверенным, повторно вычисляет derived data и сравнивает его с сохранённым результатом;
- `application` является единственной точкой orchestration и зависит от портов, а не от реализаций storage.

Архитектурный тест `internal/architecture/recap_dependencies_test.go` автоматически запрещает обратные импорты в `application`, зависимости доменных вычислений от `presentation`, импорты storage из `internal/recap` и Go-файлы в корне модуля.

## Основные точки входа

```go
configured := ruleset.DefaultRuleset()
service, err := application.NewService(profiles, analyticsStore, stateStore, recapStore)

detected := behavior.DetectWithRuleset(configured, metrics)
achievements := achievement.BuildWithRuleset(configured, metrics)
action := nextaction.BuildWithRuleset(configured, metrics, state, detected)
```

Хранилища используют типы из `model` и реализуют интерфейсы из `application`.
