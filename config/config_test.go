package config

import (
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/monkey"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewNodeConfig(t *testing.T) {
	monkey.Patch(GetLogSetting, func(*viper.Viper) log.Config {
		return log.Config{}
	})
	assert := assert.New(t)
	nodeConf := NewNodeConfig()
	assert.NotNil(nodeConf)
	assert.NotNil(nodeConf.AlgorithmConf)
	assert.Equal("SHA256", nodeConf.AlgorithmConf.HashAlgorithm)
	assert.Equal(uint64(4096), nodeConf.TxPoolConf.GlobalSlots)
	assert.NotNil("solo", nodeConf.ParticipatesConf.PolicyName)
	assert.NotNil("solo_node", nodeConf.Account)
	assert.Equal("tcp://0.0.0.0:47768", nodeConf.ApiGatewayAddr)
	assert.Equal(int64(2000), nodeConf.BlockInterval)
	assert.Equal(uint64(4), nodeConf.ParticipatesConf.Delegates)
	var address = types.Address{
		0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33,
		0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d,
	}
	assert.Equal(address, nodeConf.Account.Address)
	assert.Equal(int64(5000), nodeConf.ConsensusConf.Timeout.TimeoutToCollectResponseMsg)
	assert.Equal(int64(5000), nodeConf.ConsensusConf.Timeout.TimeoutToWaitCommitMsg)
	assert.Equal(int64(30000), nodeConf.ConsensusConf.Timeout.TimeoutToChangeView)
	monkey.UnpatchAll()
}
