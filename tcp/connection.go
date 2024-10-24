package tcp

import (
	"log"
	"net"
	"runtime/debug"
	"time"

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

	closeCb []func()

	option *Options
	//RBufSize int
	//WBufSize int
	//RMaxSize int
	//WMaxSize int

	CloseUtil
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

		CloseUtil: MakeCloseUtil(),
	}
}

func (ct *Connection) Id() int {
	return ct.id
}

func (ct *Connection) RegisterCloseCb(f func()) {
	ct.closeCb = append(ct.closeCb, f)
}

func (ct *Connection) Close() {
	ct.CloseUtil.Close(func() {
		//不直接关闭，有可能有数据要继续发送
		//等待写完毕。读不需要等待。写完之后再关闭
		//ct.conn.Close()
		//close(ct.rc)
		close(ct.wc)

		for _, cb := range ct.closeCb {
			if cb != nil {
				cb()
			}
		}
	})
}

//func (ct *Connection) ReadLoop() {
//	defer func() {
//		if err := recover(); err != nil {
//			debug.PrintStack()
//			panic(err)
//		}
//	}()
//
//	//lenValBuf := make([]byte, LEN_BYTE_CNT)
//	inBuf := make([]byte, ct.option.InBuffSize)
//	//使用bufio提高效率
//	reader := bufio.NewReaderSize(ct.conn, ct.option.InBuffSize)
//
//	for {
//		ct.setReadDeadline()
//
//		//leilog.Debug("ReadLoop\n")
//		//对上层的协定：前LEN_BYTE_CNT个字节代表协议长度（不包括长度部分）
//		n, err := io.ReadFull(reader, inBuf[:LEN_BYTE_CNT])
//		if n != LEN_BYTE_CNT || err != nil {
//			leilog.Error("id:%d, Read head error:%+v readlen:%v", ct.id, err, n)
//			ct.Close()
//			return
//		}
//
//		msgLen := int(decodeUint32(inBuf[:LEN_BYTE_CNT]))
//		//leilog.Debug("id:%d ReadLoop, msgLen:%d", ct.id, msgLen)
//		if msgLen < 0 {
//			continue
//		}
//
//		if ct.option.ReadMaxSize > 0 && msgLen > ct.option.ReadMaxSize {
//			leilog.Error("id:%d, Read msgLen:%d heat ReadMaxSize:%d", ct.id, msgLen, ct.option.ReadMaxSize)
//			ct.Close()
//			return
//		}
//
//		//msgLen协议长度，包括长度部分
//		if msgLen > len(inBuf) { //resize
//			inBuf = ct.resizeBuf(inBuf, LEN_BYTE_CNT, msgLen)
//		}
//
//		ct.setReadDeadline()
//
//		dn, err := io.ReadFull(reader, inBuf[LEN_BYTE_CNT:msgLen])
//		if dn != msgLen-LEN_BYTE_CNT || err != nil {
//			leilog.Error("id:%d Read data error:%+v, dn:%d, msgLen:%d", ct.id, err, dn, msgLen)
//			ct.Close()
//			return
//		}
//		//leilog.Debug("id:%d ReadLoop 2 msgLen:%d", ct.id, dn)
//
//		//换成ringbuf
//		msg, err := ct.proto.UnpackMsg(inBuf[LEN_BYTE_CNT:msgLen])
//		leilog.Debug("id:%d ReadLoop msg:%+v, msg_type:%+v, msgLen:%+v", ct.id, msg, reflect.TypeOf(msg), msgLen)
//		if err != nil {
//			ct.Close()
//			return
//		}
//		ct.processMsg(msg)
//	}
//
//	//leilog.Debug("Conn:%d ReadLoop done !", ct.id)
//}

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

		//pn是此次UnpackMsg处理掉的数据
		pn, err := ct.proto.UnpackMsg(ct, inBuf[:widx])
		if err != nil {
			log.Printf("id:%d Read data error:%+v, pn:%d\n", ct.id, err, pn)
			ct.Close()
			return
		}

		if pn > 0 { //需要移动数据
			if widx != len(inBuf) {
				copy(inBuf, inBuf[pn:widx])
			}
		}
		if widx == len(inBuf) {
			//这里会移动数据，所以上面作判断widx != len(inBuf)，避免两次拷贝数据
			inBuf = ct.resizeBuf(inBuf, pn, widx)
		}
		widx -= pn
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

//func (ct *Connection) processMsg(data []byte) {
//	ct.proto.Process(ct, data)
//	// select {
//	// case ct.rc <- msg:
//	// default:
//	// 	fmt.Printf("Connection:%d rc is full !!!\n", ct.id)
//	// }
//}

//func (ct *Connection) write(msg interface{}) {
//	//validLen := len(ct.outBuf) - ct.outCnt
//	validLen := len(ct.outBuf)
//
//	data, err := ct.proto.PackMsg(msg)
//	if err != nil {
//		leilog.Error("id:%d PackMsg error:%+v", ct.id, err)
//		return
//	}
//	if len(data) <= 0 {
//		return
//	}
//
//	dataLen := len(data)
//	if ct.option.WriteMaxSize > 0 && dataLen > ct.option.WriteMaxSize {
//		leilog.Error("id:%d Write dataLen:%d heat WriteMaxSize:%d", ct.id, dataLen, ct.option.WriteMaxSize)
//		ct.Close()
//		return
//	}
//
//	if validLen < LEN_BYTE_CNT+dataLen {
//		//ct.outBuf = ct.resizeBuf(ct.outBuf, ct.outCnt, ct.outCnt+LEN_BYTE_CNT+dataLen)
//		ct.outBuf = ct.resizeBuf(ct.outBuf, 0, LEN_BYTE_CNT+dataLen)
//	}
//
//	//encodeUint32(uint32(dataLen), ct.outBuf[ct.outCnt:])
//	//copy(ct.outBuf[ct.outCnt+LEN_BYTE_CNT:], data)
//	//ct.outCnt += dataLen + LEN_BYTE_CNT
//	encodeUint32(uint32(LEN_BYTE_CNT+dataLen), ct.outBuf)
//	copy(ct.outBuf[LEN_BYTE_CNT:], data)
//	wlen := dataLen + LEN_BYTE_CNT
//
//	//for ct.outCnt > 0 {
//	//	if n, err := ct.conn.Write(ct.outBuf[:ct.outCnt]); err != nil {
//	//		leilog.Error("Write error:%+v\n", err)
//	//		ct.Close()
//	//		return
//	//	} else {
//	//		//leilog.Debug("%d writed\n", n)
//	//		if n < ct.outCnt {
//	//			copy(ct.outBuf, ct.outBuf[n:ct.outCnt])
//	//		}
//	//		ct.outCnt -= n
//	//	}
//	//}
//
//	leilog.Debug("id:%d going to write:%+v", ct.id, ct.outBuf[:wlen])
//	wp := 0
//	for wp < wlen { //直到写完。
//		if n, err := ct.conn.Write(ct.outBuf[wp:wlen]); err != nil {
//			leilog.Error("id:%d Write error:%+v", ct.id, err)
//			ct.Close()
//			return
//		} else {
//			wp += n
//		}
//	}
//}

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
		case <-ct.CloseChan:
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

	if ct.IsClosed() {
		return
	}

	select {
	case ct.wc <- msg:
	default:
		log.Fatalf("Connection:%d wc is full !!!\n", ct.id)
	}
}

//func (ct *Connection) resizeBuf(oldBuff []byte, oldDataLen, minSize int) []byte {
//	l := len(oldBuff)
//	//leilog.Debug("resizeBuf old:%d\n", l)
//	for l < minSize {
//		if l < 1024 {
//			l <<= 1
//		} else {
//			l = int(float64(l) * 1.5)
//		}
//	}
//	//leilog.Debug("resizeBuf new:%d\n", l)
//	newBuff := make([]byte, l)
//	if oldDataLen > 0 {
//		copy(newBuff, oldBuff[:oldDataLen])
//	}
//	return newBuff
//}

func (ct *Connection) resizeBuf(oldBuff []byte, s, e int) []byte {
	l := len(oldBuff)
	//leilog.Debug("resizeBuf old:%d\n", l)
	if l < 10240 {
		l <<= 1
	} else {
		l = int(float64(l) * 1.5)
	}
	//leilog.Debug("resizeBuf new:%d\n", l)
	newBuff := make([]byte, l)
	if e > s {
		copy(newBuff, oldBuff[s:e])
	}
	return newBuff
}

//func (ct *Connection) expandBuf(oldBuff []byte) []byte {
//	l := len(oldBuff)
//	//leilog.Debug("resizeBuf old:%d\n", l)
//	if l < 1024 {
//		l <<= 1
//	} else {
//		l = int(float64(l) * 1.5)
//	}
//	//leilog.Debug("resizeBuf new:%d\n", l)
//	newBuff := make([]byte, l)
//	return newBuff
//}
