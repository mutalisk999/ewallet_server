package main

// Run first `go run main.go server`
// and `go run main.go client` as many times as you want.
// Originally written by: github.com/antlaw to describe an old issue.
import (
	"controller"
	"fmt"
	"github.com/kataras/iris"
	"github.com/kataras/iris/websocket"
	"session"
	"utils"
	"config"
)

// addresses in white list should be allowed
func isRemoteAddrAllowed(remoteAddr string) bool {
	for _, allowip := range config.GlobalConfig.AllowIpConfig.AllowIps {
		if remoteAddr == allowip {
			return true
		}
	}
	return false
}

// OnConnect handles incoming websocket connection
func OnConnect(c websocket.Connection) {
	fmt.Println("websocket.OnConnect()")

	remoteAddr := c.Context().RemoteAddr()
	if !isRemoteAddrAllowed(remoteAddr) {
		c.Disconnect()
		return
	}

	c.On("login", func(message string) { controller.OnLogin(message, c) })
	c.On("pubkey", func(message string) { controller.OnPubKey(message, c) })
	c.On("transaction", func(message string) { controller.OnTransaction(message, c) })
	c.On("wallet", func(message string) { controller.OnWallet(message, c) })
	// ok works too c.EmitMessage([]byte("dsadsa"))
	c.OnDisconnect(func() { OnDisconnect(c) })

}

// ServerLoop listen and serve websocket requests
func ServerLoop() {
	session.InitSessionMgr()
	utils.InitGlobalError()
	app := iris.New()

	ws := websocket.New(websocket.Config{})

	// register the server on an endpoint.
	// see the inline javascript code i the websockets.html, this endpoint is used to connect to the server.
	app.Get("/api/ws", ws.Handler())

	ws.OnConnection(OnConnect)
	app.Run(iris.Addr("127.0.0.1:8080"))
}

// OnDisconnect clean up things when a client is disconnected
func OnDisconnect(c websocket.Connection) {
	session.GlobalSessionMgr.DeleteSessionValue(c)
	fmt.Println("OnDisconnect(): client disconnected!")
}

func main() {

	ServerLoop()

}