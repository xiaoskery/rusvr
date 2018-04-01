package socket

import (
	"fmt"
	"reflect"
)

type EventHandler interface {
	Call(*Event)
}

var EnableHandlerLog bool

// HandlerName 显示handler的名称
func HandlerName(h EventHandler) string {
	if h == nil {
		return "nil"
	}

	return reflect.TypeOf(h).Elem().Name()
}

func HandlerString(h EventHandler) string {
	if sg, ok := h.(fmt.Stringer); ok {
		return sg.String()
	} else {
		return HandlerName(h)
	}
}

func HandlerLog(h EventHandler, ev *Event) {
	if EnableHandlerLog {
	}
}

func HandlerChainCall(hlist []EventHandler, ev *Event) {
	for _, h := range hlist {

		HandlerLog(h, ev)

		h.Call(ev)

		if ev.Result() != Result_OK {
			break
		}
	}
}
