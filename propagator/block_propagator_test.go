package propagator

import (
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/justitia/tools/events"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/p2p"
	pconf "github.com/DSiSc/p2p/config"
	"github.com/DSiSc/p2p/message"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func mockP2P() *p2p.P2P {
	config := &pconf.P2PConfig{}
	p, _ := p2p.NewP2P(config, events.NewEvent())
	return p
}

func TestNewBlockPropagator(t *testing.T) {
	assert := assert.New(t)
	blockOut := make(chan interface{})
	bp, err := NewBlockPropagator(mockP2P(), blockOut, events.NewEvent())
	assert.Nil(err)
	assert.NotNil(bp)
}

func TestBlockPropagator_Start(t *testing.T) {
	assert := assert.New(t)
	blockOut := make(chan interface{})
	bp, err := NewBlockPropagator(mockP2P(), blockOut, events.NewEvent())
	assert.Nil(err)
	assert.NotNil(bp)
	err = bp.Start()
	assert.Nil(err)
	assert.Equal(int32(1), bp.isRuning)
	bp.Stop()
}

func TestBlockPropagator_RecvMsg(t *testing.T) {
	defer monkey.UnpatchAll()
	assert := assert.New(t)
	blockOut := make(chan interface{})
	msgChan := make(chan *p2p.InternalMsg)
	p2pN := mockP2P()
	monkey.PatchInstanceMethod(reflect.TypeOf(p2pN), "MessageChan", func(this *p2p.P2P) <-chan *p2p.InternalMsg {
		return msgChan
	})

	bp, err := NewBlockPropagator(p2pN, blockOut, events.NewEvent())
	assert.Nil(err)
	assert.NotNil(bp)
	err = bp.Start()
	assert.Nil(err)
	assert.Equal(int32(1), bp.isRuning)

	block := &types.Block{}
	bmsg := &message.Block{
		Block: block,
	}
	iMsg := &p2p.InternalMsg{
		Payload: bmsg,
	}
	go func() {
		msgChan <- iMsg
	}()

	block1 := <-blockOut
	assert.Equal(block, block1)
	bp.Stop()
}

func TestBlockPropagator_BlockEventFunc(t *testing.T) {
	assert := assert.New(t)
	blockOut := make(chan interface{})
	bp, err := NewBlockPropagator(mockP2P(), blockOut, events.NewEvent())
	assert.Nil(err)
	assert.NotNil(bp)
	err = bp.Start()
	assert.Nil(err)
	assert.Equal(int32(1), bp.isRuning)

	block := &types.Block{}
	bp.eventCenter.Notify(types.EventBlockCommitted, block)
	bp.Stop()
}

func TestBlockPropagator_Stop(t *testing.T) {
	assert := assert.New(t)
	blockOut := make(chan interface{})
	bp, err := NewBlockPropagator(mockP2P(), blockOut, events.NewEvent())
	assert.Nil(err)
	assert.NotNil(bp)
	err = bp.Start()
	assert.Nil(err)
	assert.Equal(int32(1), bp.isRuning)
	bp.Stop()
	assert.Equal(int32(0), bp.isRuning)
}
