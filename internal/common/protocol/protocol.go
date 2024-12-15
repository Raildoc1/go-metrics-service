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
	UpdateJsonURL         = "/update"
	GetMetricValueJsonURL = "/value"
	UpdateMetricValueURL  = "/update/{" + TypeParam + "}/{" + KeyParam + "}/{" + ValueParam + "}"
	GetMetricValueURL     = "/value/{" + TypeParam + "}/{" + KeyParam + "}"
	GetAllMetricsURL      = "/"
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}
