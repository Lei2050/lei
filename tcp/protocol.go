package tcp

import (
	"github.com/Lei2050/lei-net/api"
)

type Protocoler interface {
	//解包
	//int - 此次解包处理了多少数据，剩下的留待下次解包处理
	//error - 如果不为空，视为发生致命错误，则断开连接
	UnpackMsg(api.TcpConnectioner, []byte) (int, error)
	//打包
	PackMsg(api.TcpConnectioner, interface{}) ([]byte, error)
	//连接建立时
	OnConnect(api.TcpConnectioner)
	//获取心跳包。用于自动心跳，可能返回nil。这是业务层心跳，交由上层生成。
	HeartBeatMsg() interface{}
	//设置参数，不用该配置参数的则忽略即可
	SetOption(Options)
}
