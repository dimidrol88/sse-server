package internal

import (
	"log"
	"sync"
)

type Broker struct {
	connects map[string]*Connect
	mu       sync.RWMutex
}

func NewBroker() *Broker {
	return &Broker{
		connects: make(map[string]*Connect),
	}
}

func (b *Broker) Add(c *Connect) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.connects[c.Id] = c
	log.Printf("Connect %s подписан!", c.Id)
}

func (b *Broker) Remove(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if c, ok := b.connects[id]; ok {
		close(c.DoneCh)
		delete(b.connects, id)
		log.Printf("Connect %s отписан!", id)
	}
}

func (b *Broker) Notify(msg *Message) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, connect := range b.connects {
		select {
		case connect.MessageCh <- msg:
		default:
			log.Printf("Connect %s messagechannel full, skipping", connect.Id)
		}
	}
}

func (b *Broker) Ping(connect *Connect) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	select {
	case connect.PingCh <- "{}":
	default:
		log.Printf("Connect %s ping channel full, skipping", connect.Id)
	}
}
