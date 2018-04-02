package socket

import (
	"io"
	"net"
	"sync"
	"time"
)

// Session 会话
type Session interface {
	// 发包
	Send(interface{})

	// 直接发送封包
	//RawSend(*Event)

	// 断开
	Close()

	// 标示ID
	ID() int64

	// 归属端
	FromPeer() Peer

	// 将一个用户数据保存在session
	SetTag(tag interface{})

	// 取出与session关联的用户数据
	Tag() interface{}

	// 取原始连接net.Conn
	RawConn() interface{}
}

type socketSession struct {
	OnClose func() // 关闭函数回调

	id int64

	p Peer

	endSync sync.WaitGroup

	needNotifyWrite bool // 是否需要通知写线程关闭

	sendList *eventList

	conn net.Conn

	tag interface{}

	tagGuard sync.RWMutex

	readChain *HandlerChain

	writeChain *HandlerChain
}

func (self *socketSession) RawConn() interface{} {
	return self.conn
}

func (self *socketSession) Tag() interface{} {
	self.tagGuard.RLock()
	defer self.tagGuard.RUnlock()
	return self.tag
}
func (self *socketSession) SetTag(tag interface{}) {
	self.tagGuard.Lock()
	self.tag = tag
	self.tagGuard.Unlock()
}

func (self *socketSession) ID() int64 {
	return self.id
}

func (self *socketSession) SetID(id int64) {
	self.id = id
}

func (self *socketSession) FromPeer() Peer {
	return self.p
}

func (self *socketSession) DataSource() io.ReadWriter {
	return self.conn
}

func (self *socketSession) Close() {
	self.sendList.Add(nil)
}

func (self *socketSession) Send(data interface{}) {

}

func (self *socketSession) recvThread() {
	for {
		ev := NewEvent(Event_Recv, self)

		read, _ := self.FromPeer().(SocketOptions).SocketDeadline()

		if read != 0 {
			self.conn.SetReadDeadline(time.Now().Add(read))
		}

		self.readChain.Call(ev)

		if ev.Result() != Result_OK {
			goto onClose
		}

		continue

	onClose:
		break
	}

	if self.needNotifyWrite {
		self.Close()
	}

	// 通知接收线程ok
	self.endSync.Done()
}

// 发送线程
func (self *socketSession) sendThread() {
	for {
		// 写超时
		_, write := self.FromPeer().(SocketOptions).SocketDeadline()

		if write != 0 {
			self.conn.SetWriteDeadline(time.Now().Add(write))
		}

		writeList, willExit := self.sendList.Pick()

		// 写队列
		for _, ev := range writeList {
			// 发送链处理: encode等操作
			if ev.ChainSend != nil {
				ev.ChainSend.Call(ev)
			}

			if ev.Result() != Result_OK {
				willExit = true
			}

			// 发送日志
			//MsgLog(ev)

			// 写链处理
			self.writeChain.Call(ev)

			if ev.Result() != Result_OK {
				willExit = true
			}
		}

		//if err := self.conn.Flush(); err != nil {
		//	willExit = true
		//}

		if willExit {
			goto exitsendloop
		}
	}
exitsendloop:
	// 不需要读线程再次通知写线程
	self.needNotifyWrite = false

	// 关闭socket,触发读错误, 结束读循环
	self.conn.Close()

	// 通知发送线程ok
	self.endSync.Done()
}

func (self *socketSession) run() {
	// 布置接收和发送2个任务
	self.endSync.Add(2)

	go func() {
		// 等待2个任务结束
		self.endSync.Wait()

		// 在这里断开session与逻辑的所有关系
		if self.OnClose != nil {
			self.OnClose()
		}
	}()

	// 接收线程
	go self.recvThread()

	// 发送线程
	go self.sendThread()
}

func newSession(conn net.Conn, p Peer) *socketSession {
	p.(interface {
		Apply(conn net.Conn)
	}).Apply(conn)

	self := &socketSession{
		conn:            conn,
		p:               p,
		needNotifyWrite: true,
		//sendList:        NewPacketList(),
	}

	self.readChain = p.CreateChainRead()

	self.writeChain = p.CreateChainWrite()

	return self
}
