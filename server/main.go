package main

import (
	"encoding/json"
	"flag"
	"github.com/armon/go-socks5"
	"github.com/mccutchen/go-httpbin/v2/httpbin"
	"golang.org/x/net/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var cfgPath, listen, path, token, tls, certFile, keyFile string
var socks, _ = socks5.New(&socks5.Config{})

func isValidToken(ws *websocket.Conn) (bool, string) {
	v := ws.Request().Header.Get("token")
	return v == token, v
}

func ws2socks(ws *websocket.Conn) {
	log.Printf("[INFO] receive ws: %p\n", ws)
	defer func() {
		log.Printf("[INFO] close ws: %p\n", ws)
		_ = ws.Close()
	}()
	if ok, token := isValidToken(ws); !ok {
		log.Println("[WARNING] incorrect token:", token)
		return
	}
	err := socks.ServeConn(ws)
	if err != nil {
		//log.Println("[ERROR] socks5 serve error:", err)
		return
	}
}

func main() {
	parseArgs()
	initConfig()
	http.Handle(path, websocket.Handler(ws2socks))
	if path != "/" {
		http.Handle("/", httpbin.New().Handler())
	}
	if strings.ToLower(tls) == "true" {
		log.Printf("listen: wss://%s%s, token: %s.", listen, path, token)
		log.Fatalln(http.ListenAndServeTLS(listen, certFile, keyFile, nil))
	} else {
		log.Printf("listen: ws://%s%s, token: %s.\n", listen, path, token)
		log.Fatalln(http.ListenAndServe(listen, nil))
	}
}

func parseArgs() {
	flag.StringVar(&cfgPath, "config", "server.json", "Config File Path")
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
	listen, path, token = config["listen"], config["path"], config["token"]
	tls, certFile, keyFile = config["tls"], config["certFile"], config["keyFile"]
}
