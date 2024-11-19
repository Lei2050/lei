package tcp

import (
	"io"

	"github.com/Lei2050/lei-net/api"
)

type Protocoler interface {
	//上层接口读取数据
	//error - 如果不为空，视为发生致命错误，则断开连接
	Read(api.TcpConnectioner, io.Reader) error
	//上层接口写数据
	Write(io.Writer, any) error
	//连接建立时
	OnConnect(api.TcpConnectioner)
	//获取心跳包。用于自动心跳，可能返回nil。这是业务层心跳，交由上层生成。
	HeartBeatMsg() any
	//设置参数，不用该配置参数的则忽略即可
	SetOption(Options)
}
