package tools

import (
	"encoding/hex"
	"github.com/DSiSc/craft/types"
)

func HexToAddress(s string) types.Address {
	return BytesToAddress(FromHex(s))
}

// FromHex returns the bytes represented by the hexadecimal string s.
// s may be prefixed with "0x".
func FromHex(s string) []byte {
	if len(s) > 1 {
		if s[0:2] == "0x" || s[0:2] == "0X" {
			s = s[2:]
		}
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	return Hex2Bytes(s)
}

// Hex2Bytes returns the bytes represented by the hexadecimal string str.
func Hex2Bytes(str string) []byte {
	h, _ := hex.DecodeString(str)
	return h
}

func BytesToAddress(b []byte) types.Address {
	var a types.Address
	SetBytes(b, &a)
	return a
}

func SetBytes(b []byte, a *types.Address) {
	if len(b) > len(a) {
		b = b[len(b)-types.AddressLength:]
	}
	copy(a[types.AddressLength-len(b):], b)
}
