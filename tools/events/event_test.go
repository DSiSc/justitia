package events

import (
	"fmt"
	"github.com/DSiSc/craft/types"
	"github.com/stretchr/testify/assert"
	"testing"
	)

func TestNewEvent(t *testing.T) {
	event := NewEvent()
	assert := assert.New(t)
	var EventSaveBlock types.EventType = 1
	var EventReplyTx types.EventType = 2

	var subscriber1 types.EventFunc = func(v interface{}) {
		fmt.Println("subscriber1 event func #1.")
	}

	var subscriber2 types.EventFunc = func(v interface{}) {
		fmt.Println("subscriber2 event func #2.")
	}

	fmt.Println("Subscribe...")
	sub1 := event.Subscribe(EventReplyTx, subscriber1)
	assert.NotNil(sub1)
	sub2 := event.Subscribe(EventReplyTx, subscriber1)
	assert.NotEqual(sub1, sub2)
	event.Subscribe(EventSaveBlock, subscriber2)
	event.Subscribe(EventReplyTx, subscriber2)
	fmt.Println("Notify...")
	err := event.Notify(EventSaveBlock, nil)
	assert.Nil(err)
	fmt.Println("Notify All...")
	errs := event.NotifyAll()
	assert.Equal(0, len(errs))
	err = event.UnSubscribe(EventReplyTx, sub1)
	assert.Nil(err)
	fmt.Println("Notify All after unsubscribe sub1...")
	errs = event.NotifyAll()
	assert.Equal(0, len(errs))
	fmt.Println("Notify All after unsubscribeall...")
	event.UnSubscribeAll()
	errs = event.NotifyAll()
	assert.Equal(0, len(errs))
	fmt.Println("Notify All after subscribeall...")
	event.Subscribe(EventReplyTx, subscriber1)
	event.NotifyAll()
}
