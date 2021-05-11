package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	caCertFileName = "cacert.cer"
)

func main() {
	// log打印设置: Lshortfile文件名+行号  LstdFlags日期加时间
	log.SetFlags(log.Llongfile | log.LstdFlags | log.Lmicroseconds)

	caCer, err := ioutil.ReadFile(caCertFileName)
	if err != nil {
		log.Fatalln("failed to read", caCertFileName)
	}

	pool := x509.NewCertPool()
	ok := pool.AppendCertsFromPEM(caCer)
	if !ok {
		log.Fatalln("failed to parse root certificate")
	}

	// 设置http请求超时时间,
	client := &http.Client{
		//Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},	// 跳过https校验
		Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}},
		Timeout:   time.Second * 5,
	}

	/*
		注意：请求url中的host必须是创建证书时指定的host之一，详见cert/main.go里面的host字段
		报错信息如下：
			x509: certificate is not valid for any names, but wanted to match localhost
	*/
	resp, err := client.Post("https://192.168.1.99:9090", "application/text", nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	log.Println(resp)
}
