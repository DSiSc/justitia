package config

import (
	"github.com/DSiSc/txpool/common"
)

type NodeConfig struct {
	// txpool
	GlobalSlots uint64

	// default account
	Account common.Address
}

func NewNodeConfig() NodeConfig {
	// TODO: get account and globalSlots from genesis.json
	var temp common.Address
	var globalSlots uint64 = 10
	return NodeConfig{
		GlobalSlots: globalSlots,
		Account:     temp,
	}
}
