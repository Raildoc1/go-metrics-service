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
	UpdateMetricValueURL = "/update/{" + TypeParam + "}/{" + KeyParam + "}/{" + ValueParam + "}"
	GetMetricValueURL    = "/value/{" + TypeParam + "}/{" + KeyParam + "}"
)
