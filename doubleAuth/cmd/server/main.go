package main

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"time"
)


const (
	rootCertFileName = "../../cert/ca.cer"
	serverCertFileName = "../../cert/server/server.cer"
	serverKeyFileName  = "../../cert/server/server.key"
)

func process(conn net.Conn) {
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
	flag.Parse()
	buf, err := ioutil.ReadFile(rootCertFileName)
	if err != nil {
		log.Fatal(err)
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(buf)

	crt, err := tls.LoadX509KeyPair(serverCertFileName, serverKeyFileName)
	if err != nil {
		log.Fatalln(err.Error())
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{crt},
		ClientAuth:         tls.RequireAndVerifyClientCert,
		ClientCAs:          pool,
	}

	tlsConfig.Time = time.Now

	tlsConfig.Rand = rand.Reader
	l, err := tls.Listen("tcp", ":8888", tlsConfig)
	if err != nil {
		log.Fatalln(err.Error())
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err.Error())
			continue
		} else {
			go process(conn)
		}
	}
}
