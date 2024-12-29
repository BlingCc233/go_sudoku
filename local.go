package sudoku_go

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sudoku_go/sudoku"
	"time"
)

type LsLocal struct {
	ListenAddr *net.TCPAddr
	RemoteAddr *net.TCPAddr
}

// 新建一个本地端
// 本地端的职责是:
// 1. 监听来自本机浏览器的代理请求
// 2. 转发前加密数据
// 3. 转发socket数据到墙外代理服务端
// 4. 把服务端返回的数据转发给用户的浏览器
func NewLsLocal(listenAddr, remoteAddr string) (*LsLocal, error) {
	structListenAddr, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	structRemoteAddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
	if err != nil {
		return nil, err
	}
	return &LsLocal{
		ListenAddr: structListenAddr,
		RemoteAddr: structRemoteAddr,
	}, nil
}

// 本地端启动监听，接收来自本机浏览器的连接
func (local *LsLocal) Listen(didListen func(listenAddr net.Addr)) error {
	trafficStat()
	return ListenSecureTCP(local.ListenAddr, sudoku.DefaultRequest.Code, local.handleConn, didListen)
}

func (local *LsLocal) handleConn(userConn *SecureTCPConn) {
	defer userConn.Close()
	proxyServer, err := DialTCPSecure(local.RemoteAddr, sudoku.DefaultRequest.Code)
	log.Print("Connected to Server : ", local.RemoteAddr)
	if err != nil {
		log.Println(err)
		return
	}
	defer proxyServer.Close()

	// Encode traffic received from the local client and forward it to the remote proxy server
	go func() {

		// Create a sudoku request
		sudokuReq := sudoku.DefaultRequest

		// 在Encode之前以sudoku作为header，但不Encode
		if _, err := sudokuReq.WriteTo(proxyServer); err != nil {
			log.Print(err)
		}

		//err := userConn.DirectEncodeCopy(proxyServer)
		err := userConn.EncodeCopy(proxyServer)
		if err != nil {
			log.Print(err)
			userConn.Close()
			proxyServer.Close()
		}
	}()

	// 接收sudoku响应
	sudokuResp := &sudoku.Response{}
	if _, err := sudokuResp.ReadFrom(proxyServer); err != nil {
		log.Print(err)
	}

	if sudokuResp.Status != sudoku.StatusOK {
		log.Print("sudoku status not ok")
		return
	}

	// Decode traffic received from the remote proxy server and send it back to the local client
	//err = proxyServer.DecodeCopy(userConn)
	err = proxyServer.DirectDEcodeCopy(userConn)
	if err != nil {
		log.Print(err)
		// 在 copy 的过程中可能会存在网络超时等 error 被 return，只要有一个发生了错误就退出本次工作
		userConn.Close()
		proxyServer.Close()
	}
}

func trafficStat() {
	printTicker := time.NewTicker(10 * time.Second)
	statTicker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case _ = <-printTicker.C:
				printTrafficStat()
			case _ = <-statTicker.C:
				sendTrafficStat()
			}
		}
	}()
}

func printTrafficStat() {
	RxLock.RLock()
	TxLock.RLock()
	fmt.Printf("Receive: %dM Send: %dM\n", Rx/1024/1024, Tx/1024/1024)
	RxLock.RUnlock()
	TxLock.RUnlock()
}

func sendTrafficStat() {
	addr, err := net.ResolveUnixAddr("unix", "stat_main")
	if err != nil {
		//log.Printf("Failed to resolve: %v\n", err)
		return
	}

	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		//log.Printf("Failed to dial: %v\n", err)
		return
	}
	defer conn.Close()

	bs := make([]byte, 8)
	TxLock.RLock()
	binary.LittleEndian.PutUint64(bs, Tx)
	TxLock.RUnlock()
	conn.Write(bs)
	RxLock.RLock()
	binary.LittleEndian.PutUint64(bs, Rx)
	RxLock.RUnlock()
	conn.Write(bs)

	if err != nil {
		log.Printf("Failed to report stat: %v\n", err)
		return
	}
}
