/*
____  ___   _____ ___________
\   \/  /  /     \\__    ___/
 \     /  /  \ /  \ |    |
 /     \ /    Y    \|    |
/___/\  \\____|__  /|____|
      \_/        \/
createTime:2022/9/11
createUser:Administrator
*/
package main

import (
	"fmt"
	"log"
	"net"
	"test/customer_pool"
	"time"
)

type PoolTest struct {
	Conn *net.UDPConn
}

var myPool *customer_pool.MyConnPool

func init() {
	myPool = customer_pool.NewMyConnPool(3, 1, func() (interface{}, error) {
		return connectUdp()
	})
	if myPool == nil {
		log.Panicln("NewMyConnPool error")
		return
	}
	log.Println("myPool == ", myPool)
}

func main() {
	for i := 0; i < 10; i++ {
		msg := fmt.Sprintf("send udp data is %d", i)
		go SendMsg(msg)
	}

	time.Sleep(10 * time.Second)
}

// 初始化链接类
func connectUdp() (*PoolTest, error) {
	// 创建一个 udp 句柄
	log.Println(">>>>> 创建一个 udp 句柄 ... ")
	// 连接服务器
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 9998,
	})

	if err != nil {
		log.Println("Connect to udp server failed,err:", err)
		return nil, err
	}
	log.Printf("<<<<<< new udp connect %+v", conn)
	return &PoolTest{Conn: conn}, nil
}
func SendMsg(msg string) {
	var client *PoolTest
	// 从连接池中获取一个实例
	obj, err := myPool.GetObject()
	if err != nil {
		log.Printf("GetObject error : %+v", err)
		return
	}
	client = obj.(*PoolTest)
	// 调用需要的方法
	log.Println(msg)
	client.SendMsg([]byte(msg))
	// 交还连接池
	myPool.ReturnObject(client)
}
func (this *PoolTest) SendMsg(data []byte) {
	_, err := this.Conn.Write(data)
	if err != nil {
		log.Printf("write udp error : %+v", err)
		return
	}
	//读取回信
	result := make([]byte, 1024)
	//if err := this.Conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
	//	log.Printf("SetReadDeadlineerror: %+v", err)
	//	return
	//}

	n, remoteAddr, err := this.Conn.ReadFromUDP(result)
	if err != nil {
		log.Printf("read udp server msg error [data:%s] :%+v", string(data), err)
		return
	}
	log.Printf("Recived msg from %s, data:%s \n", remoteAddr, string(result[:n]))
}
