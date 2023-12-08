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
		case CmdSubscribe:
			server.subscribe(c, strings.Split(cmd.Topic, ","))
		case CmdUnsubscribe:
			server.unsubscribe(c, strings.Split(cmd.Topic, ","))
		case CmdPublish:
			server.dispatchTopic(cmd.Topic, msg)
		default:
			logger.Warnf("server recv unknown cmd: %v", cmd.Cmd)
		}
	})
	return server
}

// MatchTopicFunc topic匹配函数
type MatchTopicFunc func(pubtopic, subtopic string) bool

// SetMatchTopicFunc 设置topic匹配函数。可以定制
func (s *Server) SetMatchTopicFunc(match MatchTopicFunc) *Server {
	s.MatchTopicFunc = match
	return s
}

// Start 启动服务端
func (s *Server) Start() error {
	return s.server.Start()
}

// Stop 停止服务端
func (s *Server) Stop() {
	s.server.Stop()
}

// subscribe peer订阅topic
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

// unsubscribe peer取消订阅topic
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

// delPeer 删除peer
func (s *Server) delPeer(c *tcp.Client) {
	s.peersMutex.Lock()
	for _, peers := range s.peersMap {
		delete(peers, c)
	}
	s.peersMutex.Unlock()
}

// dispatchTopic 根据收到的topic分发给订阅的peer。m是完整的command bytes
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
