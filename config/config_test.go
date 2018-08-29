package config

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func Test_New(t *testing.T) {
	assert := assert.New(t)
	conf := New("config.json")
	assert.NotNil(conf)
	path := conf.filePath
	ok := strings.Contains(path, "/config/config.json")
	assert.True(ok)
}

func Test_NewNodeConfig(t *testing.T) {
	assert := assert.New(t)
	nodeConf := NewNodeConfig()
	assert.NotNil(nodeConf)
	assert.Equal(uint64(10), nodeConf.TxPoolConf.GlobalSlots)
	assert.NotNil("solo", nodeConf.ParticipatesConf.PolicyName)
	assert.NotNil("solo_node", nodeConf.Account)
}
