package server

import (
	"net"
	"testing"

	"github.com/bestruirui/octopus/internal/conf"
)

func Test_Start_returns_error_when_address_is_already_in_use(t *testing.T) {
	// Given
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	defer listener.Close()

	address := listener.Addr().(*net.TCPAddr)
	previous := conf.AppConfig.Server
	conf.AppConfig.Server.Host = "127.0.0.1"
	conf.AppConfig.Server.Port = address.Port
	t.Cleanup(func() {
		conf.AppConfig.Server = previous
	})

	// When
	err = Start()

	// Then
	if err == nil {
		t.Fatal("Start() error = nil, want address-in-use error")
	}
}
