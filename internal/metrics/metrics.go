package metrics

import "sync/atomic"

type Metrics struct {
	successCount atomic.Int64
	errorCount   atomic.Int64
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) IncSuccess() {
	m.successCount.Add(1)
}

func (m *Metrics) IncErrors() {
	m.errorCount.Add(1)
}

func (m *Metrics) GetSuccessCount() int64 {
	return m.successCount.Load()
}

func (m *Metrics) GetErrorCount() int64 {
	return m.errorCount.Load()
}

func (m *Metrics) GetTotalCount() int64 {
	return m.successCount.Load() + m.errorCount.Load()
}
