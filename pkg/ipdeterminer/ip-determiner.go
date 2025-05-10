package ipdeterminer

import (
	"fmt"
	"net"

	"go.uber.org/zap"
)

func GetPreferredOutboundIP(logger *zap.Logger) (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			logger.Error("failed to close connection", zap.Error(err))
		}
	}(conn)

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}
