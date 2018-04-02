package socket

import (
	"fmt"
	"sync/atomic"
)

type EventType int32

const (
	Event_None EventType = iota
	Event_Connected
	Event_ConnectFailed
	Event_Accepted
	Event_AcceptFailed
	Event_Closed
	Event_Recv
	Event_Send
)

func (self EventType) String() string {
	switch self {
	case Event_Recv:
		return "recv"
	case Event_Send:
		return "send"
	case Event_Connected:
		return "connected"
	case Event_ConnectFailed:
		return "connectfailed"
	case Event_Accepted:
		return "accepted"
	case Event_AcceptFailed:
		return "acceptfailed"
	case Event_Closed:
		return "closed"
	}

	return fmt.Sprintf("unknown(%d)", self)
}

type Result int32

const (
	Result_OK            Result = iota
	Result_SocketError          // 网络错误
	Result_SocketTimeout        // Socket超时
	Result_PackageCrack         // 封包破损
	Result_CodecError
	Result_RequestClose // 请求关闭
	Result_NextChain
	Result_RPCTimeout
)

// 会话事件
type Event struct {
	UID int64

	Type EventType // 事件类型

	MsgID uint32      // 消息ID
	Msg   interface{} // 消息对象
	Data  []byte      // 消息序列化后的数据

	Tag         interface{} // 事件的连接, 一个处理流程后被Reset
	TransmitTag interface{} // 接收过程可以传递到发送过程, 不会被清空

	Ses       Session       // 会话
	ChainSend *HandlerChain // 发送handler override

	r Result // 出现错误, 将结束ChainCall

	chainid int64 // 所在链, 调试用
}

func (self *Event) Clone() *Event {
	c := &Event{
		UID:         self.UID,
		Type:        self.Type,
		MsgID:       self.MsgID,
		Msg:         self.Msg,
		Tag:         self.Tag,
		TransmitTag: self.TransmitTag,
		Ses:         self.Ses,
		ChainSend:   self.ChainSend,
		Data:        make([]byte, len(self.Data)),
	}

	copy(c.Data, self.Data)

	return c
}

func (self *Event) Result() Result {
	return self.r
}

func (self *Event) SetResult(r Result) {
	self.r = r
}

func (self *Event) PeerName() string {
	if self.Ses == nil {
		return ""
	}

	name := self.Ses.FromPeer().Name()
	if name != "" {
		return name
	}

	return self.Ses.FromPeer().Address()
}

func (self *Event) SessionID() int64 {
	if self.Ses == nil {
		return 0
	}

	return self.Ses.ID()
}

func (self *Event) MsgSize() int {
	if self.Data == nil {
		return 0
	}

	return len(self.Data)
}

func (self *Event) MsgString() string {
	if self.Msg == nil {
		return ""
	}

	if stringer, ok := self.Msg.(interface {
		String() string
	}); ok {
		return stringer.String()
	}

	return ""
}

func NewEvent(t EventType, s Session) *Event {
	self := &Event{
		Type: t,
		Ses:  s,
	}

	if EnableHandlerLog {
		self.UID = genSesEvUID()
	}

	return self
}

var evuid int64

func genSesEvUID() int64 {
	return atomic.AddInt64(&evuid, 1)
}
