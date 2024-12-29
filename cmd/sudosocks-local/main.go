package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sudoku_go"
	"sudoku_go/cmd"
	"sudoku_go/global"
	"sudoku_go/sudoku"
)

const (
	DefaultListenAddr = "127.0.0.1:7789"
	//DefaultRemoteAddr = "172.245.242.86:17789"
	DefaultRemoteAddr = "127.0.0.1:17789"
)

func init() {
	sudoku.GenByteMap()
	sudoku.AllPuzzle()
	global.ByteList = &[256][]string{}
	for i := 0; i < 256; i++ {
		global.ByteList[i] = sudoku.ByteToSudokuList(byte(i))
	}
}

func main() {
	log.SetFlags(log.Lshortfile)

	listenAddr := flag.String("l", DefaultListenAddr, "Local listen address")
	remoteAddr := flag.String("r", DefaultRemoteAddr, "Remote server address")

	flag.Parse()

	// 默认配置
	config := &cmd.Config{
		ListenAddr: *listenAddr,
		RemoteAddr: *remoteAddr,
	}

	// 启动 local 端并监听
	lsLocal, err := sudoku_go.NewLsLocal(config.ListenAddr, config.RemoteAddr)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println()
	log.Fatalln(lsLocal.Listen(func(listenAddr net.Addr) {
		fmt.Println(fmt.Sprintf(`
sudosocks-local 启动成功，配置如下：
本地监听地址：%s
远程服务地址：%s
`, listenAddr, config.RemoteAddr))
	}))
}
