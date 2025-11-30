package domain

import (
	"testing"
)

func TestMetrics_Reset(t *testing.T) {
	tests := []struct {
		name   string
		before *Metrics
		after  *Metrics
	}{
		{
			name: "reset all fields",
			before: &Metrics{
				ID:    "test-id",
				MType: Counter,
				Delta: ptr(int64(100)),
				Value: ptr(float64(3.14)),
				Hash:  "test-hash",
			},
			after: &Metrics{
				ID:    "",
				MType: "",
				Delta: ptr(int64(0)),
				Value: ptr(float64(0.0)),
				Hash:  "",
			},
		},
		{
			name: "reset with nil pointers",
			before: &Metrics{
				ID:    "test-id",
				MType: Gauge,
				Delta: nil,
				Value: nil,
				Hash:  "hash",
			},
			after: &Metrics{
				ID:    "",
				MType: "",
				Delta: nil,
				Value: nil,
				Hash:  "",
			},
		},
		{
			name:   "reset nil struct",
			before: nil,
			after:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Вызываем Reset
			tt.before.Reset()

			// Проверяем nil
			if tt.before == nil && tt.after == nil {
				return
			}

			// Проверяем строковые поля
			if tt.before.ID != tt.after.ID {
				t.Errorf("ID: got %q, want %q", tt.before.ID, tt.after.ID)
			}
			if tt.before.MType != tt.after.MType {
				t.Errorf("MType: got %q, want %q", tt.before.MType, tt.after.MType)
			}
			if tt.before.Hash != tt.after.Hash {
				t.Errorf("Hash: got %q, want %q", tt.before.Hash, tt.after.Hash)
			}

			// Проверяем Delta
			if (tt.before.Delta == nil) != (tt.after.Delta == nil) {
				t.Errorf("Delta nil mismatch: got %v, want %v", tt.before.Delta, tt.after.Delta)
			} else if tt.before.Delta != nil && *tt.before.Delta != *tt.after.Delta {
				t.Errorf("Delta: got %d, want %d", *tt.before.Delta, *tt.after.Delta)
			}

			// Проверяем Value
			if (tt.before.Value == nil) != (tt.after.Value == nil) {
				t.Errorf("Value nil mismatch: got %v, want %v", tt.before.Value, tt.after.Value)
			} else if tt.before.Value != nil && *tt.before.Value != *tt.after.Value {
				t.Errorf("Value: got %f, want %f", *tt.before.Value, *tt.after.Value)
			}
		})
	}
}

// ptr - вспомогательная функция для создания указателей
func ptr[T any](v T) *T {
	return &v
}
