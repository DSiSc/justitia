package common

import (
	"bytes"
	gconf "github.com/DSiSc/craft/config"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/rlp"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/crypto-suite/crypto/sha3"
	"hash"
	"math/big"
)

type NodeType int

const (
	UnknownNode   NodeType = iota // UnknownNode --> 0, Unknown node type
	ConsensusNode                 // ConsensusNode --> 1, Consensus function node
	FullNode                      // FullNode --> 2, Full node
	LightNode                     // LightNode --> 3, Light node with simply function
	MaxNodeType                   // MaxNodeType is the boundary of node type
)

const (
	InvalidInt  int    = 255
	BlankString string = ""
)

func HashAlg() hash.Hash {
	var alg string
	if value, ok := gconf.GlobalConfig.Load(gconf.HashAlgName); ok {
		alg = value.(string)
	} else {
		alg = "SHA256"
	}
	return sha3.NewHashByAlgName(alg)
}

func rlpHash(x interface{}) (h types.Hash) {
	hw := HashAlg()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

// Hash hashes the RLP encoding of tx.
// It uniquely identifies the transaction.
func TxHash(tx *types.Transaction) types.Hash {
	if hash := tx.Hash.Load(); hash != nil {
		return hash.(types.Hash)
	}
	v := rlpHash(tx)
	tx.Hash.Store(v)
	return v
}

func HeaderHash(block *types.Block) types.Hash {
	//var defaultHash types.Hash
	if !(block.HeaderHash == types.Hash{}) {
		var hash types.Hash
		copy(hash[:], block.HeaderHash[:])
		return hash
	}
	return rlpHash(block.Header)
}

func CopyBytes(b []byte) (copiedBytes []byte) {
	var temp []byte
	if bytes.Equal(b, temp) {
		log.Error("src byte is nil, please confirm.")
		return
	}
	copiedBytes = make([]byte, len(b))
	copy(copiedBytes, b)
	return
}

// New a transaction
func newTransaction(nonce uint64, to *types.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, from *types.Address) *types.Transaction {
	if len(data) > 0 {
		data = CopyBytes(data)
	}
	d := types.TxData{
		AccountNonce: nonce,
		Recipient:    to,
		From:         from,
		Payload:      data,
		Amount:       new(big.Int),
		GasLimit:     gasLimit,
		Price:        new(big.Int),
		V:            new(big.Int),
		R:            new(big.Int),
		S:            new(big.Int),
	}
	if amount != nil {
		d.Amount.Set(amount)
	}
	if gasPrice != nil {
		d.Price.Set(gasPrice)
	}

	return &types.Transaction{Data: d}
}

func NewTransaction(nonce uint64, to types.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, from types.Address) *types.Transaction {
	return newTransaction(nonce, &to, amount, gasLimit, gasPrice, data, &from)
}

type MsgType uint8

const (
	MsgNull               MsgType = iota // meaningless msg type
	MsgBlockCommitSuccess                //  block commit success
	MsgBlockCommitFailed                 //  block commit failed
	MsgBlockVerifyFailed                 //  block verify failed
	MsgNodeServiceStopped                //  stop node service
	MsgRoundRunFailed                    //  round run failed with some reasons
	MsgToConsensusFailed                 //  failed to consensus
	MsgChangeMaster                      //  change master
	MsgOnline                            //  node online
	MsgBlockWithoutTx                    // block without transaction
)

type SysConfig struct {
	LogLevel log.Level
	LogPath  string
	LogStyle string
}
