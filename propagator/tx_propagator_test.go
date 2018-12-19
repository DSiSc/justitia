package propagator

import (
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/p2p"
	"github.com/DSiSc/p2p/message"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestNewTxPropagator(t *testing.T) {
	assert := assert.New(t)
	txOut := make(chan interface{})
	tp, err := NewTxPropagator(mockP2P(), txOut)
	assert.Nil(err)
	assert.NotNil(tp)
}

func TestTxPropagator_Start(t *testing.T) {
	assert := assert.New(t)
	txOut := make(chan interface{})
	tp, err := NewTxPropagator(mockP2P(), txOut)
	assert.Nil(err)
	assert.NotNil(tp)
	err = tp.Start()
	assert.Nil(err)
	assert.Equal(int32(1), tp.isRuning)
	tp.Stop()
}

func TestTxPropagator_RecvTx(t *testing.T) {
	defer monkey.UnpatchAll()
	assert := assert.New(t)
	txOut := make(chan interface{})
	msgChan := make(chan *p2p.InternalMsg)
	p2pN := mockP2P()
	monkey.PatchInstanceMethod(reflect.TypeOf(p2pN), "MessageChan", func(this *p2p.P2P) <-chan *p2p.InternalMsg {
		return msgChan
	})

	tp, err := NewTxPropagator(p2pN, txOut)
	assert.Nil(err)
	assert.NotNil(tp)
	err = tp.Start()
	assert.Nil(err)
	assert.Equal(int32(1), tp.isRuning)

	tx := &types.Transaction{}
	tmsg := &message.Transaction{
		Tx: tx,
	}
	iMsg := &p2p.InternalMsg{
		Payload: tmsg,
	}
	go func() {
		msgChan <- iMsg
	}()
	tx1 := <-txOut
	assert.Equal(tx, tx1)
	tp.Stop()
}

func TestTxPropagator_TxSwitchOutPutFunc(t *testing.T) {
	assert := assert.New(t)
	txOut := make(chan interface{})
	tp, err := NewTxPropagator(mockP2P(), txOut)
	assert.Nil(err)
	assert.NotNil(tp)
	err = tp.Start()
	assert.Nil(err)
	assert.Equal(int32(1), tp.isRuning)

	outFunc := tp.TxSwitchOutPutFunc()
	outFunc(&types.Transaction{})
	outFunc(&types.Block{})
	tp.Stop()
}

func TestTxPropagator_Stop(t *testing.T) {
	assert := assert.New(t)
	txOut := make(chan interface{})
	tp, err := NewTxPropagator(mockP2P(), txOut)
	assert.Nil(err)
	assert.NotNil(tp)
	err = tp.Start()
	assert.Nil(err)
	assert.Equal(int32(1), tp.isRuning)
	tp.Stop()
	assert.Equal(int32(0), tp.isRuning)
}
