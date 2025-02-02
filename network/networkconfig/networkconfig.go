package networkconfig

import "github.com/titosilva/drmchain-pos/internal/di"

type NetworkConfig struct {
	HandshakeHost string
	GossipHost    string
}

func Factory(diCtx *di.DIContext) *NetworkConfig {
	return &NetworkConfig{
		HandshakeHost: "localhost:2503",
		GossipHost:    "localhost:2504",
	}
}

func GetFromDI(diCtx *di.DIContext) *NetworkConfig {
	return di.GetService[NetworkConfig](diCtx)
}
