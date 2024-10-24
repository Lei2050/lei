# lei/network
这是一个轻量tcp网络库，能够快速建立tcp服务器和客户端。  
主要是使用golang官方库net包实现。每一个连接都使用两个goroutine分别进行读和写的操作。  
业务层需要提供一个实现了接口`protocoler`的协议处理器。
><code>type Protocoler interface {
	//解包
	//int - 此次解包处理了多少数据，剩下的留待下次解包处理
	//error - 如果不为空，视为发生致命错误，则断开连接
	UnpackMsg(lei.TcpConnectioner, []byte) (int, error)
	//打包
	PackMsg(lei.TcpConnectioner, interface{}) ([]byte, error)
	//连接建立时
	OnConnect(lei.TcpConnectioner)
	//获取心跳包。用于自动心跳，可能返回nil。这是业务层心跳，交由上层生成。
	HeartBeatMsg() interface{}
	//设置参数，不用该配置参数的则忽略即可
	SetOption(Options)
}</code>  

其中：  
`UnpackMsg`：将二进制数据解包（解码）成自己业务成的数据结构。  
`PackMsg`：将自己业务层的数据结构打包（编码）成二进制数据。  
`OnConnect`：连接建立事件。  

参考示例：**lei/chat_server**和**lei/chat_client**。  
