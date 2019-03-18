package config

import (
	"github.com/DSiSc/justitia/compiler"
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
	monkey.Patch(compiler.SolidityCompile, func(string) string {
		return "608060405234801561001057600080fd5b506040805190810160405280600d81526020017f48656c6c6f2c20776f72"
	})
	block, err := GenerateGenesisBlock()
	assert.NotNil(block)
	assert.Nil(err)
	assert.Equal(uint64(5), uint64(len(block.GenesisAccounts)))
	assert.Equal(uint64(0), block.Block.Header.Height)
}
