package repository

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/storage"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
)

func TestMemRepository_Save(t *testing.T) {
	// Arrange
	storage := storage.NewMemStorage()
	repo := NewMemRepository(storage)

	value := 42.5
	metric := &domain.Metrics{
		ID:    "test_gauge",
		MType: domain.Gauge,
		Value: &value,
	}

	// Act
	err := repo.Save(metric)

	// Assert
	require.NoError(t, err)

	// Verify metric was saved
	saved, err := repo.Get("test_gauge")
	require.NoError(t, err)
	assert.Equal(t, metric.ID, saved.ID)
	assert.Equal(t, metric.MType, saved.MType)
	assert.Equal(t, *metric.Value, *saved.Value)
}

func TestMemRepository_Get(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*MemRepository)
		id       string
		wantErr  bool
		expected *domain.Metrics
	}{
		{
			name: "existing gauge metric",
			setup: func(repo *MemRepository) {
				value := 123.45
				metric := &domain.Metrics{
					ID:    "existing_gauge",
					MType: domain.Gauge,
					Value: &value,
				}
				repo.Save(metric)
			},
			id:      "existing_gauge",
			wantErr: false,
			expected: &domain.Metrics{
				ID:    "existing_gauge",
				MType: domain.Gauge,
				Value: func() *float64 { v := 123.45; return &v }(),
			},
		},
		{
			name: "existing counter metric",
			setup: func(repo *MemRepository) {
				delta := int64(100)
				metric := &domain.Metrics{
					ID:    "existing_counter",
					MType: domain.Counter,
					Delta: &delta,
				}
				repo.Save(metric)
			},
			id:      "existing_counter",
			wantErr: false,
			expected: &domain.Metrics{
				ID:    "existing_counter",
				MType: domain.Counter,
				Delta: func() *int64 { v := int64(100); return &v }(),
			},
		},
		{
			name:     "non-existing metric",
			setup:    func(repo *MemRepository) {},
			id:       "non_existing",
			wantErr:  true,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			storage := storage.NewMemStorage()
			repo := NewMemRepository(storage)
			tt.setup(repo)

			// Act
			result, err := repo.Get(tt.id)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected.ID, result.ID)
				assert.Equal(t, tt.expected.MType, result.MType)
				if tt.expected.Value != nil {
					require.NotNil(t, result.Value)
					assert.Equal(t, *tt.expected.Value, *result.Value)
				}
				if tt.expected.Delta != nil {
					require.NotNil(t, result.Delta)
					assert.Equal(t, *tt.expected.Delta, *result.Delta)
				}
			}
		})
	}
}

func TestMemRepository_GetAll(t *testing.T) {
	// Arrange
	storage := storage.NewMemStorage()
	repo := NewMemRepository(storage)

	// Add test metrics
	value1 := 10.5
	delta1 := int64(5)
	value2 := 20.7

	metrics := []*domain.Metrics{
		{ID: "gauge1", MType: domain.Gauge, Value: &value1},
		{ID: "counter1", MType: domain.Counter, Delta: &delta1},
		{ID: "gauge2", MType: domain.Gauge, Value: &value2},
	}

	for _, metric := range metrics {
		err := repo.Save(metric)
		require.NoError(t, err)
	}

	// Act
	result, err := repo.GetAll()

	// Assert
	require.NoError(t, err)
	assert.Len(t, result, 3)

	// Check that all metrics are present
	ids := make(map[string]bool)
	for _, metric := range result {
		ids[metric.ID] = true
	}

	assert.True(t, ids["gauge1"])
	assert.True(t, ids["counter1"])
	assert.True(t, ids["gauge2"])
}

func TestMemRepository_Delete(t *testing.T) {
	// Arrange
	storage := storage.NewMemStorage()
	repo := NewMemRepository(storage)

	value := 42.0
	metric := &domain.Metrics{
		ID:    "to_delete",
		MType: domain.Gauge,
		Value: &value,
	}

	err := repo.Save(metric)
	require.NoError(t, err)

	// Verify metric exists
	_, err = repo.Get("to_delete")
	require.NoError(t, err)

	// Act
	err = repo.Delete("to_delete")

	// Assert
	require.NoError(t, err)

	// Verify metric was deleted
	_, err = repo.Get("to_delete")
	assert.Error(t, err)
}

func TestMemRepository_ConcurrentAccess(t *testing.T) {
	// Arrange
	storage := storage.NewMemStorage()
	repo := NewMemRepository(storage)

	// Act & Assert - test concurrent writes and reads
	done := make(chan bool, 2)

	// Goroutine 1: Write metrics
	go func() {
		for i := range 100 {
			value := float64(i)
			metric := &domain.Metrics{
				ID:    fmt.Sprintf("metric_%d", i),
				MType: domain.Gauge,
				Value: &value,
			}
			err := repo.Save(metric)
			assert.NoError(t, err)
		}
		done <- true
	}()

	// Goroutine 2: Read metrics
	go func() {
		for range 100 {
			_, err := repo.GetAll()
			assert.NoError(t, err)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Verify final state
	metrics, err := repo.GetAll()
	require.NoError(t, err)
	assert.Len(t, metrics, 100)
}
