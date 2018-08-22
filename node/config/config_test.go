package config

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func Test_NewNodeConfig(t *testing.T) {
	assert := assert.New(t)
	nodeConf := NewNodeConfig()
	assert.NotNil(nodeConf)
	assert.Equal(uint64(10), nodeConf.TxPoolConf.GlobalSlots)
	assert.NotNil("leveldb", nodeConf.LedgerConf.PluginName)
	assert.NotNil("./data", nodeConf.LedgerConf.DataPath)
	assert.NotNil("solo", nodeConf.ParticipatesConf.PolicyName)
}

func Test_New(t *testing.T) {
	assert := assert.New(t)
	conf := New("config.json")
	assert.NotNil(conf)
	path := conf.filePath
	ok := strings.Contains(path, "node/config/config.json")
	assert.True(ok)
}
