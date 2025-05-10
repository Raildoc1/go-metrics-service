package middleware

import (
	"fmt"
	"net"
	"net/http"

	"go.uber.org/zap"
)

type SubnetFilter struct {
	trustedSubnet *net.IPNet
	logger        *zap.Logger
}

func NewSubnetFilter(logger *zap.Logger, trustedSubnet string) (*SubnetFilter, error) {
	_, ipNet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		return nil, fmt.Errorf("error parsing trusted subnet: %w", err)
	}
	return &SubnetFilter{
		trustedSubnet: ipNet,
		logger:        logger,
	}, nil
}

func (sf *SubnetFilter) CreateHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ipStr = r.Header.Get("X-Real-IP")
		var ip = net.ParseIP(ipStr)
		if ip == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !sf.trustedSubnet.Contains(ip) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.ServeHTTP(w, r)
	})
}
