package mmq

import (
	"time"
)

// mmq 通讯库
type mmq struct {
	server  *mmqServer
	client  *mmqClient   // 包含commCloud
	outChan chan Message // 业务通过recv接口获取message的channel
	// callback MessageCallback
}

// MessageCallback 消息回调接口
// type MessageCallback interface {
// 	MessageCallback(*Message)
// }

// func init() {
// 	// set tcp_retries2
// 	// _, err := util.ExecShell("sysctl -w net.ipv4.tcp_retries2=5")
// 	// if err != nil {
// 	// 	util.I().Printf("set tcp_retries2 error: %v", err)
// 	// 	return
// 	// }
// 	// logger.Debugf("set tcp_retries success")
// }

// newMmq 创建新的通讯库实例
func newMmq() *mmq {
	return &mmq{
		outChan: make(chan Message, 128),
	}
}

// input 内部接收到消息，转发给调用方
func (comm *mmq) input(msg *Message) {
	logger.Debugf("comm[%p].input() %v", comm, msg)
	select {
	case comm.outChan <- *msg:
	default:
		logger.Debugf("comm.input chan full")
	}
}

// Recv 调用方接收消息。用于单元测试调试。
func (comm *mmq) Recv() *Message {
	return comm.RecvWithTimeout(-1)
}

func (comm *mmq) TryRecv() *Message {
	return comm.RecvWithTimeout(0)
}

func (comm *mmq) RecvWithTimeout(timeout time.Duration) *Message {
	if timeout == -1 {
		msg := <-comm.outChan
		return &msg
	} else {
		select {
		case msg := <-comm.outChan:
			return &msg
		case <-time.After(timeout):
			logger.Infof("comm.Recv() timeout")
			return nil
		}
	}
}

// // RegisterMessageCallback 注册comm消息回调
// func (comm *mmq) RegisterMessageCallback(cb MessageCallback) {
// 	comm.callback = cb
// 	go func() {
// 		for m := range comm.outChan {
// 			cb.MessageCallback(&m)
// 		}
// 	}()
// }
