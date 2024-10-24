package tcp

import "sync"

type CloseUtil struct {
	CloseChan chan struct{}
	once      sync.Once
}

func MakeCloseUtil() CloseUtil {
	return CloseUtil{
		CloseChan: make(chan struct{}),
	}
}

func (c *CloseUtil) IsClosed() bool {
	select {
	case <-c.CloseChan:
		return true
	default:
	}
	return false
}

func (c *CloseUtil) Close(f func()) {
	c.once.Do(func() {
		close(c.CloseChan)
		if f != nil {
			f()
		}
	})
}
