package mmq

import (
	"strings"
	"sync"
)

// Topics TODO 当前不支持通配符
type Topics struct {
	sync.Mutex
	topics map[string]interface{}
}

func (t *Topics) add(topics string) {
	ts := strings.Split(topics, ",")
	t.Lock()
	for _, topic := range ts {
		t.topics[topic] = nil
	}
	t.Unlock()
}

func (t *Topics) remove(topics string) {
	ts := strings.Split(topics, ",")
	t.Lock()
	for _, topic := range ts {
		delete(t.topics, topic)
	}
	t.Unlock()
}

func (t *Topics) match(topic string) bool {
	t.Lock()
	_, ok := t.topics[topic]
	t.Unlock()
	return ok
}
