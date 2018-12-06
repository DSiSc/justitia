package propagator

import (
	"errors"
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/gossipswitch/port"
	"github.com/DSiSc/justitia/common"
	"github.com/DSiSc/p2p"
	"github.com/DSiSc/p2p/message"
	"reflect"
	"sync"
)

//TxPropagator transaction message propagator
type TxPropagator struct {
	p2p      p2p.P2PAPI
	txOut    chan<- interface{}
	quitChan chan interface{}
	isRuning int32
	lock     sync.Mutex
}

// NewBlockPropagator create a new NewBlockPropagator instance.
func NewTxPropagator(p2p p2p.P2PAPI, txOut chan<- interface{}) (*TxPropagator, error) {
	return &TxPropagator{
		p2p:      p2p,
		txOut:    txOut,
		quitChan: make(chan interface{}),
		isRuning: 0,
	}, nil
}

// TxSwitchOutPutFunc get a OutPutFunc that can be bound to gossip switch
func (tp *TxPropagator) TxSwitchOutPutFunc() port.OutPutFunc {
	return func(msg interface{}) error {
		switch msg.(type) {
		case *types.Transaction:
			tp.broadCastTx(msg.(*types.Transaction))
		default:
			return fmt.Errorf("unknown transaction message, message type: %v", reflect.TypeOf(msg))
		}
		return nil
	}
}

// broadcast tx message to p2p network
func (tp *TxPropagator) broadCastTx(tx *types.Transaction) {
	tmsg := &message.Transaction{
		Tx: tx,
	}
	tp.p2p.BroadCast(tmsg)
}

// Start start propagator
func (tp *TxPropagator) Start() error {
	tp.lock.Lock()
	defer tp.lock.Unlock()
	if tp.isRuning == 1 {
		log.Error("transaction propagator already started")
		return errors.New("transaction propagator already started")
	}
	tp.isRuning = 1
	go tp.recvHandler()
	return nil
}

// Stop start propagator
func (tp *TxPropagator) Stop() {
	tp.lock.Lock()
	defer tp.lock.Unlock()
	if tp.isRuning == 0 {
		return
	}
	tp.isRuning = 0
	close(tp.quitChan)
}

// receive handler will receive block from p2p, and send the block to gossip switch
func (tp *TxPropagator) recvHandler() {
	for {
		select {
		case msg := <-tp.p2p.MessageChan():
			switch msg.Payload.(type) {
			case *message.Transaction:
				txmsg := msg.Payload.(*message.Transaction)
				log.Debug("received a transaction %x", common.TxHash(txmsg.Tx))
				tp.txOut <- txmsg.Tx
			default:
				log.Error("received an invalid transaction message, message type: %v", msg.Payload.MsgType())
			}
		case <-tp.quitChan:
			log.Info("exit propagator receive handler, as propagator already stopped")
			return
		}
	}
}
