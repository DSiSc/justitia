package tools

import (
	"github.com/DSiSc/craft/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

var addr = types.Address{
	0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33,
	0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d}

func TestHexToAddress(t *testing.T) {
	assert := assert.New(t)
	address := HexToAddress("333c3310824b7c685133f2bedb2ca4b8b4df633d")
	assert.NotNil(address)
	assert.Equal(addr, address)
}

func TestBytesToAddress(t *testing.T) {
	assert := assert.New(t)
	var addr1 = []byte{
		0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33,
		0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d}
	address := BytesToAddress(addr1)
	assert.Equal(addr, address)
}

func TestFromHex(t *testing.T) {
	assert := assert.New(t)
	oddStr := "0x12345"
	var oddExcept = []byte{
		0x1, 0x23, 0x45,
	}
	bytes := FromHex(oddStr)
	assert.Equal(oddExcept, bytes)
	evenStr := "0x123456"
	var evenExcept = []byte{
		0x12, 0x34, 0x56,
	}
	bytes = FromHex(evenStr)
	assert.Equal(evenExcept, bytes)
}

func TestHex2Bytes(t *testing.T) {
	assert := assert.New(t)
	oddStr := "12"
	var oddExcept = []byte{0x12}
	bytes := Hex2Bytes(oddStr)
	assert.Equal(oddExcept, bytes)
}
