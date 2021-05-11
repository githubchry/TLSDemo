## 基于 CA 的 TLS 证书认证

```
1. 自建根证书私钥ca.key
2. 根据私钥ca.key创建公钥ca.crt

3. 创建自建服务证书私钥server.key
4. 根据server.key创建证书请求server.csr
5. 利用根证书私钥ca.key和公钥ca.crt签署SSL自建证书请求server.csr得到自建证书公钥server.crt
```





[使用OpenSsl自己CA根证书,二级根证书和颁发证书(亲测步骤)](https://www.cnblogs.com/lzpong/p/10450293.html)

```shell
注意Common Name的填写要求
1. CA证书/服务端证书/客户端证书的Common Name不能重复!
2. 记住服务端证书的Common Name, 在客户端连接时可能需要指定!


生成根证书私钥
openssl genrsa -out ca.key 2048
生成根证书公钥
openssl req -new -x509 -days 3650 -key ca.key -out ca.pem
 
生成服务端证书私钥
openssl genrsa -out server.key 2048
根据服务端证书私钥生成证书请求文件
openssl req -new -key server.key -out server.csr
基于CA证书签发服务端证书公钥
openssl x509 -req -sha256 -days 3650 -CA ca.pem -CAkey ca.key -CAcreateserial -in server.csr -out server.pem


生成客户端证书私钥
openssl ecparam -genkey -name secp384r1 -out client.key
根据客户端证书私钥生成证书请求文件
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





## 相关概念

[一文读懂Https的安全性原理、数字证书、单项认证、双项认证](https://blog.csdn.net/hellojackjiang2011/article/details/103622323)

1. 单向认证只要求站点部署了ssl证书就行，任何用户都可以去访问（IP被限制除外等），只是服务端提供了身份认证。而双向认证则是需要是服务端需要客户端提供身份认证，只能是服务端允许的客户能去访问，安全性相对于要高一些

2. 双向认证SSL 协议的具体通讯过程，这种情况要求服务器和客户端双方都有证书。

3. 单向认证SSL 协议不需要客户端拥有CA证书，以及在协商对称密码方案，对称通话密钥时，服务器发送给客户端的是没有加过密的（这并不影响SSL过程的安全性）密码方案。

4. 如果有第三方攻击，获得的只是加密的数据，第三方要获得有用的信息，就需要对加密的数据进行解密，这时候的安全就依赖于密码方案的安全。而幸运的是，目前所用的密码方案，只要通讯密钥长度足够的长，就足够的安全。这也是我们强调要求使用128位加密通讯的原因。

5. 一般Web应用都是采用单向认证的，原因很简单，用户数目广泛，且无需做在通讯层做用户身份验证，一般都在应用逻辑层来保证用户的合法登入。但如果是企业应用对接，情况就不一样，可能会要求对客户端（相对而言）做身份验证。这时就需要做双向认证。





Windows中的证书扩展名有好几种，比如.cer和.crt。通常而言，.cer文件是二进制数据，而.crt文件包含的是ASCII数据。

cer文件包含依据DER（Distinguished Encoding Rules）规则编码的证书数据，这是x.690标准中指定的编码格式。

X.509是一个最基本的公钥格式标准，里面规定了证书需要包含的各种信息。通常我们提到的证书，都是这个格式的，里面包含了公钥、发布者的数字签名、有效期等内容。要强调的是，它只里面是不包含私钥的。相关的格式有：DER、PEM、CER、CRT。



## 单向认证

服务器：公钥+私钥

客户端：公钥



### 自建证书方式

如果应用为B/S架构，必须创建**根**证书公钥并安装到系统，否则浏览器报错：你的连接不是专用连接。

如果应用为C/S架构，没有该限制。

建议直接创建根证书。

- 创建根证书私钥

- 生成根证书公钥



**安装根证书公钥到系统：**

双击公钥 - 常规 - 安装证书 - 下一步 - 将所有的证书都放入下列存储 - 浏览 - 受信任的证书颁发机构 - 下一步 - 完成 - 确认安装 - 重启浏览器



### 第三方机构颁发方式

通过购买和白嫖方式可直接得到公钥和私钥，不需要自己创建生成，也不需要在浏览器安装证书。

也可自行生成私钥，然后生成证书请求文件，把证书请求文件提交给第三方签名，得到公钥证书。

- 创建非根证书私钥

- 生成证书请求
- 将证书请求提交给第三方机构签署，得到公钥



## 双向认证

服务器：服务器公钥+服务器私钥+根证书公钥

客户端：客户端公钥+客户端私钥+根证书公钥



### 自建证书方式

如果应用为B/S架构，必须创建**根**证书公钥并安装到系统，否则浏览器报错：你的连接不是专用连接。

如果应用为C/S架构，没有该限制。

建议直接创建根证书。

- 创建根证书私钥

- 生成根证书公钥
- 
- 创建服务端证书私钥
- 生成服务端证书请求
- 基于根证书签发服务端证书公钥
- 
- 创建客户端证书私钥
- 生成客户端证书请求
- 基于根证书签发客户端证书公钥



