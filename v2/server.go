package mmq

import (
	"sync"

	"github.com/chuqingq/minimal_message_queue/v2/tcpjson"
	sjson "github.com/chuqingq/simple-json"
)

// Server 服务端
type Server struct {
	server         *tcpjson.Server
	MatchTopicFunc MatchTopicFunc

	peersMutex sync.Mutex
	peersMap   map[string]map[*tcpjson.Client]interface{}
}

func defaultMatchTopicFunc(pubtopic, subtopic string) bool {
	return pubtopic == subtopic
}

// NewServer 创建并启动服务端
func NewServer(addr string) *Server {
	s := tcpjson.NewServer(addr)
	server := &Server{
		server:         s,
		peersMap:       make(map[string]map[*tcpjson.Client]interface{}),
		MatchTopicFunc: defaultMatchTopicFunc,
	}
	s.SetOnPeerStateChange(func(c *tcpjson.Client, state tcpjson.ClientState) {
		if state == tcpjson.ClientConnected {
			logger.Debugf("peer[%p] connected", c)
		} else if state == tcpjson.ClientDisconnected {
			server.delPeer(c)
		}
	})
	s.SetOnMsgRecv(func(c *tcpjson.Client, msg []byte, err error) {
		if err != nil || msg == nil {
			return
		}
		m, err := sjson.FromBytes(msg)
		if err != nil {
			return
		}

		switch m.Get("cmd").MustString() {
		case "subscribe":
			server.subscribe(c, m.Get("topics").MustStringArray())
		case "unsubscribe":
			server.unsubscribe(c, m.Get("topics").MustStringArray())
		case "publish":
			server.dispatchTopic(m.Get("topic").MustString(), m.Get("msg"))
		default:
			logger.Warnf("server recv unknown cmd: %v", m.Get("cmd").MustString())
		}
	})
	return server
}

type MatchTopicFunc func(pubtopic, subtopic string) bool

func (s *Server) SetCluster(addr string) *Server {
	// TODO
	return s
}

func (s *Server) SetMatchTopicFunc(match MatchTopicFunc) *Server {
	s.MatchTopicFunc = match
	return s
}

func (s *Server) Start() error {
	return s.server.Start()
}

func (s *Server) Stop() {
	s.server.Stop()
}

// peers.subscribe peer订阅topic
func (s *Server) subscribe(c *tcpjson.Client, topics []string) {
	s.peersMutex.Lock()
	for _, topic := range topics {
		_, ok := s.peersMap[topic]
		if !ok {
			s.peersMap[topic] = make(map[*tcpjson.Client]interface{})
		}
		s.peersMap[topic][c] = nil
	}
	s.peersMutex.Unlock()
}

// peers.unsubscribe peer取消订阅topic
func (s *Server) unsubscribe(c *tcpjson.Client, topics []string) {
	s.peersMutex.Lock()
	for _, topic := range topics {
		_, ok := s.peersMap[topic]
		if !ok {
			continue
		}
		delete(s.peersMap[topic], c)
	}
	s.peersMutex.Unlock()
}

// peers.delPeer() 删除peer
func (s *Server) delPeer(c *tcpjson.Client) {
	s.peersMutex.Lock()
	for _, peers := range s.peersMap {
		delete(peers, c)
	}
	s.peersMutex.Unlock()
}

// dispatchTopic 根据收到的topic分发给订阅的client
func (s *Server) dispatchTopic(topic string, m *sjson.Json) {
	logger.Debugf("server[%p].dispatchTopic(%v)", s, topic)
	msg := &sjson.Json{}
	msg.Set("topic", topic)
	msg.Set("msg", m)
	s.peersMutex.Lock()
	for subtopic, peers := range s.peersMap {
		if s.MatchTopicFunc(topic, subtopic) {
			for peer := range peers {
				logger.Debugf("dispatchTopic(%v, %v) to peer: %p", topic, msg, peer)
				peer.Send(msg.ToBytes())
			}
		}
	}
	s.peersMutex.Unlock()
}
