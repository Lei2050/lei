package tcp

import (
	"log"
	"net"
	"sync"

	"github.com/Lei2050/lei-net/api"
)

var _ api.TcpServerer = &Server{}

type Server struct {
	ln           *net.TCPListener
	connections  map[int]*Connection
	connectIdCnt int
	proto        Protocoler
	opts         *Options

	sync.RWMutex
}

func NewServer(proto Protocoler, opts ...Option) (*Server, error) {
	options := newOptions(opts...)

	tcpAddr, err := net.ResolveTCPAddr("tcp4", options.Address)
	if err != nil {
		log.Printf("resolved address:%s error:%v\n", options.Address, err)
		return nil, err
	}
	ln, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Printf("can not listen on %+v, %v\n", tcpAddr, err)
		return nil, err
	}

	proto.SetOption(*options)

	log.Printf("listen on %+v\n", tcpAddr)

	return &Server{
		ln:           ln,
		connections:  make(map[int]*Connection),
		connectIdCnt: 0,
		proto:        proto,
		opts:         options,
	}, nil
}

func (s *Server) AddConnection(conn *net.TCPConn) *Connection {
	s.connectIdCnt++
	connection := newConnection(conn, s.connectIdCnt, s.proto, s.opts)
	s.Lock()
	s.connections[connection.id] = connection
	s.Unlock()
	return connection
}

//func (s *Server) onConnectionClose(id int) {
//	delete(s.connections, id)
//	leilog.Debug("onConnectionClose:%d current num of conns:%d!\n", id, len(s.connections))
//}

func (s *Server) Start() {
	for {
		conn, err := s.ln.AcceptTCP()
		if err != nil {
			log.Printf("accept error:%v\n", err)
			return
		}

		if s.opts.MaxConn > 0 {
			s.RLock()
			connNum := len(s.connections)
			s.RUnlock()
			if connNum >= s.opts.MaxConn {
				log.Printf("connection full ! maxConn:%v conn:%v", s.opts.MaxConn, connNum)
				conn.Close()
				continue
			}
		}

		connTcp := s.AddConnection(conn)
		id := connTcp.Id()
		connTcp.RegisterCloseCb(func(connId int) func() {
			id := connId
			return func() {
				s.Lock()
				delete(s.connections, id)
				s.Unlock()
				log.Printf("onConnectionClose:%d current num of conns:%d!", id, len(s.connections))
			}
		}(id))
		log.Printf("in comming connection:%s id:%d, local_addr:%s conn_tcp:%d",
			connTcp.conn.RemoteAddr().String(), id, connTcp.conn.LocalAddr().String(), s.connectIdCnt)
		go connTcp.ReadLoop()
		go connTcp.WriteLoop()
		s.proto.OnConnect(connTcp)
	}
}

func (s *Server) Broadcast(msg interface{}) {
	s.RLock()
	for _, v := range s.connections {
		v.Write(msg)
	}
	s.RUnlock()
}
