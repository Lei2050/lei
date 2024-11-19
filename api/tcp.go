package api

type TcpServerer interface {
	Start()
	Broadcast(any) int
}

type TcpClienter interface {
	Id() int          //底层分配的id
	Reconnect() error //重连
	Write(any)
	Close()
	IsClosed() bool
	CloseC() <-chan struct{}
	RegisterCloseCb(func()) //注册链接关闭时的回调函数
}

type TcpConnectioner interface {
	Id() int
	Write(any)
	Close()
	IsClosed() bool
	RegisterCloseCb(func())
}
