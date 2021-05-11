package main

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
)

const (
	organization = "CHRY"
	certFileName = "cacert.cer"
	keyFileName  = "capriv.key"
)

/*
golang标准库提供了flag包来处理命令行参数
1. 定义flag字段：字段名/默认值/帮助描述
2. 调用flag.Parse()解析所有命令行参数到预定义好的flag字段里面
3. 如果解析失败则会打印所有flag的usage
示例：
go run server.go  使用默认参数
go run server.go -host=127.0.0.1 -ca=false 指定参数
go run server.go -ca=123 无法解析123为bool，打印帮助描述
*/
var (
	host      = flag.String("host", "192.168.0.160,192.168.1.99,www.chenruiyun.com", "用逗号分隔的主机名和IP来生成证书,不能有空格")
	validFrom = flag.String("start-date", "", "默认创建日期格式为Jan 1 15:04:05 2011")
	validFor  = flag.Duration("duration", 3650*24*time.Hour, "该证书的有效期")
	isCA      = flag.Bool("ca", true, "该证书是否应该是它自己的证书权威机构")
	// 以下算法三选一 程序判断优先级： ecdsaCurve > ed25519Key > rsaBits
	rsaBits    = flag.Int("rsa-bits", 2048, "要生成的RSA密钥的大小. 如果设置了--ecdsa-curve，则忽略")
	ecdsaCurve = flag.String("ecdsa-curve", "", "用ECDSA曲线生成密钥. 有效值: P224, P256 (推荐), P384, P521.")
	ed25519Key = flag.Bool("ed25519", false, "生成Ed25519密钥")
)

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	case ed25519.PrivateKey:
		return k.Public().(ed25519.PublicKey)
	default:
		return nil
	}
}

func main() {
	// log打印设置: Lshortfile文件名+行号  LstdFlags日期加时间
	log.SetFlags(log.Llongfile | log.LstdFlags | log.Lmicroseconds)

	//解析命令行参数到定义的flag
	flag.Parse()

	if len(*host) == 0 {
		log.Fatalf("Missing required --host parameter")
	}

	// 生成证书私钥
	var priv interface{}
	var err error
	switch *ecdsaCurve {
	case "":
		if *ed25519Key {
			_, priv, err = ed25519.GenerateKey(rand.Reader)
		} else {
			priv, err = rsa.GenerateKey(rand.Reader, *rsaBits)
		}
	case "P224":
		priv, err = ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	case "P256":
		priv, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "P384":
		priv, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case "P521":
		priv, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		log.Fatalf("Unrecognized elliptic curve: %q", *ecdsaCurve)
	}
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	// 至此私钥已经生成好了，下面保存到文件
	keyOut, err := os.OpenFile(keyFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Failed to open key.pem for writing: %v", err)
		return
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		log.Fatalf("Unable to marshal private key: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		log.Fatalf("Failed to write data to key.pem: %v", err)
	}
	if err := keyOut.Close(); err != nil {
		log.Fatalf("Error closing %s: %v", keyFileName, err)
	}
	log.Printf("生成%s证书私钥并保存到文件：", keyFileName)

	// 准备生成公钥

	// ECDSA、ED25519和RSA主体密钥应在x509.Certificate模板中设置DigitalSignature KeyUsage位
	keyUsage := x509.KeyUsageDigitalSignature
	//只有RSA主题密钥需要设置KeyEncipherment KeyUsage位。在TLS上下文中，这个KeyUsage是RSA密钥交换和身份验证特有的。
	if _, isRSA := priv.(*rsa.PrivateKey); isRSA {
		keyUsage |= x509.KeyUsageKeyEncipherment
	}

	var notBefore time.Time
	if len(*validFrom) == 0 {
		notBefore = time.Now()
	} else {
		notBefore, err = time.Parse("Jan 2 15:04:05 2006", *validFrom)
		if err != nil {
			log.Fatalf("Failed to parse creation date: %v", err)
		}
	}

	notAfter := notBefore.Add(*validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("Failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{organization},
		},
		NotBefore: notBefore, //证书有效期开始时间
		NotAfter:  notAfter,  //证书有效期结束时间

		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(*host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if *isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	// 生成证书
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %v", err)
	}

	certOut, err := os.Create(certFileName)
	if err != nil {
		log.Fatalf("Failed to open cert.pem for writing: %v", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		log.Fatalf("Failed to write data to cert.pem: %v", err)
	}
	if err := certOut.Close(); err != nil {
		log.Fatalf("Error closing cert.pem: %v", err)
	}
	log.Println("wrote", certFileName)

}
