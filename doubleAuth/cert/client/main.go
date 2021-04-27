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
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
)

const (
	organization = "CHRY"
	rootCertFileName   = "../ca.cer"
	rootKeyFileName    = "../ca.key"
	clientCertFileName = "client.cer"
	clientKeyFileName  = "client.key"
	clientCsrFileName  = "client.csr"
)

var (
	host       = flag.String("host", 			"chry-client", 	"用逗号分隔的主机名和IP来生成证书")
	validFrom  = flag.String("start-date", 	"", 				"创建日期格式为Jan 1 15:04:05 2011")
	validFor   = flag.Duration("duration", 	3650*24*time.Hour, 		"该证书的有效期")
	isCA       = flag.Bool("ca", 				false, 			"该证书是否应该是它自己的证书权威机构")
	rsaBits    = flag.Int("rsa-bits", 		2048, 			"要生成的RSA密钥的大小. 如果设置了--ecdsa-curve，则忽略")
	ecdsaCurve = flag.String("ecdsa-curve", 	"", 				"用ECDSA曲线生成密钥. 有效值: P224, P256 (推荐), P384, P521")
	ed25519Key = flag.Bool("ed25519", 		false, 			"生成Ed25519密钥")
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
	log.SetFlags(log.Llongfile | log.LstdFlags)

	flag.Parse()

	if len(*host) == 0 {
		log.Fatalf("Missing required --host parameter")
	}

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

	// ECDSA, ED25519 and RSA subject keys should have the DigitalSignature
	// KeyUsage bits set in the x509.Certificate template
	keyUsage := x509.KeyUsageDigitalSignature
	// Only RSA subject keys should have the KeyEncipherment KeyUsage bits set. In
	// the context of TLS this KeyUsage is particular to RSA key exchange and
	// authentication.
	if _, isRSA := priv.(*rsa.PrivateKey); isRSA {
		keyUsage |= x509.KeyUsageKeyEncipherment
	}

	// 保存私钥到文件
	keyOut, err := os.OpenFile(clientKeyFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Failed to open %s for writing: %v", clientKeyFileName, err)
		return
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		log.Fatalf("Unable to marshal private key: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		log.Fatalf("Failed to write data to %s: %v", clientKeyFileName, err)
	}
	if err := keyOut.Close(); err != nil {
		log.Fatalf("Error closing %s: %v", clientKeyFileName, err)
	}
	log.Println("已经生成私钥:", clientKeyFileName)

	//=========================================================================
	csrtmpl := x509.CertificateRequest{
		Subject: pkix.Name{
			Organization: []string{organization},
		},
	}

	hosts := strings.Split(*host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			csrtmpl.IPAddresses = append(csrtmpl.IPAddresses, ip)
		} else {
			csrtmpl.DNSNames = append(csrtmpl.DNSNames, h)
		}
	}

	// 生成证书请求
	csrOut, err := os.OpenFile(clientCsrFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Failed to open %s for writing: %v", clientCsrFileName, err)
		return
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &csrtmpl, priv)
	if err != nil {
		log.Fatalf("Failed to create CSR: %v", err)
	}

	if err = pem.Encode(csrOut, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes}); err != nil {
		log.Fatalf("Failed to Encode CSR: %v", err)
	}

	if err := csrOut.Close(); err != nil {
		log.Fatalf("Error closing %s: %v", clientCsrFileName, err)
	}
	log.Println("已经生成证书请求:", clientCsrFileName)

	// 基于CA证书签发服务端证书公钥
	//解析根证书
	rootCertFile, err := ioutil.ReadFile(rootCertFileName)
	if err != nil {
		log.Fatalf("Failed to read %s:%v", rootCertFileName, err)
	}
	rootCertBlock, _ := pem.Decode(rootCertFile)

	rootCert, err := x509.ParseCertificate(rootCertBlock.Bytes)
	if err != nil {
		log.Fatalf("Failed to ParseCertificate %s:%v", rootCertFileName, err)
	}

	//解析私钥
	rootKeyFile, err := ioutil.ReadFile(rootKeyFileName)
	if err != nil {
		log.Fatalf("Failed to read %s:%v", rootKeyFileName, err)
	}
	rootKeyBlock, _ := pem.Decode(rootKeyFile)
	rootKey, err := x509.ParsePKCS8PrivateKey(rootKeyBlock.Bytes)
	if err != nil {
		log.Fatalf("Failed to ParsePKCS1PrivateKey %s:%v", rootKeyFileName, err)
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
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	hosts = strings.Split(*host, ",")
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

	// 基于CA证书 签发服务端证书公钥
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, rootCert, publicKey(priv), rootKey)
	if err != nil {
		log.Fatalf("Failed to create certificate: %v", err)
	}

	certOut, err := os.Create(clientCertFileName)
	if err != nil {
		log.Fatalf("Failed to open cert.pem for writing: %v", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		log.Fatalf("Failed to write data to cert.pem: %v", err)
	}
	if err := certOut.Close(); err != nil {
		log.Fatalf("Error closing cert.pem: %v", err)
	}
	log.Println("已经签发公钥", clientCertFileName)
}