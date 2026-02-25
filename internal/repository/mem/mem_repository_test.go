package mem

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/storage"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
)

func TestMemRepository_Get(t *testing.T) {
	type args struct {
		id string
		t  string
	}
	r := NewMemRepository(storage.NewMemStorage())
	r.SaveOrUpdate(context.Background(), &domain.Metrics{
		ID:    "Alloc",
		MType: domain.Gauge,
		Value: lo.ToPtr(float64(1024)),
	})
	r.SaveOrUpdate(context.Background(), &domain.Metrics{
		ID:    "PollCount",
		MType: domain.Counter,
		Delta: lo.ToPtr(int64(5)),
	})

	tests := []struct {
		name    string
		r       *MemRepository
		args    args
		want    *domain.Metrics
		wantErr bool
	}{
		{
			name: "get gauge",
			r:    r,
			args: args{
				id: "Alloc",
				t:  domain.Gauge,
			},
			want: &domain.Metrics{
				ID:    "Alloc",
				MType: domain.Gauge,
				Value: lo.ToPtr(float64(1024)),
			},
			wantErr: false,
		},
		{
			name: "get counter",
			r:    r,
			args: args{
				id: "PollCount",
				t:  domain.Counter,
			},
			want: &domain.Metrics{
				ID:    "PollCount",
				MType: domain.Counter,
				Delta: lo.ToPtr(int64(5)),
			},
			wantErr: false,
		},
		{
			name: "get not found",
			r:    r,
			args: args{
				id: "NotExist",
				t:  domain.Gauge,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.Metric(context.Background(), tt.args.id, tt.args.t)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMemRepository_Save(t *testing.T) {
	type args struct {
		metric *domain.Metrics
	}
	r := NewMemRepository(storage.NewMemStorage())
	tests := []struct {
		name    string
		r       *MemRepository
		args    args
		wantErr bool
	}{
		{
			name: "save gauge",
			r:    r,
			args: args{
				metric: &domain.Metrics{
					ID:    "Alloc",
					MType: domain.Gauge,
					Value: lo.ToPtr(float64(1024)),
				},
			},
			wantErr: false,
		},
		{
			name: "save counter",
			r:    r,
			args: args{
				metric: &domain.Metrics{
					ID:    "PollCount",
					MType: domain.Counter,
					Delta: lo.ToPtr(int64(5)),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.r.SaveOrUpdate(context.Background(), tt.args.metric)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMemRepository_GetAll(t *testing.T) {
	tests := []struct {
		name    string
		r       *MemRepository
		want    []domain.Metrics
		wantErr bool
	}{
		{
			name: "get all",
			r:    createRepoWithData(),
			want: []domain.Metrics{
				{
					ID:    "Alloc",
					MType: domain.Gauge,
					Value: lo.ToPtr(float64(1024)),
				},
				{
					ID:    "PollCount",
					MType: domain.Counter,
					Delta: lo.ToPtr(int64(5)),
				},
			},
			wantErr: false,
		},
		{
			name:    "get all empty",
			r:       NewMemRepository(storage.NewMemStorage()),
			want:    []domain.Metrics{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.MetricList(context.Background())
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.ElementsMatch(t, tt.want, got)
			}
		})
	}
}

func createRepoWithData() *MemRepository {
	r := NewMemRepository(storage.NewMemStorage())
	a, p := &domain.Metrics{
		ID:    "Alloc",
		MType: domain.Gauge,
		Value: lo.ToPtr(float64(1024)),
	}, &domain.Metrics{
		ID:    "PollCount",
		MType: domain.Counter,
		Delta: lo.ToPtr(int64(5)),
	}
	r.SaveOrUpdate(context.Background(), a)
	r.SaveOrUpdate(context.Background(), p)
	return r
}
