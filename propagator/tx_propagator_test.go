package propagator

import (
	"errors"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/justitia/tools/events"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/p2p"
	"github.com/DSiSc/p2p/message"
	"github.com/DSiSc/txpool"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func TestNewTxPropagator(t *testing.T) {
	assert := assert.New(t)
	txOut := make(chan interface{})
	tp, err := NewTxPropagator(mockP2P(), txOut, events.NewEvent())
	assert.Nil(err)
	assert.NotNil(tp)
}

func TestTxPropagator_Start(t *testing.T) {
	assert := assert.New(t)
	txOut := make(chan interface{})
	tp, err := NewTxPropagator(mockP2P(), txOut, events.NewEvent())
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

	tp, err := NewTxPropagator(p2pN, txOut, events.NewEvent())
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

func TestTxPropagator_Stop(t *testing.T) {
	assert := assert.New(t)
	txOut := make(chan interface{})
	tp, err := NewTxPropagator(mockP2P(), txOut, events.NewEvent())
	assert.Nil(err)
	assert.NotNil(tp)
	err = tp.Start()
	assert.Nil(err)
	assert.Equal(int32(1), tp.isRuning)
	tp.Stop()
	assert.Equal(int32(0), tp.isRuning)
}

func TestTxPropagator_TxEventFunc(t *testing.T) {
	assert := assert.New(t)
	defer monkey.UnpatchAll()
	msgChan := make(chan *p2p.InternalMsg)
	p2pN := mockP2P()
	monkey.PatchInstanceMethod(reflect.TypeOf(p2pN), "MessageChan", func(this *p2p.P2P) <-chan *p2p.InternalMsg {
		return msgChan
	})
	broadcastChan := make(chan message.Message)
	monkey.PatchInstanceMethod(reflect.TypeOf(p2pN), "BroadCast", func(this *p2p.P2P, msg message.Message) {
		broadcastChan <- msg
	})
	monkey.Patch(txpool.GetTxByHash, func(hash types.Hash) *types.Transaction {
		return &types.Transaction{}
	})
	txOut := make(chan interface{})
	tp, err := NewTxPropagator(p2pN, txOut, events.NewEvent())
	assert.Nil(err)
	assert.NotNil(tp)
	err = tp.Start()
	assert.Nil(err)
	assert.Equal(int32(1), tp.isRuning)

	tx := &types.Transaction{}
	tp.eventCenter.Notify(types.EventAddTxToTxPool, tx)

	tk := time.NewTicker(time.Second)
	select {
	case <-broadcastChan:
	case <-tk.C:
		assert.Nil(errors.New("failed to broadcast tx msg"))
	}
	tp.Stop()
}
