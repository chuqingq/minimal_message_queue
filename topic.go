package mmq

import (
	"strings"
	"sync"
)

// commTopics TODO 当前不支持通配符
type commTopics struct {
	sync.Mutex
	topics map[string]interface{}
}

func (t *commTopics) add(topics string) {
	ts := strings.Split(topics, ",")
	t.Lock()
	for _, topic := range ts {
		t.topics[topic] = nil
	}
	t.Unlock()
}

func (t *commTopics) remove(topics string) {
	ts := strings.Split(topics, ",")
	t.Lock()
	for _, topic := range ts {
		delete(t.topics, topic)
	}
	t.Unlock()
}

func (t *commTopics) match(topic string) bool {
	t.Lock()
	_, ok := t.topics[topic]
	t.Unlock()
	return ok
}
