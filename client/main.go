package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"golang.org/x/net/websocket"
	"io"
	"io/ioutil"
	"log"
	"net"
	"sync"
)

var cfgPath, listen, url, token, certFile string
var wsConfig *websocket.Config

func socks2ws(socks *net.TCPConn) {
	ws, err := websocket.DialConfig(wsConfig)
	if err != nil {
		_ = socks.Close()
		log.Println("dial ws error:", err)
		return
	}

	var wg sync.WaitGroup
	ioCopy := func(dst io.Writer, src io.Reader) {
		defer func() { _, _ = socks.Close(), ws.Close(); wg.Done() }()
		_, _ = io.Copy(dst, src)
	}
	wg.Add(2)
	go ioCopy(ws, socks)
	go ioCopy(socks, ws)
	wg.Wait()
}

func main() {
	parseArgs()
	initConfig()
	listener, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("listen tcp at:", listen)
	}

	for {
		conn, _ := listener.Accept()
		if conn != nil {
			go socks2ws(conn.(*net.TCPConn))
		}
	}
}
func parseArgs() {
	flag.StringVar(&cfgPath, "config", "client.json", "Config File Path")
	flag.Parse()
}

func initConfig() {
	content, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		log.Fatalln("read config file error:", err)
	}
	var config map[string]string
	if err = json.Unmarshal(content, &config); err != nil {
		log.Fatalln("unmarshal json error:", err)
	}
	listen, url, token, certFile = config["listen"], config["url"], config["token"], config["certFile"]
	// generate websocket config
	wsConfig, _ = websocket.NewConfig(url, url)
	// load certificate file
	content, err = ioutil.ReadFile(certFile)
	if err != nil {
		log.Fatalln("read cert file error:", err)
	}
	roots := x509.NewCertPool()
	if ok := roots.AppendCertsFromPEM(content); ok != true {
		log.Fatalln("append cert fail")
	}
	wsConfig.TlsConfig = &tls.Config{RootCAs: roots}
	wsConfig.Header.Set("token", token)
}
