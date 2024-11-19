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

	wc     chan any
	outBuf []byte
	idle   time.Duration

	option *Options

	cls.CloseUtil
}

func newConnection(conn *net.TCPConn, id int, proto Protocoler, option *Options) *Connection {
	return &Connection{
		conn:   conn,
		id:     id,
		proto:  proto,
		wc:     make(chan any, 128),
		outBuf: make([]byte, 256), //简化上层配置，这个参数不开放配置
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
		//close(ct.wc)
	})
}

func (ct *Connection) ReadLoop() {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			panic(err)
		}
	}()

	for {
		ct.setReadDeadline()
		err := ct.proto.Read(ct, ct.conn)
		if err != nil {
			log.Printf("id:%d Read data error:%+v\n", ct.id, err)
			ct.Close()
			return
		}
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

func (ct *Connection) write(msg any) {
	err := ct.proto.Write(ct.conn, msg)
	if err != nil {
		log.Printf("id:%d PackMsg error:%+v\n", ct.id, err)
		return
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

func (ct *Connection) Write(msg any) {
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
