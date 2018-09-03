package events

import (
	"fmt"
	"github.com/DSiSc/craft/types"
	"testing"
)

func TestNewEvent(t *testing.T) {
	event := NewEvent()

	var EventSaveBlock types.EventType = 1
	var EventReplyTx types.EventType = 2

	var subscriber1 types.EventFunc = func(v interface{}) {
		fmt.Println("subscriber1 event func #1.")
	}

	var subscriber2 types.EventFunc = func(v interface{}) {
		fmt.Println("subscriber2 event func #2.")
	}

	fmt.Println("Subscribe...")
	// sub1 := event.Subscribe(EventReplyTx, subscriber1)
	sub1 := event.Subscribe(EventReplyTx, subscriber1)
	event.Subscribe(EventSaveBlock, subscriber2)
	event.Subscribe(EventReplyTx, subscriber2)

	fmt.Println("Notify...")
	event.Notify(EventSaveBlock, nil)

	fmt.Println("Notify All...")
	event.NotifyAll()

	event.UnSubscribe(EventReplyTx, sub1)
	fmt.Println("Notify All after unsubscribe sub1...")
	event.NotifyAll()

}
