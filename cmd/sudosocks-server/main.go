package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"sudoku_go"
	"sudoku_go/cmd"
	"sudoku_go/global"
	"sudoku_go/sudoku"
)

var version = "master"

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

	// 优先从环境变量中获取监听端口
	//port, err := strconv.Atoi(os.Getenv("LIGHTSOCKS_SERVER_PORT"))
	//if err != nil {
	//	// 其次采用默认端口
	//	port = 17789
	//}
	// 再次服务端监听端口随机生成
	//if err != nil {
	//	port, err = freeport.GetFreePort()
	//}
	argPort := flag.String("p", "17789", "Port")
	port, _ := strconv.Atoi(*argPort)

	fmt.Println("Port: ", port)

	// 默认配置
	config := &cmd.Config{
		ListenAddr: fmt.Sprintf(":%d", port),
	}
	config.ReadConfig()
	config.SaveConfig()

	// 启动 server 端并监听
	lsServer, err := sudoku_go.NewLsServer(config.ListenAddr)
	if err != nil {
		log.Fatalln(err)
	}
	lsServer.Listen(func(listenAddr net.Addr) {
		log.Println(fmt.Sprintf(`
sudosocks-server:%s 启动成功，配置如下：
服务监听地址：
%s`, version, listenAddr))
	})
}
