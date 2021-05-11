package main

import (
	"fmt"
	"log"
	"net/http"
)

const (
	certFileName = "cacert.cer"
	keyFileName  = "capriv.key"
)

func sayhelloName(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!") //这个写入到w的是输出到客户端的
}

func main() {
	// log打印设置: Lshortfile文件名+行号  LstdFlags日期加时间
	log.SetFlags(log.Llongfile | log.LstdFlags | log.Lmicroseconds)

	http.HandleFunc("/", sayhelloName) //设置访问的路由

	// 启动https服务  https://192.168.0.160:9090/
	err := http.ListenAndServeTLS(":9090", certFileName, keyFileName, nil) //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
