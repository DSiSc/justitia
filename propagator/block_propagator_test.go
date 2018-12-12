package propagator

import (
	"github.com/DSiSc/justitia/tools/events"
	"github.com/DSiSc/p2p"
	pconf "github.com/DSiSc/p2p/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func mockP2P() p2p.P2PAPI {
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
