package consts

import (
	"fmt"
)

const (
	FcDaemonProxyPort = 50050
	FcDaemonPort      = 50051
	ResMngrPort       = 50052
	AvaMngrPort       = 50053
	AvaMngrGPUPort    = 3333
	ResMngrIP         = "127.0.0.1"
	AvaWorkerPortBase = 4000
)

func PortToString(p int) string {
	return fmt.Sprintf(":%d", p)
}
