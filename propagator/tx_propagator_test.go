package propagator

import (
	"github.com/stretchr/testify/assert"
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
