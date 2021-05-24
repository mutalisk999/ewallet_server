package controller

import (


	"github.com/kataras/iris/websocket"
	"session"
	"fmt"
	"encoding/json"
	"utils"
	"github.com/kataras/iris/core/errors"
	"model"
	"coin"
	"encoding/hex"
)


type QueryMessageParam struct {

}

type QueryMessageRequest struct {
	Id      int                 `json:"id"`
	JsonRpc string              `json:"jsonrpc"`
	Method  string              `json:"method"`
	Params  []QueryMessageParam `json:"params"`
}

type QueryMessageResponse struct {
	Id     int                    `json:"id"`
	Result string `json:"result"`
	Error  *utils.Error           `json:"error"`
}

func OnQueryMessage(message string,c websocket.Connection){
	var req QueryMessageRequest
	sessionValue,err := session.GlobalSessionMgr.NewNullSessionValue(c)
	if err!=nil{
		fmt.Println(err.Error())
		var res QueryMessageResponse
		res.Id = req.Id
		res.Error = utils.MakeError(900000, err.Error())
		resData,_ := json.Marshal(res)
		c.EmitMessage(resData)
		fmt.Println("OnQueryMessage",req,res)
		return
	}

	err = json.Unmarshal([]byte(message),&req)
	if err!=nil {
		// can not reach here
		fmt.Println(err.Error())
		var res utils.JsonRpcResponse
		res.Id = 0
		res.Result= nil
		res.Error= utils.MakeError(100000)
		emitMessage,_ := json.Marshal(res)
		c.EmitMessage(emitMessage)
		return
	}

	var res QueryMessageResponse
	res.Id = req.Id
	res.Error = nil
	res.Result = sessionValue.Random
	resData,_ := json.Marshal(res)
	c.EmitMessage(resData)
	fmt.Println("OnQueryMessage",req,res)
}


type VerifyMessageParam struct {
	ServerId  int    `json:"serverid"`
	SignedMessage string `json:"signedmessage"`
}

type VerifyMessageRequest struct {
	Id      int                 `json:"id"`
	JsonRpc string              `json:"jsonrpc"`
	Method  string              `json:"method"`
	Params  []VerifyMessageParam `json:"params"`
}

type VerifyMessageResponse struct {
	Id     int                    `json:"id"`
	Result bool 	`json:"result"`
	Error  *utils.Error           `json:"error"`
}

func OnVerifyMessage(message string,c websocket.Connection){
	var req VerifyMessageRequest
	sessionValue, hasSession := session.GlobalSessionMgr.GetSessionValue(c)
	if !hasSession{
		fmt.Println("OnVerifyMessage: can not get sessionValue")
		var res QueryMessageResponse
		res.Id = req.Id
		res.Error = utils.MakeError(900000, errors.New("can not get session value"))
		resData,_ := json.Marshal(res)
		c.EmitMessage(resData)
		fmt.Println("OnVerifyMessage",req,res)
		return
	}

	err := json.Unmarshal([]byte(message),&req)
	if err!=nil {
		// can not reach here
		fmt.Println(err.Error())
		var res utils.JsonRpcResponse
		res.Id = 0
		res.Result= nil
		res.Error= utils.MakeError(100000)
		emitMessage,_ := json.Marshal(res)
		c.EmitMessage(emitMessage)
		return
	}

	if len(req.Params) != 1 {
		var res utils.JsonRpcResponse
		res.Id = req.Id
		res.Result= nil
		res.Error = utils.MakeError(100001)
		emitMessage,_ := json.Marshal(res)
		c.EmitMessage(emitMessage)
		return
	}

	verifyResult := false
	pubkey, err := model.GlobalDBMgr.ServerPubkeyPoolMgr.GetPubkeyByIdIndex(1, req.Params[0].ServerId)
	if err != nil {
		var res utils.JsonRpcResponse
		res.Id = req.Id
		res.Result= nil
		res.Error = utils.MakeError(300001, model.GlobalDBMgr.ServerPubkeyPoolMgr.TableName, "query", "query verify message pubkey")
		emitMessage,_ := json.Marshal(res)
		c.EmitMessage(emitMessage)
		return
	}

	pubkeyBytes, err := hex.DecodeString(pubkey.Pubkey)
	if err != nil {
		var res utils.JsonRpcResponse
		res.Id = req.Id
		res.Result= nil
		res.Error = utils.MakeError(600000, "pubkey")
		emitMessage,_ := json.Marshal(res)
		c.EmitMessage(emitMessage)
		return
	}

	signedBytes, err := hex.DecodeString(req.Params[0].SignedMessage)
	if err != nil {
		var res utils.JsonRpcResponse
		res.Id = req.Id
		res.Result= nil
		res.Error = utils.MakeError(600000, "signedmessage")
		emitMessage,_ := json.Marshal(res)
		c.EmitMessage(emitMessage)
		return
	}

	// verify signature
	verifyResult, err = coin.CoinVerifyTrx2(pubkeyBytes, []byte(sessionValue.Random), signedBytes)
	if err != nil {
		var res utils.JsonRpcResponse
		res.Id = req.Id
		res.Result= nil
		res.Error = utils.MakeError(600001)
		emitMessage,_ := json.Marshal(res)
		c.EmitMessage(emitMessage)
		return
	}

	var res VerifyMessageResponse
	res.Id = req.Id
	res.Error = nil
	res.Result = verifyResult
	resData,_ := json.Marshal(res)
	c.EmitMessage(resData)
	fmt.Println("OnVerifyMessage",req,res)

	if verifyResult {
		// update session
		sessionValue.ServerId = req.Params[0].ServerId
		sessionValue.IsLogin = true
		session.GlobalSessionMgr.UpdateSessionValue(sessionValue)
	}
}

func OnLogin(message string,c websocket.Connection){
	id,method,err := utils.DecodeRequest(message)
	var res utils.JsonRpcResponse
	if err!=nil{
		fmt.Println(err.Error())
		res.Id = 0
		res.Result= nil
		res.Error= utils.MakeError(100000)
		emitMessage,_ := json.Marshal(res)
		c.EmitMessage(emitMessage)
	}

	if method == "query_message"{
		OnQueryMessage(message,c)
	} else if method == "verify_message" {
		OnVerifyMessage(message,c)
	}else {
		res.Id = id
		res.Result =nil
		res.Error = utils.MakeError(200000)
		emitMessage,_ := json.Marshal(res)
		c.EmitMessage(emitMessage)
	}

}



