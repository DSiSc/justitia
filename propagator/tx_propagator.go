package propagator

import (
	"errors"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/justitia/common"
	"github.com/DSiSc/p2p"
	"github.com/DSiSc/p2p/message"
	"sync"
)

//TxPropagator transaction message propagator
type TxPropagator struct {
	p2p         p2p.P2PAPI
	txOut       chan<- interface{}
	quitChan    chan interface{}
	isRuning    int32
	lock        sync.Mutex
	eventCenter types.EventCenter
	subscribers map[types.EventType]types.Subscriber
}

// NewBlockPropagator create a new NewBlockPropagator instance.
func NewTxPropagator(p2p p2p.P2PAPI, txOut chan<- interface{}, eventCenter types.EventCenter) (*TxPropagator, error) {
	return &TxPropagator{
		p2p:         p2p,
		txOut:       txOut,
		quitChan:    make(chan interface{}),
		isRuning:    0,
		eventCenter: eventCenter,
		subscribers: make(map[types.EventType]types.Subscriber),
	}, nil
}

// BlockEventFunc get a EventFunc that can be bound to event center
func (tp *TxPropagator) TxEventFunc(event interface{}) {
	switch event.(type) {
	case *types.Transaction:
		tp.broadCastTx(event.(*types.Transaction))
	default:
		log.Warn("received a unknown transaction event")
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

	tp.subscribers[types.EventAddTxToTxPool] = tp.eventCenter.Subscribe(types.EventAddTxToTxPool, tp.TxEventFunc)

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
	for eventType, subscriber := range tp.subscribers {
		delete(tp.subscribers, eventType)
		tp.eventCenter.UnSubscribe(eventType, subscriber)
	}
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
