package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"os"
)

const (
	rootCertFileName   = "ca.cer"
	clientCertFileName = "client.cer"
	clientKeyFileName  = "client.key"
)

func main() {

	// log打印设置: Lshortfile文件名+行号  LstdFlags日期加时间
	log.SetFlags(log.Llongfile | log.LstdFlags | log.Lmicroseconds)

	rootCertBytes, err := ioutil.ReadFile(rootCertFileName)
	if err != nil {
		return
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(rootCertBytes)

	cert, err := tls.LoadX509KeyPair(clientCertFileName, clientKeyFileName)
	if err != nil {
		log.Fatalln(err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
		ServerName:   "chry-server", //注意这里要使用证书中包含的主机名称
	}

	conn, err := tls.Dial("tcp", "192.168.1.99:8864", tlsConfig)
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
