package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"time"
)

const (
	rootCertFileName = "../../cert/ca.cer"
	clientCertFileName = "../../cert/client/client.cer"
	clientKeyFileName  = "../../cert/client/client.key"
)



func main() {
	flag.Parse()
	buf, err := ioutil.ReadFile(rootCertFileName)
	if err != nil {
		return
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(buf)

	cert, err := tls.LoadX509KeyPair(clientCertFileName, clientKeyFileName)
	if err != nil {
		log.Fatalln(err)
	}

	config := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:          pool,
		InsecureSkipVerify: false,
		ServerName: 	"chry-server",
	}

	//注意这里要使用证书中包含的主机名称
	conn, err := tls.Dial("tcp", "127.0.0.1:8888", config)
	if err != nil {
		log.Fatalln(err.Error())
	}

	defer conn.Close()
	log.Println("Client Connect To ", conn.RemoteAddr())
	status := conn.ConnectionState()
	fmt.Printf("%#v\n", status)
	buf = make([]byte, 1024)
	ticker := time.NewTicker(1 * time.Millisecond * 500)
	for {
		select {
		case <-ticker.C:
			{
				_, err = io.WriteString(conn, "hello")
				if err != nil {
					log.Fatalln(err.Error())
				}
				len, err := conn.Read(buf)
				if err != nil {
					fmt.Println(err.Error())
				} else {
					fmt.Println("Receive From Server:", string(buf[:len]))
				}
			}
		}
	}

}
