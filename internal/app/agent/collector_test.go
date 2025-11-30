package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMetricsCollector_CollectRuntimeMetrics(t *testing.T) {
	test := struct {
		name string
		c    *MetricsCollector
	}{
		name: "collect runtime metrics",
		c:    NewMetricsCollector(zap.NewNop()),
	}
	t.Run(test.name, func(t *testing.T) {
		test.c.CollectRuntimeMetrics()
		assert.NotEqual(t, 0, len(test.c.metrics))
		assert.Equal(t, int64(1), test.c.pollCount)
	})
}
