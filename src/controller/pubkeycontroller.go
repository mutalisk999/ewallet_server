package controller

import (
	"encoding/json"
	"fmt"
	"github.com/kataras/iris/websocket"
	"model"
	"session"
	"utils"
)

type PubKey struct {
	ServerName string    `json:"servername"`
	StartIndex int       `json:"startindex"`
	Keys map[int]string          `json:"keys"`
}

type InitPubKeyRequest struct {
	Id      int                 `json:"id"`
	JsonRpc string              `json:"jsonrpc"`
	Method  string              `json:"method"`
	Params  []PubKey          `json:"params"`
}

type InitPubKeyResponseParams struct{
	ServerId int    `json:"serverid"`
}

type InitPubKeyResponse struct {
	Id     int                            `json:"id"`
	Result *InitPubKeyResponseParams	  `json:"result"`
	Error  *utils.Error                   `json:"error"`
}

type QueryPubKeysRequest struct {
	Id      int                 `json:"id"`
	JsonRpc string              `json:"jsonrpc"`
	Method  string              `json:"method"`
	Params  PubKey              `json:"params"`
}

type QueryPubKeysResponse struct {
	Id     int                    `json:"id"`
	Result []QueryPubKeysResponseParam			  `json:"result"`
	Error  *utils.Error           `json:"error"`
}

type QueryPubKeysResponseParam struct {
	ServerId int `json:"serverid"`
	ServerName string `json:"servername"`
	StartIndex int `json:"startindex"`
	Keys map[int]string `json:"keys"`
}

func OnInitPubKey(id int,message string,c websocket.Connection){
	error_return :=func(res InitPubKeyResponse,error_code int,errArgs ...interface{})[]byte{
		res.Error = utils.MakeError(error_code,errArgs)
		res.Result = nil
		res_data,_ := json.Marshal(res)
		return res_data
	}

	var req InitPubKeyRequest
	err := json.Unmarshal([]byte(message), &req)
	var res InitPubKeyResponse
	res.Id = id

	if err != nil {
		c.EmitMessage(error_return(res,900000,err.Error()))
		return
	}
	_,exist :=session.GlobalSessionMgr.GetSessionValue(c)
	if !exist{
		c.EmitMessage(error_return(res,100002,err.Error()))
		return
	}

	var serverName = req.Params[0].ServerName
	var startIndex = req.Params[0].StartIndex
	serverId,err := model.GlobalDBMgr.ServerInfoMgr.CreateServerInfo(serverName,startIndex)
	if err!=nil {
		c.EmitMessage(error_return(res,900000,err.Error()))
		return
	}

	if serverId < 0 {
		c.EmitMessage(error_return(res,900000,"Serverid is error when initial pubkey"))
		return
	}

	pubkeys := make([]map[string]interface{},0,len(req.Params[0].Keys))

	for keyIndex,pubkey := range req.Params[0].Keys{

		one_pubkey := make(map[string]interface{})
		one_pubkey["serverid"] = serverId
		one_pubkey["keyindex"] = keyIndex
		one_pubkey["pubkey"] = pubkey
		pubkeys = append(pubkeys,one_pubkey)
		ifSuccess,err := model.GlobalDBMgr.ServerPubkeyPoolMgr.InsertManyPubkey(pubkeys)
		if err!=nil {
			c.EmitMessage(error_return(res,900000,err.Error()))
			return
		}

		if ifSuccess != true{
			c.EmitMessage(error_return(res,900000,"An unknown error occurred when insert table pubkey."))
			return
		}
	}

	var resParam InitPubKeyResponse
	resParam.Result.ServerId = serverId
	res.Error = nil
	res_data,_ := json.Marshal(res)
	c.EmitMessage(res_data)
	return

}

func OnQueryPubKeys(id int,message string,c websocket.Connection){
	error_return :=func(res QueryPubKeysResponse,error_code int,errArgs ...interface{})[]byte{
	res.Error = utils.MakeError(error_code,errArgs)
	res_data,_ := json.Marshal(res)
	return res_data
	}

	var req QueryPubKeysRequest
	err := json.Unmarshal([]byte(message), &req)
	var res QueryPubKeysResponse
	res.Id = id

	if err != nil {
	c.EmitMessage(error_return(res,900000,err.Error()))
	return
	}
	_,exist :=session.GlobalSessionMgr.GetSessionValue(c)
	if !exist{
	c.EmitMessage(error_return(res,100002,err.Error()))
	return
	}

	pubkey_datas,err := model.GlobalDBMgr.ServerPubkeyPoolMgr.GetAllPubkey()
	if err != nil {
	c.EmitMessage(error_return(res,900000,err.Error()))
	return
	}

	pubkeyResMap := make(map[int]QueryPubKeysResponseParam)
	for _,one_pubkey := range pubkey_datas{
	var query_data QueryPubKeysResponseParam
	query_data.ServerId = int(one_pubkey.Serverid)
	querydata, ok := pubkeyResMap[query_data.ServerId]
	if ok{
		querydata.Keys[one_pubkey.Keyindex] = one_pubkey.Pubkey
		pubkeyResMap[query_data.ServerId] = querydata
	}else{
		one_serverinfo,err:= model.GlobalDBMgr.ServerInfoMgr.GetServerInfoByServerID(one_pubkey.Serverid)
		if err != nil {
			c.EmitMessage(error_return(res,900000,err.Error()))
			return
		}
		query_data.ServerName = one_serverinfo.Servername
		query_data.StartIndex = one_serverinfo.Serverstartindex

		query_data.Keys = make(map[int]string)
		query_data.Keys[one_pubkey.Keyindex] = one_pubkey.Pubkey
		pubkeyResMap[query_data.ServerId] = query_data

	}
	}

	for serverID := range pubkeyResMap{
		res.Result = append(res.Result,pubkeyResMap[serverID])
	}
	res_data,_ := json.Marshal(res)
	c.EmitMessage(res_data)
	return
}



func OnPubKey(message string,c websocket.Connection){
	id,method,err := utils.DecodeRequest(message)
	var res utils.JsonRpcResponse
	if err!=nil{
		fmt.Println(err.Error())
		res.Id = 0
		res.Result= nil
		res.Error= utils.MakeError(100000)
		emit_message,_ := json.Marshal(res)
		c.EmitMessage(emit_message)
	}

	if method == "init_pubkey"{
		OnInitPubKey(id,message,c)
	}else if method == "query_pubkeys"{
		OnQueryPubKeys(id,message,c)
	}else {
		res.Id = id
		res.Result =nil
		res.Error = utils.MakeError(200000)
		emit_message,_ := json.Marshal(res)
		c.EmitMessage(emit_message)
	}

}