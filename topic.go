package mmq

import (
	"strings"
	"sync"
)

// mmqTopics TODO 当前不支持通配符
type mmqTopics struct {
	sync.Mutex
	topics map[string]interface{}
}

func (t *mmqTopics) add(topics string) {
	ts := strings.Split(topics, ",")
	t.Lock()
	for _, topic := range ts {
		t.topics[topic] = nil
	}
	t.Unlock()
}

func (t *mmqTopics) remove(topics string) {
	ts := strings.Split(topics, ",")
	t.Lock()
	for _, topic := range ts {
		delete(t.topics, topic)
	}
	t.Unlock()
}

func (t *mmqTopics) match(topic string) bool {
	t.Lock()
	_, ok := t.topics[topic]
	t.Unlock()
	return ok
}
