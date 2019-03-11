package config

import (
	"github.com/DSiSc/justitia/tools"
	"github.com/DSiSc/monkey"
	"github.com/stretchr/testify/assert"
	"testing"
)

// test build default genesis block
func TestBuildDefaultGensisBlock(t *testing.T) {
	assert := assert.New(t)
	monkey.Patch(tools.PathExists, func(string) bool {
		return false
	})
	block, err := GenerateGenesisBlock()
	assert.NotNil(block)
	assert.Nil(err)
	monkey.UnpatchAll()
}

// test build genesis block from config file
func TestBuildGensisBlockFromFile(t *testing.T) {
	assert := assert.New(t)
	block, err := GenerateGenesisBlock()
	assert.NotNil(block)
	assert.Nil(err)
	assert.Equal(uint64(2), uint64(len(block.GenesisAccounts)))
	assert.Equal(uint64(0), block.Block.Header.Height)
}
