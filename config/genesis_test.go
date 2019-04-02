package config

import (
	"errors"
	"github.com/DSiSc/blockchain"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/justitia/compiler"
	"github.com/DSiSc/justitia/tools"
	"github.com/DSiSc/monkey"
	"github.com/stretchr/testify/assert"
	"reflect"
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

// test import genesis block: exists block in local database
func TestImportGenesisBlockExistBlockInDB(t *testing.T) {
	assert := assert.New(t)
	defer func() {
		r := recover()
		assert.Nil(r)
	}()
	defer monkey.UnpatchAll()
	monkey.Patch(tools.PathExists, func(string) bool {
		return false
	})
	chain := &blockchain.BlockChain{}
	monkey.Patch(blockchain.NewLatestStateBlockChain, func() (*blockchain.BlockChain, error) {
		return chain, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(chain), "GetCurrentBlock", func(*blockchain.BlockChain) *types.Block {
		return &types.Block{
			Header: &types.Header{
				Height: 1,
			},
		}
	})
	ImportGenesisBlock()
}

// test import genesis block: have no block in local database
func TestImportGenesisBlockNoBlockInDB(t *testing.T) {
	assert := assert.New(t)
	defer func() {
		r := recover()
		assert.NotNil(r)
	}()
	defer monkey.UnpatchAll()
	monkey.Patch(tools.PathExists, func(string) bool {
		return false
	})
	chain := &blockchain.BlockChain{}
	monkey.Patch(blockchain.NewLatestStateBlockChain, func() (*blockchain.BlockChain, error) {
		return chain, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(chain), "GetCurrentBlock", func(*blockchain.BlockChain) *types.Block {
		return nil
	})
	monkey.Patch(GenerateGenesisBlock, func() (*GenesisBlock, error) {
		return nil, errors.New("failed to build genesis block")
	})
	ImportGenesisBlock()
}
