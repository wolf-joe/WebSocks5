package main

import (
	"flag"
	"github.com/armon/go-socks5"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
)

var listen, token string
var socks, _ = socks5.New(&socks5.Config{})

func isValidToken(ws *websocket.Conn) (bool, string) {
	v := ws.Request().Header.Get("token")
	return v == token, v
}

func ws2socks(ws *websocket.Conn) {
	defer ws.Close()
	if ok, token := isValidToken(ws); !ok {
		log.Println("incorrect token:", token)
		return
	}
	err := socks.ServeConn(ws)
	if err != nil {
		log.Println("socks5 serve error:", err)
		return
	}
}

func main() {
	parseArgs()
	log.Printf("listen: %s, token: %s.", listen, token)
	http.Handle("/", websocket.Handler(ws2socks))
	log.Fatal(http.ListenAndServe(listen, nil))
}

func parseArgs() {
	flag.StringVar(&listen, "listen", "127.0.0.1:5000", "WebSocks5 server listen address.")
	flag.StringVar(&token, "token", "HelloWorld", "WebSocks5 server token (password).")
	flag.Parse()
}
