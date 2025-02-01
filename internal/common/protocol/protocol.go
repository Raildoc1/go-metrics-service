package protocol

const (
	Gauge   = "gauge"
	Counter = "counter"
)

const (
	TypeParam  = "type"
	KeyParam   = "key"
	ValueParam = "value"
)

const (
	HashHeader = "HashSHA256"
)

const (
	UpdateMetricURL           = "/update/"
	UpdateMetricsURL          = "/updates/"
	GetMetricURL              = "/value/"
	UpdateMetricPathParamsURL = "/update/{" + TypeParam + "}/{" + KeyParam + "}/{" + ValueParam + "}"
	GetMetricPathParamsURL    = "/value/{" + TypeParam + "}/{" + KeyParam + "}"
	PingURL                   = "/ping"
	GetAllMetricsURL          = "/"
)

//nolint:govet // field alignment
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}
