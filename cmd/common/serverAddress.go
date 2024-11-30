package common

import (
	"fmt"
	"strconv"
	"strings"
)

type ServerAddress struct {
	Host string
	Port int
}

func (s *ServerAddress) String() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func (s *ServerAddress) Set(flagValue string) error {
	split := strings.Split(flagValue, ":")
	if len(split) != 2 {
		return fmt.Errorf("invalid server address format: %s", flagValue)
	}
	s.Host = split[0]
	port, err := strconv.Atoi(split[1])
	if err != nil {
		return err
	}
	if port < 0 || port > 65535 {
		return fmt.Errorf("invalid port: %d", port)
	}
	s.Port = port
	return nil
}
