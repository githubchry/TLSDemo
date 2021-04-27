## 基于 CA 的 TLS 证书认证

```shell
注意Common Name的填写要求
1. CA证书/服务端证书/客户端证书的Common Name不能重复!
2. 记住服务端证书的Common Name, 在客户端连接时可能需要指定!


生成根证书密钥
openssl genrsa -out ca.key 2048
生成根证书公钥
openssl req -new -x509 -days 3650 -key ca.key -out ca.pem
 
生成服务端证书密钥
openssl genrsa -out server.key 2048
根据服务端证书公钥密成证书请求文件
openssl req -new -key server.key -out server.csr
基于CA证书签发服务端证书公钥
openssl x509 -req -sha256 -days 3650 -CA ca.pem -CAkey ca.key -CAcreateserial -in server.csr -out server.pem


生成客户端证书密钥
openssl ecparam -genkey -name secp384r1 -out client.key
根据客户端证书密钥生成证书请求文件
openssl req -new -key client.key -out client.csr
基于CA证书签发客户端证书公钥
openssl x509 -req -sha256 -days 3650 -CA ca.pem -CAkey ca.key -CAcreateserial -in client.csr -out client.pem



以上在go1.15版本行不通了!!!
请使用go自带的接口创建相关证书!!!
疑难问题解决方案:
https://github.com/golang/go/issues/39568#issuecomment-671424481

https://zhuanlan.zhihu.com/p/105232920
数字证书和golang的研究
https://blog.csdn.net/u010846177/article/details/54357239
使用golang进行证书签发和双向认证
https://blog.csdn.net/weixin_34419326/article/details/89058910
grpc使用自制CA证书校验公网上的连接请求
https://www.jianshu.com/p/751066a6c689

Golang gRPC笔记03 基于 CA 的 TLS 证书认证
https://www.cnblogs.com/qq037/p/13284461.html

```