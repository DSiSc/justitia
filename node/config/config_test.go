package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewNodeConfig(t *testing.T) {
	assert := assert.New(t)
	nodeConf := NewNodeConfig()
	assert.NotNil(nodeConf)
	assert.Equal(uint64(10), nodeConf.GlobalSlots)
	assert.NotNil(nodeConf.Account)
}
