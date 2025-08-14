-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS metrics (
    id VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('counter', 'gauge')),
    delta BIGINT,
    value DOUBLE PRECISION,
    hash VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    PRIMARY KEY (id, type)
);

-- Создаем индексы отдельно
CREATE INDEX IF NOT EXISTS idx_metrics_type ON metrics(type);
CREATE INDEX IF NOT EXISTS idx_metrics_updated_at ON metrics(updated_at);

-- Добавляем ограничения
ALTER TABLE metrics ADD CONSTRAINT chk_counter_has_delta 
    CHECK ((type = 'counter' AND delta IS NOT NULL) OR type != 'counter');

ALTER TABLE metrics ADD CONSTRAINT chk_gauge_has_value 
    CHECK ((type = 'gauge' AND value IS NOT NULL) OR type != 'gauge');

-- Комментарии для документации
COMMENT ON TABLE metrics IS 'Таблица для хранения метрик приложения';
COMMENT ON COLUMN metrics.id IS 'Уникальный идентификатор метрики';
COMMENT ON COLUMN metrics.type IS 'Тип метрики: counter или gauge';
COMMENT ON COLUMN metrics.delta IS 'Значение для counter метрик (накопительное)';
COMMENT ON COLUMN metrics.value IS 'Значение для gauge метрик (текущее)';
COMMENT ON COLUMN metrics.hash IS 'Хеш для проверки целостности данных';
COMMENT ON COLUMN metrics.created_at IS 'Время создания записи';
COMMENT ON COLUMN metrics.updated_at IS 'Время последнего обновления';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS metrics;
-- +goose StatementEnd
