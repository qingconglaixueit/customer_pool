package customer_pool

import (
	"fmt"
	"github.com/pkg/errors"
	"sync"
)

type CustomerPool interface {
	GetObject() (interface{}, error)
	ReturnObject(interface{}) error
}
type ConnFunc func() (interface{}, error)

type MyConnPool struct {
	sync.Mutex
	maxConn     int
	minConn     int
	currentConn int
	pool        chan interface{}
	connFun     ConnFunc
	isClosed    bool
}

func NewMyConnPool(maxConn, minConn int, connFun ConnFunc) *MyConnPool {
	if minConn > maxConn || maxConn <= 0 {
		return nil
	}
	myPool := &MyConnPool{
		maxConn:     maxConn,
		minConn:     minConn,
		currentConn: 0,
		pool:        make(chan interface{}, maxConn),
		connFun:     connFun,
		isClosed:    false,
	}

	// 按照最小的了连接数进行创建连接，放入通道 pool 中
	for i := 0; i < minConn; i++ {
		con, err := connFun()
		if err != nil {
			continue
		}
		myPool.pool <- con
		myPool.currentConn++
	}

	return myPool
}

func (conn *MyConnPool) GetObject() (interface{}, error) {
	return conn.getObject()
}
func (conn *MyConnPool) getObject() (interface{}, error) {
	if conn.isClosed {
		return nil, errors.New("pool is closed")
	}
	// 从通道里面读，如果通道里面没有则新建一个
	select {
	case object := <-conn.pool:
		return object, nil
	default:

	}
	// 校验当前的连接数是否大于最大连接数，若是，则还是需要从 pool 中取
	// 此时使用 mutex 主要是为了锁 MyConnPool 的非通道的其他成员
	conn.Lock()
	if conn.currentConn >= conn.maxConn {
		object := <-conn.pool
		conn.Unlock()
		return object, nil
	}
	// 逻辑走到此处需要新建对象放到 pool 中
	object, err := conn.connFun()
	if err != nil {
		conn.Unlock()
		return nil, fmt.Errorf("create conn error : %+v", err)
	}
	// 当前 pool 已有连接数+1
	conn.currentConn++
	conn.Unlock()

	return object, nil
}

func (conn *MyConnPool) ReturnObject(object interface{}) error {
	return conn.returnObject(object)
}
func (conn *MyConnPool) returnObject(object interface{}) error {
	if conn.isClosed {
		return errors.New("pool is closed")
	}
	conn.pool <- object
	return nil
}
