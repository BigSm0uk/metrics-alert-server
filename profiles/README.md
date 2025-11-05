# Результаты профилирования

## Анализ аллокаций (analysis.txt)

**Дата:** 2025-11-05 10:07:56  
**Всего аллокаций:** 418,405,129 (418 миллионов)

### Основные проблемы

#### 1. 🔴 HTML Templates - 74% аллокаций (309M)

```
reflect.Value.call          34.65% (145M аллокаций)
text/template.evalCall      88.79% кумулятивно (371M)
html/template.htmlReplacer   7.64% (32M)
```

**Проблема:** HTML шаблоны используют рефлексию и создают огромное количество временных объектов.


**Решение:**
- ✅ Кешировать результат рендеринга
- ✅ Использовать `sync.Pool` для буферов
- ✅ Рассмотреть JSON API вместо HTML для production

#### 2. 🟡 PostgreSQL Repository - 8.61% (36M)

```
pg.(*PostgresRepository).MetricList  4.11% (17M)
pgtype.scanPlanString.Scan          2.36% (9.8M)
```

**Проблема:** Создание слайсов результатов без capacity

**Решение:**
```go
// В MetricList
result := make([]domain.Metrics, 0, expectedSize)  // предвыделить capacity
```

---

## Результаты оптимизации (diff base → result)

На основании сравнения профилей памяти (inuse_space) стало заметно легче в ключевых местах при той же нагрузке:

- text/template/html/template: суммарно около −26.5 MB
  - `text/template.(*Template).Execute`: −26.5 MB
  - `text/template.(*state).eval* / walk*`: −4.6 MB (совокупно)
- Handler `GetAllMetrics` и обвязка (router/openapi wrapper): около −28.3 MB
  - `pkg/openapi/metric.(*ServerInterfaceWrapper).GetAllMetrics`: −29.1 MB
  - `internal/handler.(*MetricHandler).GetAllMetrics`: −29.1 MB (по цепочке)
- GZIP middleware (ответ): около −21.9 MB
  - `middleware.(*gzipResponseWriter).Write`: −21.9 MB

Фиксации достигнуты за счёт кеширования HTML-рендера и снижения аллокаций в пути выдачи страницы.

## Рекомендации по приоритетам

### Высокий приоритет 🔴

**1. Оптимизировать HTML endpoint**

```go
// Было
func (h *Handler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
    metrics, _ := h.service.GetAllMetrics(ctx)
    tmpl.Execute(w, metrics)  // каждый раз рендерим
}

// Стало
var (
    cachedHTML     []byte
    cachedTime     time.Time
    cacheDuration  = 5 * time.Second
    cacheMu        sync.RWMutex
)

func (h *Handler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
    cacheMu.RLock()
    if time.Since(cachedTime) < cacheDuration && cachedHTML != nil {
        w.Write(cachedHTML)
        cacheMu.RUnlock()
        return
    }
    cacheMu.RUnlock()
    
    // Рендерим только если кеш устарел
    metrics, _ := h.service.GetAllMetrics(ctx)
    
    var buf bytes.Buffer
    tmpl.Execute(&buf, metrics)
    
    cacheMu.Lock()
    cachedHTML = buf.Bytes()
    cachedTime = time.Now()
    cacheMu.Unlock()
    
    w.Write(cachedHTML)
}
```

**Ожидаемый эффект:** Снижение аллокаций на 70-80% (с 309M до ~60M)

### Средний приоритет 🟡

**2. Оптимизировать PostgreSQL запросы**

```go
// В MetricList
func (r *PostgresRepository) MetricList(ctx context.Context) ([]domain.Metrics, error) {
    // Сначала узнать размер
    var count int
    r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM metrics").Scan(&count)
    
    // Предвыделить capacity
    result := make([]domain.Metrics, 0, count)
    
    // ...
}
```

**Ожидаемый эффект:** Снижение аллокаций на 5-10%

---