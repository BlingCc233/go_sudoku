package sudoku_go

import (
	"encoding/binary"
	"log"
	"net"
	"sudoku_go/sudoku"
)

type LsServer struct {
	ListenAddr *net.TCPAddr
}

// 新建一个服务端
// 服务端的职责是:
// 1. 监听来自本地代理客户端的请求
// 2. 解密本地代理客户端请求的数据，解析 SOCKS5 协议，连接用户浏览器真正想要连接的远程服务器
// 3. 转发用户浏览器真正想要连接的远程服务器返回的数据的加密后的内容到本地代理客户端
func NewLsServer(listenAddr string) (*LsServer, error) {
	structListenAddr, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	return &LsServer{
		ListenAddr: structListenAddr,
	}, nil

}

// 运行服务端并且监听来自本地代理客户端的请求
func (lsServer *LsServer) Listen(didListen func(listenAddr net.Addr)) error {
	return ListenSecureTCP(lsServer.ListenAddr, sudoku.DefaultRequest.Code, lsServer.handleConn, didListen)
}

// 解 SOCKS5 协议
// https://www.ietf.org/rfc/rfc1928.txt
func (lsServer *LsServer) handleConn(localConn *SecureTCPConn) {
	defer localConn.Close()

	// 构建sudoku响应
	sudokuResp := &sudoku.Response{
		TlsObf:  [3]byte{0x16, 0x03, 0x03},
		Version: sudoku.Version1,
		Status:  sudoku.StatusOK,
		Code:    0x01,
	}

	// 首先处理sudoku请求
	sudokuReq := &sudoku.Request{}
	if _, err := sudokuReq.ReadFrom(localConn); err != nil {
		log.Printf("Failed to read sudoku request: %v", err)
		return
	}

	maskCode := sudokuReq.Code
	if maskCode != 0x01 && maskCode != 0x00 {
		sudokuResp.Status = sudoku.StatusBadRequest
		sudokuResp.WriteTo(localConn)
		return
	}

	// 返回sudoku响应
	sudokuResp.WriteTo(localConn)

	//Version identifier/method selection request from client
	buf := make([]byte, 3)
	/**
	   The localConn connects to the dstServer, and sends a ver
	   identifier/method selection message:
		          +----+----------+----------+
		          |VER | NMETHODS | METHODS  |
		          +----+----------+----------+
		          | 1  |    1     | 1 to 255 |
		          +----+----------+----------+
	   The VER field is set to X'05' for this ver of the protocol.  The
	   NMETHODS field contains the number of method identifier octets that
	   appear in the METHODS field.
	*/
	// 第一个字段VER代表Socks的版本，Socks5默认为0x05，其固定长度为1个字节
	n, err := localConn.DecodeRead(buf)
	//n, err := localConn.Read(buf)
	log.Printf("first request: %v", buf)
	// 只支持版本5
	if err != nil || buf[0] != 0x05 || n < 3 {
		log.Printf("Can't handle the request: %v %v", buf, err)
		return
	}
	/**
	   The dstServer selects from one of the methods given in METHODS, and
	   sends a METHOD selection message:

		          +----+--------+
		          |VER | METHOD |
		          +----+--------+
		          | 1  |   1    |
		          +----+--------+
	*/
	// 不需要验证，直接验证通过
	//localConn.EncodeWrite([]byte{0x05, 0x00})
	localConn.Write([]byte{0x05, 0x00})

	/**
	  +----+-----+-------+------+----------+----------+
	  |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	  +----+-----+-------+------+----------+----------+
	  | 1  |  1  | X'00' |  1   | Variable |    2     |
	  +----+-----+-------+------+----------+----------+
	*/

	//Connect request
	buf = make([]byte, 4)
	n, err = localConn.DecodeRead(buf)
	//n, err = localConn.Read(buf)
	log.Printf("first half of second request: %v", buf)
	if err != nil || n < 4 {
		log.Printf("Can't handle the request: %v %v", buf, err)
		return
	}

	// CMD代表客户端请求的类型，值长度也是1个字节，有三种类型
	// CONNECT X'01'
	if buf[1] != 0x01 {
		// 目前只支持 CONNECT
		log.Println("Can't handle command: ", buf[1])
		return
	}

	var addrLength int
	addrType := buf[3]
	switch addrType {
	case 0x01:
		//	IP V4 address: X'01'
		addrLength = net.IPv4len
	case 0x03:
		//	DOMAINNAME: X'03'
		buf = make([]byte, 1)
		localConn.DecodeRead(buf)
		//localConn.Read(buf)
		addrLength = int(buf[0])
	case 0x04:
		//	IP V6 address: X'04'
		addrLength = net.IPv6len
	default:
		log.Println("Invalid address type: ", addrType)
		return
	}
	addrLength += 2

	buf = make([]byte, addrLength)
	n, err = localConn.DecodeRead(buf)
	//n, err = localConn.Read(buf)
	log.Printf("second half of second request: %v", buf)
	log.Printf(string(buf[0 : n-2]))
	if err != nil || n < addrLength {
		log.Printf("Error reading address, address length: %v, buffer:%v  %v n: %v ", addrLength, buf, err, n)
		return
	}

	var dIP []byte
	// aType 代表请求的远程服务器地址类型，值长度1个字节，有三种类型
	switch addrType {
	case 0x01:
		//	IP V4 address: X'01'
		dIP = buf[0:net.IPv4len]
	case 0x03:
		//	DOMAINNAME: X'03'
		ipAddr, err := net.ResolveIPAddr("ip", string(buf[0:n-2]))
		if err != nil {
			log.Println("Can't resolve IP: ", err)
			return
		}
		dIP = ipAddr.IP
	case 0x04:
		//	IP V6 address: X'04'
		dIP = buf[0:net.IPv6len]
	default:
		log.Println("Can't handle address: ", buf[3])
		return
	}
	dPort := buf[n-2:]
	dstAddr := &net.TCPAddr{
		IP:   dIP,
		Port: int(binary.BigEndian.Uint16(dPort)),
	}

	// 连接真正的远程服务
	dstServer, err := net.DialTCP("tcp", nil, dstAddr)
	if err != nil {
		log.Println("Error occurred when connecting to real server : ", dstAddr)
		return
	} else {
		log.Println("Connected to real server : ", dstAddr)
		defer dstServer.Close()
		// Conn被关闭时直接清除所有数据 不管没有发送的数据
		dstServer.SetLinger(0)

		// 响应客户端连接成功
		/**
		  +----+-----+-------+------+----------+----------+
		  |VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
		  +----+-----+-------+------+----------+----------+
		  | 1  |  1  | X'00' |  1   | Variable |    2     |
		  +----+-----+-------+------+----------+----------+
		*/
		// 响应客户端连接成功
		//localConn.EncodeWrite([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		localConn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	}

	// 进行转发
	// Decode traffic received from the client side proxy and send it to the real server
	go func() {
		//err := localConn.DirectDEcodeCopy(dstServer)
		err := localConn.DecodeCopy(dstServer)
		if err != nil {
			log.Print(err)
			localConn.Close()
			dstServer.Close()
		}
	}()

	// Encode response from the real server adn send it back to the client side proxy
	//err = (&SecureTCPConn{
	//	EncodeCipher:    localConn.EncodeCipher,
	//	DecodeCipher:    localConn.DecodeCipher,
	//	ReadWriteCloser: dstServer,
	//}).EncodeCopy(localConn)
	err = (&SecureTCPConn{
		EncodeCipher:    localConn.EncodeCipher,
		DecodeCipher:    localConn.DecodeCipher,
		ReadWriteCloser: dstServer,
	}).DirectEncodeCopy(localConn)
	if err != nil {
		localConn.Close()
		dstServer.Close()
	}
}
