package driver

import (
	"context"
	"errors"
	"fmt"
	"go-metrics-service/internal/common/protocol"

	pb "go-metrics-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcDriver struct {
	conn          *grpc.ClientConn
	updateMetrics pb.UpdateMetricsClient
}

type GRPCConfig struct {
	Port uint16
}

func NewGrpcDriver(cfg GRPCConfig) (*GrpcDriver, error) {
	options := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(fmt.Sprintf(":%v", cfg.Port), options)
	if err != nil {
		return nil, err
	}
	updateMetrics := pb.NewUpdateMetricsClient(conn)
	return &GrpcDriver{
		conn:          conn,
		updateMetrics: updateMetrics,
	}, nil
}

func (s *GrpcDriver) SendUpdates(ctx context.Context, metrics []protocol.Metrics) error {
	ms, err := ConvertMetrics(metrics)
	if err != nil {
		return err
	}
	request := pb.UpdateMetricsRequest_builder{
		Values: ms,
	}.Build()
	response, err := s.updateMetrics.UpdateMetrics(ctx, request)
	if err != nil {
		return err
	}
	if response.GetError() != "" {
		return errors.New(response.GetError())
	}
	return nil
}

func ConvertMetrics(ms []protocol.Metrics) ([]*pb.Metric, error) {
	res := make([]*pb.Metric, len(ms))
	for i, metric := range ms {
		m, err := ConvertMetric(metric)
		if err != nil {
			return nil, err
		}
		res[i] = m
	}
	return res, nil
}

func ConvertMetric(m protocol.Metrics) (*pb.Metric, error) {
	switch m.MType {
	case protocol.Gauge:
		id := m.ID
		metricType := pb.Metric_GAUGE
		val := *m.Value
		return pb.Metric_builder{
			Id:    &id,
			Type:  &metricType,
			Value: &val,
		}.Build(), nil
	case protocol.Counter:
		id := m.ID
		metricType := pb.Metric_COUNTER
		delta := *m.Delta
		return pb.Metric_builder{
			Id:    &id,
			Type:  &metricType,
			Delta: &delta,
		}.Build(), nil
	default:
		return nil, errors.New("unknown metric type " + m.MType)
	}
}
