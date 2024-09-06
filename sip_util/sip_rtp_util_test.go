package sip_util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUDPListenPortsConflict(t *testing.T) {
	MediaPortStart = 40100
	InitializePorts()
	for i := 0; i < 201; i++ {
		Port := FindNextFreePort()
		assert.Greater(t, Port, 0, "Non zero free port not found")
		ReleasePort(Port)
	}
}
