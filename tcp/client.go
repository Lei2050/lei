package tcp

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"

	"github.com/Lei2050/lei-net/api"
)

var _ api.TcpClienter = &Client{}

type Client struct {
	conn *Connection
	addr *net.TCPAddr
}

var idmgr int // 加锁
var lock sync.RWMutex

func Dial(addr string, proto Protocoler, opts ...Option) (*Client, error) {
	option := newOptions(opts...)

	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	proto.SetOption(*option)

	log.Printf("%s connected.\n", addr)

	var id int
	lock.Lock()
	idmgr++
	id = idmgr
	lock.Unlock()
	connTcp := newConnection(conn, id, proto, option)

	go connTcp.ReadLoop()
	go connTcp.WriteLoop()
	proto.OnConnect(connTcp)

	c := &Client{conn: connTcp, addr: tcpAddr}
	if option.HeartBeat > 0 {
		go c.heartBeatLoop()
	}
	//connTcp.RegisterCloseCb(c.onConnClose)
	return c, nil
}

func (c *Client) heartBeatLoop() {
	if c.IsClosed() || c.conn.option.HeartBeat <= 100 { //100，防止配置错误
		return
	}

	ticker := time.NewTicker(time.Duration(c.conn.option.HeartBeat) * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			heartBeatMsg := c.conn.proto.HeartBeatMsg()
			if heartBeatMsg != nil {
				c.Write(heartBeatMsg)
			}
		case <-c.CloseC():
			return
		}
	}
}

//func (c *Client) onConnClose() {}

func (c *Client) RegisterCloseCb(f func()) {
	c.conn.RegisterCloseCb(f)
}

func (c *Client) Reconnect() error {
	if c.conn == nil {
		return errors.New("use Dial when first time to connect")
	}

	proto, option := c.conn.proto, c.conn.option

	if !c.conn.IsClosed() {
		c.conn.Close()
	}

	conn, err := net.DialTCP("tcp", nil, c.addr)
	if err != nil {
		return err
	}

	log.Printf("%+v reconnected.\n", c.addr)

	connTcp := newConnection(conn, 0, proto, option)
	c.conn = connTcp
	go connTcp.ReadLoop()
	go connTcp.WriteLoop()
	proto.OnConnect(connTcp)

	if option.HeartBeat > 0 {
		go c.heartBeatLoop()
	}
	//conn.RegisterCloseCb(c.onConnClose)

	return nil
}

func (c *Client) Write(msg interface{}) {
	c.conn.Write(msg)
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) IsClosed() bool {
	return c.conn.IsClosed()
}

func (c *Client) Id() int {
	return c.conn.Id()
}

func (c *Client) Addr() *net.TCPAddr {
	return c.addr
}

func (c *Client) CloseC() <-chan struct{} {
	return c.conn.CloseChan
}
