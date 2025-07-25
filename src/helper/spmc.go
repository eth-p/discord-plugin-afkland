package main

import (
	"context"
	"sync"
)

type SPMC[T any] struct {
	mu        sync.Mutex
	producer  <-chan T
	consumers map[chan T]struct{}
}

func NewSPMC[T any](producer <-chan T) *SPMC[T] {
	spmc := &SPMC[T]{
		producer:  producer,
		consumers: make(map[chan T]struct{}),
	}

	go spmc.run()
	return spmc
}

func (c *SPMC[T]) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for consumer := range c.consumers {
		close(consumer)
	}

	c.consumers = nil
	return nil
}

func (c *SPMC[T]) Consumer(buffer int) (<-chan T, context.CancelFunc) {
	consumer := make(chan T, buffer)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.consumers[consumer] = struct{}{}
	return consumer, func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		delete(c.consumers, consumer)
		close(consumer)
	}
}

func (c *SPMC[T]) run() {
	for {
		msg, ok := <-c.producer
		if !ok {
			break
		}

		c.broadcast(msg)
	}
}

// broadcast copies messages from the producer and sends them to the consumer.
func (c *SPMC[T]) broadcast(msg T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for consumer := range c.consumers {
		select {
		case consumer <- msg:
		default:
			// Oh well.
		}
	}
}
