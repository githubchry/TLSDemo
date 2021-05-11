package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net"
)

const (
	rootCertFileName   = "ca.cer"
	serverCertFileName = "server.cer"
	serverKeyFileName  = "server.key"
)

func echo(conn net.Conn) {
	defer conn.Close()
	for {
		reader := bufio.NewReader(conn)
		var buf [128]byte
		n, err := reader.Read(buf[:])
		if err != nil {
			log.Println("read from client failed:", err)
			break
		}
		recvStr := string(buf[:n])
		log.Println("recv", n, "byte from client:", recvStr)
		conn.Write([]byte(recvStr))
	}
}

func main() {

	// log打印设置: Lshortfile文件名+行号  LstdFlags日期加时间
	log.SetFlags(log.Llongfile | log.LstdFlags | log.Lmicroseconds)

	rootCertBytes, err := ioutil.ReadFile(rootCertFileName)
	if err != nil {
		log.Fatal(err)
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(rootCertBytes)

	cert, err := tls.LoadX509KeyPair(serverCertFileName, serverKeyFileName)
	if err != nil {
		log.Fatalln(err.Error())
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pool,
	}

	listener, err := tls.Listen("tcp", ":8864", tlsConfig)
	if err != nil {
		log.Fatalln(err.Error())
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err.Error())
			continue
		} else {
			go echo(conn)
		}
	}
}
