package utils

import (
	"sync"
	"github.com/kataras/iris/websocket"
	"fmt"
	xwebsocket "golang.org/x/net/websocket"
	"encoding/json"
	"time"
	"github.com/kataras/iris/core/errors"
)

var GlobalReqMap sync.Map
func RequestFuture(index int) *chan interface{}{
	future := make(chan interface{},1)
	GlobalReqMap.Store(index,&future)

	return &future
}

func Request(WS *xwebsocket.Conn, method, message string)(interface{},error){
	var req JsonRpcRequest
	err := json.Unmarshal([]byte(message),&req)
	if err!=nil{
		return nil,err
	}
	req.Id = GetJsonId()
	data,_ := json.Marshal(req)
	err = SendMessage(WS,method,string(data))
	if err!=nil{
		return nil,err
	}
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(10*time.Second)	//等待1秒钟
		timeout <- true
	}()

	future := RequestFuture(req.Id)
	select {
	case res_data:=<- *future:
		return res_data,nil
	case <- timeout:
		return nil,errors.New("request timeout!")
	}


}

// SendMessage broadcast a message to server
func SendMessage(WS *xwebsocket.Conn, method, message string) error {
	buffer := []byte(message)
	return SendtBytes(WS,method, buffer)
}

// SendtBytes broadcast a message to server
func SendtBytes(WS *xwebsocket.Conn, method string, message []byte) error {
	// look https://github.com/kataras/iris/blob/master/websocket/message.go , client.go and client.js
	// to understand the buffer line:
	buffer := []byte(fmt.Sprintf("%s%v;0;", websocket.DefaultEvtMessageKey, method))
	buffer = append(buffer, message...)
	_, err := WS.Write(buffer)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}