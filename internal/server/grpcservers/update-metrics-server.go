package grpcservers

import (
	"context"
	"errors"
	"go-metrics-service/internal/common/protocol"
	pb "go-metrics-service/proto"
)

var _ pb.UpdateMetricsServer = (*UpdateMetricsServer)(nil)

type UpdateMetricsServer struct {
	pb.UnimplementedUpdateMetricsServer
	controller Controller
}

type Controller interface {
	UpdateMany(ctx context.Context, metrics []protocol.Metrics) error
}

func NewUpdateMetricsServer(controller Controller) *UpdateMetricsServer {
	return &UpdateMetricsServer{
		controller: controller,
	}
}

func (s UpdateMetricsServer) UpdateMetrics(ctx context.Context, request *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	var response pb.UpdateMetricsResponse

	metrics, err := ConvertMetrics(request.Values)
	if err != nil {
		return nil, err
	}

	err = s.controller.UpdateMany(ctx, metrics)
	if err != nil {
		response.Error = err.Error()
	}

	return &response, nil
}

func ConvertMetrics(ms []*pb.Metric) ([]protocol.Metrics, error) {
	metrics := make([]protocol.Metrics, len(ms))
	for i, value := range ms {
		m, err := ConvertMetric(value)
		if err != nil {
			return nil, err
		}
		metrics[i] = m
	}
	return metrics, nil
}

func ConvertMetric(m *pb.Metric) (protocol.Metrics, error) {
	switch m.Type {
	case pb.Metric_COUNTER:
		delta := m.Delta
		return protocol.Metrics{
			ID:    m.Id,
			MType: protocol.Counter,
			Value: nil,
			Delta: &delta,
		}, nil
	case pb.Metric_GAUGE:
		value := m.Value
		return protocol.Metrics{
			ID:    m.Id,
			MType: protocol.Gauge,
			Value: &value,
			Delta: nil,
		}, nil
	default:
		return protocol.Metrics{}, errors.New("unknown type " + m.Type.String())
	}
}
