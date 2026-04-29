package metrics

type Metrics struct {
	successCount int
	errorCount   int
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) IncSuccess() {
	m.successCount++
}

func (m *Metrics) IncErrors() {
	m.errorCount++
}

func (m *Metrics) GetSuccessCount() int {
	return m.successCount
}

func (m *Metrics) GetErrorCount() int {
	return m.errorCount
}

func (m *Metrics) GetTotalCount() int {
	return m.successCount + m.errorCount
}
