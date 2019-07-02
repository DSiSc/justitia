package common

import (
	"bytes"
	"github.com/DSiSc/craft/types"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

var defaultAddress = types.Address{
	0xb2, 0x6f, 0x2b, 0x34, 0x2a, 0xab, 0x24, 0xbc, 0xf6, 0x3e,
	0xa2, 0x18, 0xc6, 0xa9, 0x27, 0x4d, 0x30, 0xab, 0x9a, 0x15,
}

var emptyTransaction = NewTransaction(
	0, defaultAddress, big.NewInt(0), 0, big.NewInt(0), defaultAddress[:10], defaultAddress)

func TestTxHash(t *testing.T) {
	assert := assert.New(t)
	hash := TxHash(emptyTransaction)
	expect := types.Hash{0xba, 0xbf, 0x2e, 0xd4, 0x6b, 0x36, 0x9b, 0x43, 0x72, 0x15, 0x0, 0xb9, 0xdf, 0xbb, 0x38, 0xca, 0x41, 0x1e, 0x3e, 0x73, 0x44, 0x48, 0xbe, 0x7e, 0x53, 0x39, 0xf6, 0xe, 0xba, 0xa2, 0x38, 0x4f}
	assert.Equal(expect, hash)
}

func TestCopyBytes(t *testing.T) {
	assert := assert.New(t)
	var except []byte
	copyBytes := CopyBytes(except)
	assert.Equal(copyBytes, except)
	except = []byte{
		0xf3, 0x2c, 0x26, 0xa4, 0xee, 0x93, 0x3d, 0x72, 0x80, 0x40, 0xa5, 0xb1, 0x1b, 0x8d, 0xd3, 0x94,
		0x31, 0x83, 0xec, 0x50, 0x36, 0xfd, 0xac, 0xf9, 0x35, 0x2, 0x1, 0x1a, 0xab, 0x95, 0xb8, 0xb5,
	}
	copyBytes = CopyBytes(except)
	assert.Equal(copyBytes, except)
}

var MockHash = types.Hash{
	0x1d, 0xcf, 0x7, 0xba, 0xfc, 0x42, 0xb0, 0x8d, 0xfd, 0x23, 0x9c, 0x45, 0xa4, 0xb9, 0x38, 0xd,
	0x8d, 0xfe, 0x5d, 0x6f, 0xa7, 0xdb, 0xd5, 0x50, 0xc9, 0x25, 0xb1, 0xb3, 0x4, 0xdc, 0xc5, 0x1c,
}

var MockBlockHash = types.Hash{
	0xaf, 0x4e, 0x5b, 0xa3, 0x16, 0x97, 0x74, 0x6a, 0x26, 0x9d, 0x9b, 0x9e, 0xf1, 0x9d, 0xa8, 0xb3,
	0xf9, 0x32, 0x68, 0x16, 0xf4, 0x73, 0xd4, 0xb3, 0x6a, 0xaf, 0x2d, 0x6d, 0xfa, 0x82, 0xd9, 0x89,
}

var MockHeaderHash = types.Hash{
	0xcc, 0x88, 0x1c, 0x28, 0x30, 0x38, 0x50, 0x46, 0x2c, 0xcb, 0xae, 0xe5, 0xa4, 0x88, 0x85, 0x75,
	0xdf, 0xae, 0xd7, 0xd3, 0x39, 0x17, 0x9a, 0xfc, 0x9c, 0x4, 0x5e, 0xcd, 0x98, 0x8a, 0x39, 0xdd,
}

func MockBlock() *types.Block {
	return &types.Block{
		Header: &types.Header{
			ChainID:       1,
			PrevBlockHash: MockHash,
			StateRoot:     MockHash,
			TxRoot:        MockHash,
			ReceiptsRoot:  MockHash,
			Height:        1,
			Timestamp:     uint64(time.Date(2018, time.August, 28, 0, 0, 0, 0, time.UTC).Unix()),
		},
		Transactions: make([]*types.Transaction, 0),
	}
}

func TestHeaderHash(t *testing.T) {
	assert := assert.New(t)
	block := MockBlock()

	var tmp types.Hash
	assert.True(bytes.Equal(tmp[:], block.HeaderHash[:]))
	headerHash := HeaderHash(block)
	exceptHeaderHash := types.Hash{0x70, 0x22, 0xc, 0x81, 0xf1, 0xf1, 0xa9, 0x6c, 0x10, 0x73, 0x3a, 0x63, 0x82, 0x15, 0x5f, 0xac, 0x39, 0xdf, 0xdc, 0x3d, 0xb1, 0x92, 0x91, 0xd, 0xc2, 0x76, 0xab, 0xa3, 0x6b, 0xa2, 0xf4, 0x10}
	assert.Equal(exceptHeaderHash, headerHash)

	block.HeaderHash = HeaderHash(block)
	headerHash = HeaderHash(block)
	assert.Equal(exceptHeaderHash, headerHash)

	newBlock := &types.Block{
		Header: &types.Header{
			Height: 1,
		},
		HeaderHash: MockBlockHash,
	}
	ttt := HeaderHash(newBlock)
	assert.Equal(MockBlockHash, ttt)

	newBlock = &types.Block{}
	assert.Equal(newBlock.HeaderHash, types.Hash{})
	ttt = HeaderHash(newBlock)
	assert.NotEqual(types.Hash{}, ttt)
}
