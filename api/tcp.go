package api

type TcpServerer interface {
	Start()
	Broadcast(interface{})
}

type TcpClienter interface {
	Id() int          //底层分配的id
	Reconnect() error //重连
	Write(interface{})
	Close()
	IsClosed() bool
	CloseC() <-chan struct{}
	RegisterCloseCb(func()) //注册链接关闭时的回调函数
}

type TcpConnectioner interface {
	Id() int
	Write(interface{})
	Close()
	IsClosed() bool
	RegisterCloseCb(func())
}
