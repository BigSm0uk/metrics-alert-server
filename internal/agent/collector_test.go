package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricsCollector_CollectRuntimeMetrics(t *testing.T) {
	test := struct {
		name string
		c    *MetricsCollector
	}{
		name: "collect runtime metrics",
		c:    NewMetricsCollector(),
	}
	t.Run(test.name, func(t *testing.T) {
		test.c.CollectRuntimeMetrics()
		assert.NotEqual(t, 0, len(test.c.metrics))
		assert.Equal(t, int64(1), test.c.pollCount)
	})
}
