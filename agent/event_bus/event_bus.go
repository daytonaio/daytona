// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package event_bus

import (
	"github.com/reactivex/rxgo/v2"
)

type EventPayload interface{}

type Event struct {
	Name    EventName
	Payload EventPayload
}

type EventName string

var events = make(chan rxgo.Item)

func Publish(event Event) error {
	go func() {
		events <- rxgo.Of(event)
	}()
	return nil
}

func Subscribe(unsubscribe chan bool) chan Event {
	ob := rxgo.FromChannel(events, rxgo.WithBackPressureStrategy(rxgo.Block))
	ch := make(chan Event)
	go func() {
		for {
			select {
			case <-unsubscribe:
				return
			case item := <-ob.Observe():
				ch <- item.V.(Event)
			}
		}
	}()
	return ch
}

func SubscribeWithFilter(unsubscribe chan bool, filter func(i Event) bool) chan Event {
	ob := rxgo.FromChannel(events, rxgo.WithBackPressureStrategy(rxgo.Block)).Filter(func(i interface{}) bool {
		if _, ok := i.(Event); !ok {
			return false
		}
		return filter(i.(Event))
	})

	ch := make(chan Event)
	go func() {
		for {
			select {
			case <-unsubscribe:
				return
			case item := <-ob.Observe():
				ch <- item.V.(Event)
			}
		}
	}()
	return ch
}
