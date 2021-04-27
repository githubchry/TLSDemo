package main

import (
	"fmt"
	"log"
	"net/http"
)


const (
	certFileName = "../../cert/cacert.cer"
	keyFileName  = "../../cert/capriv.key"
)

func sayhelloName(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!") //这个写入到w的是输出到客户端的
}

func main() {
	http.HandleFunc("/", sayhelloName) //设置访问的路由
	// 启动http服务
	err := http.ListenAndServeTLS(":9090", certFileName, keyFileName,nil) //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}