package events

import (
	"errors"
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewEvent(t *testing.T) {
	event := NewEvent()
	assert := assert.New(t)
	var EventSaveBlock types.EventType = 1
	var EventReplyTx types.EventType = 2
	var EventNoneTx types.EventType = 3

	var subscriber1 types.EventFunc = func(v interface{}) {
		log.Info("TEST: subscriber1 event func #1.")
	}

	var subscriber2 types.EventFunc = func(v interface{}) {
		log.Info("TEST:subscriber2 event func #2.")
	}

	log.Info("TEST: Subscribe...")
	sub1 := event.Subscribe(EventReplyTx, subscriber1)
	assert.NotNil(sub1)
	sub2 := event.Subscribe(EventReplyTx, subscriber1)
	assert.NotEqual(sub1, sub2)
	event.Subscribe(EventSaveBlock, subscriber2)
	event.Subscribe(EventReplyTx, subscriber2)

	log.Info("TEST: Notify...")
	err := event.Notify(EventSaveBlock, nil)
	assert.Nil(err)
	err = event.Notify(EventSaveBlock, fmt.Errorf("callback failed"))
	assert.Nil(err)

	log.Info("TEST: Notify All...")
	errs := event.NotifyAll()
	assert.Equal(0, len(errs))

	log.Info("TEST: UnSubscribe who has subscribe...")
	err = event.UnSubscribe(EventReplyTx, sub1)
	assert.Nil(err)

	err = event.Notify(EventNoneTx, nil)
	errExpect := errors.New("event type not register")
	assert.Equal(errExpect, err)

	log.Info("TEST: Unsubscribe who has not subscrib...")
	err = event.UnSubscribe(EventNoneTx, nil)
	assert.Equal(err, errors.New("event type not exist"))

	log.Info("TEST: Notify All after unsubscribe sub1...")
	errs = event.NotifyAll()
	assert.Equal(0, len(errs))
	log.Info("TEST: Notify All after unsubscribe all...")
	event.UnSubscribeAll()
	errs = event.NotifyAll()
	assert.Equal(0, len(errs))
	log.Info("TEST: Notify All after subscribe all...")
	event.Subscribe(EventReplyTx, subscriber1)
	event.NotifyAll()

	// test nil eventFunc
	event.NotifySubscriber(nil, nil)
}

func TestEvent_Notify(t *testing.T) {
	event := NewEvent()
	assert := assert.New(t)
	var EventSaveBlock types.EventType = 1
	block := &types.Block{
		Header: &types.Header{
			Height: uint64(10),
		},
	}

	var subscriber1 types.EventFunc = func(v interface{}) {
		assert.NotNil(v)
		assert.Equal(block, v.(*types.Block))
		log.Info("TEST: subscriber1 event func #1.")
	}

	log.Info("TEST: Subscribe...")
	sub1 := event.Subscribe(EventSaveBlock, subscriber1)
	assert.NotNil(sub1)

	event.Notify(EventSaveBlock, block)
	time.Sleep(10 * time.Millisecond)
}
