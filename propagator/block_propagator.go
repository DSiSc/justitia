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

//BlockPropagator block message propagator
type BlockPropagator struct {
	p2p         p2p.P2PAPI
	blockOut    chan<- interface{}
	quitChan    chan interface{}
	eventCenter types.EventCenter
	subscribers map[types.EventType]types.Subscriber
	lock        sync.Mutex
	isRuning    int32
}

// NewBlockPropagator create a new NewBlockPropagator instance.
func NewBlockPropagator(p2p p2p.P2PAPI, blockOut chan<- interface{}, eventCenter types.EventCenter) (*BlockPropagator, error) {
	return &BlockPropagator{
		p2p:         p2p,
		blockOut:    blockOut,
		quitChan:    make(chan interface{}),
		eventCenter: eventCenter,
		subscribers: make(map[types.EventType]types.Subscriber),
		isRuning:    0,
	}, nil
}

// BlockEventFunc get a EventFunc that can be bound to event center
func (bp *BlockPropagator) BlockEventFunc(event interface{}) {
	switch event.(type) {
	case *types.Block:
		bp.broadCastBlock(event.(*types.Block))
	default:
		log.Warn("received a unknown block event")
	}
}

// broadcast message to p2p network
func (bp *BlockPropagator) broadCastBlock(block *types.Block) {
	bmsg := &message.Block{
		Block: block,
	}
	bp.p2p.BroadCast(bmsg)
}

// Start start propagator
func (bp *BlockPropagator) Start() error {
	bp.lock.Lock()
	defer bp.lock.Unlock()
	if bp.isRuning == 1 {
		log.Error("block propagator already started")
		return errors.New("block propagator already started")
	}
	bp.isRuning = 1

	bp.subscribers[types.EventBlockCommitted] = bp.eventCenter.Subscribe(types.EventBlockCommitted, bp.BlockEventFunc)
	bp.subscribers[types.EventBlockWritten] = bp.eventCenter.Subscribe(types.EventBlockWritten, bp.BlockEventFunc)
	go bp.recvHandler()
	return nil
}

// Stop start propagator
func (bp *BlockPropagator) Stop() {
	bp.lock.Lock()
	defer bp.lock.Unlock()
	if bp.isRuning == 0 {
		return
	}
	bp.isRuning = 0
	close(bp.quitChan)
	for eventType, subscriber := range bp.subscribers {
		delete(bp.subscribers, eventType)
		bp.eventCenter.UnSubscribe(eventType, subscriber)
	}
}

// receive handler will receive block from p2p, and send the block to gossip switch
func (bp *BlockPropagator) recvHandler() {
	for {
		select {
		case msg := <-bp.p2p.MessageChan():
			switch msg.Payload.(type) {
			case *message.Block:
				bmsg := msg.Payload.(*message.Block)
				log.Debug("received a block %x", common.HeaderHash(bmsg.Block))
				bp.blockOut <- bmsg.Block
			default:
				log.Error("received an invalid block message, message type: %v", msg.Payload.MsgType())
			}
		case <-bp.quitChan:
			log.Info("exit propagator receive handler, as propagator already stopped")
			return
		}
	}
}
