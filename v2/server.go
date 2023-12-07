package mmq

import (
	"strings"
	"sync"

	"github.com/chuqingq/minimal_message_queue/v2/tcp"
)

// Server 服务端
type Server struct {
	server         *tcp.Server
	MatchTopicFunc MatchTopicFunc

	peersMutex sync.Mutex
	peersMap   map[string]map[*tcp.Client]interface{}
}

func defaultMatchTopicFunc(pubtopic, subtopic string) bool {
	return pubtopic == subtopic
}

// NewServer 创建并启动服务端
func NewServer(addr string) *Server {
	s := tcp.NewServer(addr)
	server := &Server{
		server:         s,
		peersMap:       make(map[string]map[*tcp.Client]interface{}),
		MatchTopicFunc: defaultMatchTopicFunc,
	}
	s.SetOnPeerStateChange(func(c *tcp.Client, state tcp.ClientState) {
		if state == tcp.ClientConnected {
			logger.Debugf("peer[%p] connected", c)
		} else if state == tcp.ClientDisconnected {
			server.delPeer(c)
		}
	})
	s.SetOnMsgRecv(func(c *tcp.Client, msg []byte, err error) {
		// logger.Printf("server recv: %v", string(msg))
		if err != nil || msg == nil {
			return
		}
		cmd := &Command{}
		err = cmd.FromBytes(msg)
		if err != nil {
			return
		}
		logger.Printf("server recv cmd: %v", cmd)
		switch cmd.Cmd {
		case "subscribe":
			server.subscribe(c, strings.Split(cmd.Topic, ","))
		case "unsubscribe":
			server.unsubscribe(c, strings.Split(cmd.Topic, ","))
		case "publish":
			server.dispatchTopic(cmd.Topic, msg)
		default:
			logger.Warnf("server recv unknown cmd: %v", cmd.Cmd)
		}
	})
	return server
}

type MatchTopicFunc func(pubtopic, subtopic string) bool

// func (s *Server) SetCluster(addr string) *Server {
// 	// TODO
// 	return s
// }

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
func (s *Server) subscribe(c *tcp.Client, topics []string) {
	s.peersMutex.Lock()
	for _, topic := range topics {
		_, ok := s.peersMap[topic]
		if !ok {
			s.peersMap[topic] = make(map[*tcp.Client]interface{})
		}
		s.peersMap[topic][c] = nil
	}
	s.peersMutex.Unlock()
}

// peers.unsubscribe peer取消订阅topic
func (s *Server) unsubscribe(c *tcp.Client, topics []string) {
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
func (s *Server) delPeer(c *tcp.Client) {
	s.peersMutex.Lock()
	for _, peers := range s.peersMap {
		delete(peers, c)
	}
	s.peersMutex.Unlock()
}

// dispatchTopic 根据收到的topic分发给订阅的client。m是完整的command bytes
func (s *Server) dispatchTopic(topic string, m []byte) {
	logger.Debugf("server[%p].dispatchTopic(%v)", s, topic)
	s.peersMutex.Lock()
	for subtopic, peers := range s.peersMap {
		if s.MatchTopicFunc(topic, subtopic) {
			for peer := range peers {
				logger.Debugf("dispatchTopic(%v, %v) to peer: %p", topic, string(m), peer)
				peer.Send(m)
			}
		}
	}
	s.peersMutex.Unlock()
}
