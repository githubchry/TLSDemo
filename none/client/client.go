package main

import (
	"bufio"
	"log"
	"net"
	"os"
)

func main() {

	// log打印设置: Lshortfile文件名+行号  LstdFlags日期加时间
	log.SetFlags(log.Llongfile | log.LstdFlags | log.Lmicroseconds)

	conn, err := net.Dial("tcp", "127.0.0.1:8864")
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer conn.Close()

	log.Println("Client Connect To ", conn.RemoteAddr())

	//预先准备消息缓冲区
	buffer := make([]byte, 1024)

	//准备命令行标准输入
	reader := bufio.NewReader(os.Stdin)

	for {
		lineBytes, _, _ := reader.ReadLine()

		if string(lineBytes) == "exit" {
			break
		}

		conn.Write(lineBytes)
		log.Printf("发送msg: %s", lineBytes)

		n, err := conn.Read(buffer)
		if err != nil {
			log.Fatalln(err.Error())
		}

		log.Printf("收到msg: %s", string(buffer[0:n]))

	}
}
