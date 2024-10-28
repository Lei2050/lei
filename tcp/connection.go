package tcp

import (
	"log"
	"net"
	"runtime/debug"
	"time"

	cls "github.com/Lei2050/lei-utils/cls"

	"github.com/Lei2050/lei-net/api"
)

var _ api.TcpConnectioner = &Connection{}

type Connection struct {
	id int

	conn  *net.TCPConn
	proto Protocoler
	//rc    chan interface{}
	// inBuf  []byte
	// inCnt  int
	wc     chan interface{}
	outBuf []byte
	//outCnt int
	idle time.Duration

	option *Options
	//RBufSize int
	//WBufSize int
	//RMaxSize int
	//WMaxSize int

	cls.CloseUtil
}

func newConnection(conn *net.TCPConn, id int, proto Protocoler, option *Options) *Connection {
	return &Connection{
		conn:  conn,
		id:    id,
		proto: proto,
		// rc:    make(chan interface{}, 128),
		// inBuf:  make([]byte, 1024),
		// inCnt:  0,
		wc: make(chan interface{}, 128),
		//outBuf: make([]byte, option.OutBuffSize),
		outBuf: make([]byte, 256), //简化上层配置，这个参数不开放配置
		//outCnt: 0,
		idle:   time.Duration(option.IdleTime) * time.Millisecond,
		option: option,

		CloseUtil: cls.MakeCloseUtil(),
	}
}

func (ct *Connection) Id() int {
	return ct.id
}

func (ct *Connection) RegisterCloseCb(f func()) {
	ct.RegisterCloseCallback(f)
}

func (ct *Connection) Close() {
	ct.CloseUtil.Close(func() {
		//不直接关闭，有可能有数据要继续发送
		//等待写完毕。读不需要等待。写完之后再关闭
		//ct.conn.Close()
		//close(ct.rc)
		close(ct.wc)
	})
}

func (ct *Connection) ReadLoop() {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			panic(err)
		}
	}()

	//简化上层配置，这个参数不开放配置
	//inBuf := make([]byte, ct.option.InBuffSize)
	inBuf := make([]byte, 256)
	widx := 0

	for {
		ct.setReadDeadline()

		n, err := ct.conn.Read(inBuf[widx:])
		if err != nil {
			log.Printf("id:%d, Read head error:%+v readlen:%v\n", ct.id, err, n)
			ct.Close()
			return
		}

		widx += n

		totalPn := 0
		for {
			//pn是此次UnpackMsg处理掉的数据
			pn, err := ct.proto.UnpackMsg(ct, inBuf[totalPn:widx])
			if err != nil {
				log.Printf("id:%d Read data error:%+v, pn:%d\n", ct.id, err, pn)
				ct.Close()
				return
			}
			if pn <= 0 {
				break
			}
			totalPn += pn
		}

		if totalPn > 0 { //需要移动数据
			if widx != len(inBuf) {
				copy(inBuf, inBuf[totalPn:widx])
			}
		}
		if widx == len(inBuf) {
			//这里会移动数据，所以上面作判断widx != len(inBuf)，避免两次拷贝数据
			inBuf = ct.resizeBuf(inBuf, totalPn, widx)
		}
		widx -= totalPn
	}
}

func (ct *Connection) setReadDeadline() {
	if ct.idle > 0 {
		nowTime := time.Now()
		deadTime := nowTime.Add(ct.idle)
		err := ct.conn.SetReadDeadline(deadTime)
		if err != nil {
			log.Printf("SetDeadline err:%+v\n", err)
			ct.Close()
			return
		}
	}
}

func (ct *Connection) write(msg interface{}) {
	validLen := len(ct.outBuf)

	data, err := ct.proto.PackMsg(ct, msg)
	if err != nil {
		log.Printf("id:%d PackMsg error:%+v\n", ct.id, err)
		return
	}
	if len(data) <= 0 {
		return
	}

	dataLen := len(data)
	if ct.option.WriteMaxSize > 0 && dataLen > ct.option.WriteMaxSize {
		log.Printf("id:%d Write dataLen:%d heat WriteMaxSize:%d\n", ct.id, dataLen, ct.option.WriteMaxSize)
		ct.Close()
		return
	}

	if validLen < dataLen {
		ct.outBuf = ct.resizeBuf(ct.outBuf, 0, 0)
	}

	copy(ct.outBuf, data)

	log.Printf("id:%d going to write:%+v\n", ct.id, ct.outBuf[:dataLen])
	wp := 0
	for wp < dataLen { //直到写完。
		if n, err := ct.conn.Write(ct.outBuf[wp:dataLen]); err != nil {
			log.Printf("id:%d Write error:%+v\n", ct.id, err)
			ct.Close()
			return
		} else {
			wp += n
		}
	}
}

func (ct *Connection) WriteLoop() {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			panic(err)
		}
	}()

FOR:
	for {
		select {
		case msg := <-ct.wc:
			if msg != nil {
				ct.write(msg)
			} else {
				break FOR
			}
		case <-ct.CloseUtil.C():
			break FOR
		}
	}

	//关闭后，尽可能发送完队列中的数据
	for more := true; more; {
		select {
		case msg := <-ct.wc:
			if msg != nil {
				ct.write(msg)
			} else {
				more = false
			}
		default:
			more = false
		}
	}

	ct.conn.Close()
	log.Printf("Conn:%d WriteLoop done !", ct.id)
}

func (ct *Connection) Write(msg interface{}) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Connection write err:%+v", err)
			ct.Close()
		}
	}()

	select {
	case <-ct.C():
	case ct.wc <- msg:
	default:
		log.Fatalf("Connection:%d wc is full !!!\n", ct.id)
	}
}

func (ct *Connection) resizeBuf(oldBuff []byte, s, e int) []byte {
	l := len(oldBuff)
	if l < 10240 {
		l <<= 1
	} else {
		l = int(float64(l) * 1.5)
	}
	newBuff := make([]byte, l)
	if e > s {
		copy(newBuff, oldBuff[s:e])
	}
	return newBuff
}
