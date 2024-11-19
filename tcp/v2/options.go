package tcp

// Options 配置咯
type Options struct {
	Address  string
	IdleTime int
	MaxConn  int

	////简化配置，这两个参数不开放配置
	//InBuffSize  int //初始读缓存大小
	//OutBuffSize int //初始写缓存大小

	ReadMaxSize  int
	WriteMaxSize int
	HeartBeat    int
}

// Option ...
type Option func(*Options)

func newOptions(opt ...Option) *Options {
	opts := Options{}

	for _, o := range opt {
		o(&opts)
	}

	return &opts
}

// Address server 监听地址
func Address(a string) Option {
	return func(o *Options) {
		o.Address = a
	}
}

// 最大连接数量
func MaxConn(n int) Option {
	return func(o *Options) {
		o.MaxConn = n
	}
}

// IdleTime 最大空闲时间（秒）
func IdleTime(ms int) Option {
	return func(o *Options) {
		o.IdleTime = ms
	}
}

//// 读buff初始长度
//func InBuffSize(size int) Option {
//    return func(o *Options) {
//        o.InBuffSize = size
//    }
//}

//// 写buff初始长度
//func OutBuffSize(size int) Option {
//    return func(o *Options) {
//        o.OutBuffSize = size
//    }
//}

// 读数据长度上限
func ReadMaxSize(size int) Option {
	return func(o *Options) {
		o.ReadMaxSize = size
	}
}

// 写数据长度上限
func WriteMaxSize(size int) Option {
	return func(o *Options) {
		o.WriteMaxSize = size
	}
}

//心跳
func HeartBeat(ms int) Option {
	return func(o *Options) {
		o.HeartBeat = ms
	}
}
