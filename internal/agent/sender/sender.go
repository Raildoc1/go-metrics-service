package sender

import (
	"fmt"
	"go-metrics-service/internal/agent/requester"
	"go-metrics-service/internal/agent/storage"
	"go-metrics-service/internal/common/protocol"
)

type Sender struct {
	storage   *storage.Storage
	requester *requester.Requester
}

func New(storage *storage.Storage, requester *requester.Requester) *Sender {
	return &Sender{
		storage:   storage,
		requester: requester,
	}
}

func (s *Sender) Send() error {
	metricsDiff := s.storage.GetUncommitedData()
	metricsToUpdateCount := len(metricsDiff.CounterDeltas) + len(metricsDiff.GaugeValues)
	if metricsToUpdateCount == 0 {
		return nil
	}
	metricsToSend := make([]protocol.Metrics, 0, metricsToUpdateCount)
	for k, v := range metricsDiff.CounterDeltas {
		val := v
		metricsToSend = append(
			metricsToSend,
			protocol.Metrics{
				ID:    k,
				MType: protocol.Counter,
				Delta: &val,
			},
		)
	}
	for k, v := range metricsDiff.GaugeValues {
		val := v
		metricsToSend = append(
			metricsToSend,
			protocol.Metrics{
				ID:    k,
				MType: protocol.Gauge,
				Value: &val,
			},
		)
	}
	err := s.requester.SendUpdates(metricsToSend)
	if err != nil {
		return fmt.Errorf("sending updates failed: %w", err)
	}
	s.storage.Commit()
	return nil
}
