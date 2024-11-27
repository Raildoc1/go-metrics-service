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
	UpdateMetricValueUrl = "/update/{" + TypeParam + "}/{" + KeyParam + "}/{" + ValueParam + "}"
	GetMetricValueUrl    = "/value/{" + TypeParam + "}/{" + KeyParam + "}"
)
