package sudoku_go

import (
	"io"
	"log"
	"net"
	"sync"
	"syscall"
	"time"
)

const (
	bufSize = 1024
)

// 加密传输的 TCP Socket
type SecureTCPConn struct {
	io.ReadWriteCloser
	EncodeCipher *cipher
	DecodeCipher *cipher
}

var (
	Tx     uint64
	Rx     uint64
	TxLock sync.RWMutex
	RxLock sync.RWMutex
)

// 从输入流里读取加密过的数据，解密后把原数据放到bs里
func (secureSocket *SecureTCPConn) DecodeRead(bs []byte) (n int, err error) {
	// 开辟六倍bs的buf
	buf := make([]byte, len(bs)*6)
	n, err = secureSocket.Read(buf)
	if err != nil {
		return
	}
	copy(bs, secureSocket.DecodeCipher.Decode(buf[:n]))
	n /= 6
	return
}

// 把放在bs里的数据加密后立即全部写入输出流
func (secureSocket *SecureTCPConn) EncodeWrite(bs []byte) (int, error) {
	sixTimeBs := secureSocket.EncodeCipher.Encode(bs)
	n, err := secureSocket.Write(sixTimeBs)
	return n / 6, err
}

// 从src中源源不断的读取原数据加密后写入到dst，直到src中没有数据可以再读取
func (secureSocket *SecureTCPConn) EncodeCopy(dst io.ReadWriteCloser) error {
	buf := make([]byte, bufSize)
	for {
		readCount, errRead := secureSocket.Read(buf)
		if errRead != nil {
			if errRead != io.EOF {
				return errRead
			} else {
				return nil
			}
		}
		if readCount > 0 {
			writeCount, errWrite := (&SecureTCPConn{
				ReadWriteCloser: dst,
				EncodeCipher:    secureSocket.EncodeCipher,
				DecodeCipher:    secureSocket.DecodeCipher,
			}).EncodeWrite(buf[0:readCount])
			if errWrite != nil {
				return errWrite
			}
			if readCount != writeCount {
				return io.ErrShortWrite
			}

			TxLock.Lock()
			Tx += uint64(readCount)
			TxLock.Unlock()
		}
	}
}

// 从src中源源不断的读取加密后的数据解密后写入到dst，直到src中没有数据可以再读取
func (secureSocket *SecureTCPConn) DecodeCopy(dst io.Writer) error {
	buf := make([]byte, bufSize)
	for {
		readCount, errRead := secureSocket.DecodeRead(buf)
		if errRead != nil {
			if errRead != io.EOF {
				return errRead
			} else {
				return nil
			}
		}
		if readCount > 0 {
			writeCount, errWrite := dst.Write(buf[0:readCount])
			if errWrite != nil {
				return errWrite
			}
			if readCount != writeCount {
				return io.ErrShortWrite
			}

			RxLock.Lock()
			Rx += uint64(readCount)
			RxLock.Unlock()
		}
	}
}

func (srcSocket *SecureTCPConn) DirectEncodeCopy(dst io.Writer) error {
	for {
		buf := make([]byte, bufSize)
		readCount, errRead := srcSocket.Read(buf)
		if errRead != nil {
			if errRead != io.EOF {
				return errRead
			} else {
				return nil
			}
		}
		if readCount > 0 {
			writeCount, errWrite := dst.Write(buf[0:readCount])
			//log.Println(buf[0:readCount])
			if errWrite != nil {
				return errWrite
			}
			if readCount != writeCount {
				return io.ErrShortWrite
			}
		}

	}
}

func (srcSocket *SecureTCPConn) DirectDEcodeCopy(dst io.Writer) error {
	for {
		buf := make([]byte, bufSize)
		readCount, errRead := srcSocket.Read(buf)
		if errRead != nil {
			if errRead != io.EOF {
				return errRead
			} else {
				return nil
			}
		}
		if readCount > 0 {
			writeCount, errWrite := dst.Write(buf[0:readCount])
			//log.Println(buf[0:readCount])
			if errWrite != nil {
				return errWrite
			}
			if readCount != writeCount {
				return io.ErrShortWrite
			}
		}

	}
}

// see net.DialTCP
func DialTCPSecure(raddr *net.TCPAddr, SBcode uint8) (*SecureTCPConn, error) {
	var dialer = net.Dialer{Timeout: 5 * time.Second, KeepAlive: 5 * time.Second, Control: func(network, address string, c syscall.RawConn) error {
		c.Control(func(fd uintptr) {
			//Outbound connection needs to be protected in Android VPN mode
			protect(int(fd))
		})
		return nil
	}}
	remoteConn, err := dialer.Dial("tcp", raddr.String())
	//remoteConn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		return nil, err
	}
	return &SecureTCPConn{
		ReadWriteCloser: remoteConn,
		EncodeCipher: &cipher{
			SBcode: SBcode,
		},
		DecodeCipher: &cipher{
			SBcode: SBcode,
		},
	}, nil
}

// see net.ListenTCP
func ListenSecureTCP(laddr *net.TCPAddr, SBCode uint8, handleConn func(localConn *SecureTCPConn), didListen func(listenAddr net.Addr)) error {
	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return err
	}

	defer listener.Close()

	if didListen != nil {
		didListen(listener.Addr())
	}

	for {
		localConn, err := listener.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println("Accept client connection : ", localConn.RemoteAddr())
		// localConn被关闭时直接清除所有数据 不管没有发送的数据
		localConn.SetLinger(0)
		go handleConn(&SecureTCPConn{
			ReadWriteCloser: localConn,
			EncodeCipher: &cipher{
				SBcode: SBCode,
			},
			DecodeCipher: &cipher{
				SBcode: SBCode,
			},
		})
	}
}

func protect(fd int) {
	addr, err := net.ResolveUnixAddr("unix", "protect_path")
	if err != nil {
		log.Printf("Failed to resolve: %v\n", err)
		return
	}

	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		//log.Printf("Failed to dial: %v\n", err)
		return
	}
	defer conn.Close()

	log.Println("Connected to VPN Service")
	err = sendFD(conn, fd)
	if err != nil {
		log.Printf("Failed to protect socket: %v\n", err)
		return
	}
}

// Put sends file descriptors to Unix domain socket.
//
// Please note that the number of descriptors in one message is limited
// and is rather small.
// Use conn.File() to get a file if you want to put a network connection.
func sendFD(via *net.UnixConn, fd int) error {
	viaf, err := via.File()
	if err != nil {
		return err
	}
	socket := int(viaf.Fd())

	rights := syscall.UnixRights(fd)
	err = syscall.Sendmsg(socket, nil, rights, nil, 0)
	log.Println("Send out protected sockets")
	if err != nil {
		return err
	}
	data := make([]byte, 1024)
	_, _, _, _, err = syscall.Recvmsg(socket, nil, data, 0)
	log.Println("Recv response from VPN service")
	if err != nil {
		return err
	}
	return nil
}
