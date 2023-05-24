package mmq

import (
	"mmq/log"

	// "sync"
	"time"
)

func init() {
	// log.SetFlags(log.Flags() | log.Lshortfile)
	// log.SetOutput(os.Stdout)
}

// Comm 通讯库
type Comm struct {
	server   *commServer
	client   *commClient  // 包含commCloud
	outChan  chan Message // 业务通过recv接口获取message的channel
	callback MessageCallback
}

// MessageCallback 消息回调接口
type MessageCallback interface {
	MessageCallback(*Message)
}

func init() {
	// set tcp_retries2
	// _, err := util.ExecShell("sysctl -w net.ipv4.tcp_retries2=5")
	// if err != nil {
	// 	util.I().Printf("set tcp_retries2 error: %v", err)
	// 	return
	// }
	// log.Logger().Debugf("set tcp_retries success")
}

// NewComm 创建新的通讯库实例
func NewComm() *Comm {
	return &Comm{
		outChan: make(chan Message, 128),
	}
}

// input 内部接收到消息，转发给调用方
func (comm *Comm) input(msg *Message) {
	log.Logger().Debugf("comm[%p].input() %v", comm, msg)
	select {
	case comm.outChan <- *msg:
	default:
		log.Logger().Debugf("comm.input chan full")
	}
}

// Recv 调用方接收消息。用于单元测试调试。
func (comm *Comm) Recv() *Message {
	return comm.RecvWithTimeout(-1)
}

func (comm *Comm) TryRecv() *Message {
	return comm.RecvWithTimeout(0)
}

func (comm *Comm) RecvWithTimeout(timeout time.Duration) *Message {
	if timeout == -1 {
		select {
		case msg := <-comm.outChan:
			return &msg
		}
	} else {
		select {
		case msg := <-comm.outChan:
			return &msg
		case <-time.After(timeout):
			log.Logger().Infof("comm.Recv() timeout")
			return nil
		}
	}
}

// RegisterMessageCallback 注册comm消息回调
func (comm *Comm) RegisterMessageCallback(cb MessageCallback) {
	comm.callback = cb
	go func() {
		for m := range comm.outChan {
			cb.MessageCallback(&m)
		}
	}()
}
