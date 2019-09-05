package events

import (
	"errors"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"sync"
)

type Event struct {
	m           sync.RWMutex
	Subscribers map[types.EventType]map[types.Subscriber]types.EventFunc
}

func NewEvent() types.EventCenter {
	return &Event{
		Subscribers: make(map[types.EventType]map[types.Subscriber]types.EventFunc),
	}
}

//  adds a new subscriber to Event.
func (e *Event) Subscribe(eventType types.EventType, eventFunc types.EventFunc) types.Subscriber {
	e.m.Lock()
	defer e.m.Unlock()

	sub := make(chan interface{})
	_, ok := e.Subscribers[eventType]
	if !ok {
		e.Subscribers[eventType] = make(map[types.Subscriber]types.EventFunc)
	}
	e.Subscribers[eventType][sub] = eventFunc

	return sub
}

// UnSubscribe removes the specified subscriber
func (e *Event) UnSubscribe(eventType types.EventType, subscriber types.Subscriber) (err error) {
	e.m.Lock()
	defer e.m.Unlock()

	subEvent, ok := e.Subscribers[eventType]
	if !ok {
		err = errors.New("event type not exist")
		return
	}

	delete(subEvent, subscriber)
	close(subscriber)

	return
}

// Notify subscribers that Subscribe specified event
func (e *Event) Notify(eventType types.EventType, value interface{}) (err error) {

	e.m.RLock()
	defer e.m.RUnlock()

	subs, ok := e.Subscribers[eventType]
	if !ok {
		err = errors.New("event type not register")
		return
	}

	switch value.(type) {
	case error:
		log.Error("Receive errors is [%v].", value)
	}
	log.Info("Receive eventType is [%d].", eventType)

	for _, event := range subs {
		go e.NotifySubscriber(event, value)
	}
	return nil
}

func (e *Event) NotifySubscriber(eventFunc types.EventFunc, value interface{}) {
	if eventFunc == nil {
		return
	}

	// invoke subscriber event func
	eventFunc(value)

}

//Notify all event subscribers
func (e *Event) NotifyAll() (errs []error) {
	e.m.RLock()
	defer e.m.RUnlock()

	for eventType, _ := range e.Subscribers {
		if err := e.Notify(eventType, nil); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

// unsubscribe all event and subscriber elegant
func (e *Event) UnSubscribeAll() {
	e.m.Lock()
	defer e.m.Unlock()
	for eventtype, _ := range e.Subscribers {
		subs, ok := e.Subscribers[eventtype]
		if !ok {
			continue
		}
		for subscriber, _ := range subs {
			delete(subs, subscriber)
			close(subscriber)
		}
	}
	// TODO: open it when txswitch and blkswith stop complete
	//e.Subscribers = make(map[types.EventType]map[types.Subscriber]types.EventFunc)
	return
}
