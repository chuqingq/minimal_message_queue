package mmq

// 这个文件是针对CommInterface的一个适配层

// import (
// 	"log"
// )

// // StartServer() 是CommInterface的一部分，暂时不动
// func (comm *Comm) StartServer(serverID, addr string, key, cert, ca []byte) {
// 	comm.Start(addr, key, cert, ca)
// }

// // StopServer() 是CommInterface的一部分，被web调用，暂时不动接口
// func (comm *Comm) StopServer(serverID string) {
// 	comm.Stop()
// }

// // SendData() 从from发出消息。可以指定to，也可以指定topic
// // 说明：这个API是暴露给web的，暂时不动
// func (comm *Comm) SendData(m *Message) {
// 	log.Logger().Debugf("SendData(msg: %v)", m)
// 	from, _ := (*m)["from"].(string)
// 	// to->SendTo; topic->Publish
// 	to, _ := (*m)["to"].(string)
// 	if to != "" {
// 		comm.SendTo(from, to, m)
// 		return
// 	}
// 	topic, _ := (*m)["topic"].(string)
// 	if topic != "" {
// 		comm.Publish(from, topic, m)
// 		return
// 	}
// 	// other
// 	log.Logger().Debugf("SendData() invalid message: %v", m)
// }
