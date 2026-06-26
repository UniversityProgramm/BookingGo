package mocks

import (
	"BookingGo/pkg/logger"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	logger.Init()

	code := m.Run()

	os.Exit(code)
}
