-- Инициализация базы данных для development среды
-- Этот файл выполняется автоматически при первом запуске PostgreSQL контейнера

-- Создание таблицы метрик
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

CREATE INDEX IF NOT EXISTS idx_metrics_type ON metrics(type);
CREATE INDEX IF NOT EXISTS idx_metrics_updated_at ON metrics(updated_at);

ALTER TABLE metrics ADD CONSTRAINT IF NOT EXISTS chk_counter_has_delta 
    CHECK ((type = 'counter' AND delta IS NOT NULL) OR type != 'counter');

ALTER TABLE metrics ADD CONSTRAINT IF NOT EXISTS chk_gauge_has_value 
    CHECK ((type = 'gauge' AND value IS NOT NULL) OR type != 'gauge');

-- Комментарии для документации
COMMENT ON TABLE metrics IS 'Таблица для хранения метрик приложения';
COMMENT ON COLUMN metrics.id IS 'Уникальный идентификатор метрики';
COMMENT ON COLUMN metrics.type IS 'Тип метрики: counter или gauge';
COMMENT ON COLUMN metrics.delta IS 'Значение для counter метрик (накопительное)';
COMMENT ON COLUMN metrics.value IS 'Значение для gauge метрик (текущее)';
COMMENT ON COLUMN metrics.hash IS 'Хеш для проверки целостности данных';

-- Вставляем тестовые данные для development
INSERT INTO metrics (id, type, delta, value, hash) VALUES 
    ('test_counter', 'counter', 100, NULL, ''),
    ('test_gauge', 'gauge', NULL, 42.5, '')
ON CONFLICT (id, type) DO NOTHING;
