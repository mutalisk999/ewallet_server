package controller

import (
	"encoding/json"
	"fmt"
	"github.com/kataras/iris/websocket"
	"model"
	"session"
	"utils"
	"strings"
)

type ToDetailStruct struct {
	To_address string        `json:"to_address"`
	Value string        `json:"value"`
}

type CreateTransactionParams struct{
	CoinSymbol string        `json:"coinsymbol"`
	ServerId   int           `json:"serverid"`
	WalletId   int           `json:"walletid"`
	Fromaddress   string           `json:"fromaddress"`
    Todetail   []ToDetailStruct `json:"todetail"`
	Totalfee   string `json:"totalfee"`
	Signedtrx   []string `json:"signedtrx"`
}

type CreateTransactionRequest struct {
	Id      int                 `json:"id"`
	JsonRpc string              `json:"jsonrpc"`
	Method  string              `json:"method"`
	Params  []CreateTransactionParams          `json:"params"`
}

type CreateTransactionResponse struct {
	Id     int                            `json:"id"`
	Result *CreateTransactionResponseParams	  `json:"result"`
	Error  *utils.Error                   `json:"error"`
}

type CreateTransactionResponseParams struct{
	Taskid int    `json:"taskid"`
}

type QueryTransactionRequestParams struct{
	Taskid int    `json:"taskid"`
}

type QueryTransactionRequest struct {
	Id      int                 `json:"id"`
	JsonRpc string              `json:"jsonrpc"`
	Method  string              `json:"method"`
	Params  QueryTransactionRequestParams     `json:"params"`
}

type QueryTransactionParams struct{
	CoinSymbol string        `json:"coinsymbol"`
	ServerId   int           `json:"serverid"`
	WalletId   int           `json:"walletid"`
	Fromaddress   string           `json:"fromaddress"`
	Todetail   []ToDetailStruct `json:"todetail"`
	Totalfee   string `json:"totalfee"`
	Signedtrx   []string `json:"signedtrx"`
	Createserverid   int `json:"createserverid"`
	Status   int `json:"status"`
	Signedserverids   []string `json:"signedserverids"`
}

type QueryTransactionResponse struct {
	Id     int                            `json:"id"`
	Result *QueryTransactionParams	  `json:"result"`
	Error  *utils.Error                   `json:"error"`
}

type QueryRelatedTransactionRequestParams struct{
	Serverid int    `json:"serverid"`
	Taskid []int    `json:"taskid"`
	Status []int    `json:"status"`
}

type QueryRelatedTransactionRequest struct {
	Id      int                 `json:"id"`
	JsonRpc string              `json:"jsonrpc"`
	Method  string              `json:"method"`
	Params  QueryRelatedTransactionRequestParams     `json:"params"`
}

type QueryRelatedTransactionResponse struct {
	Id     int                            `json:"id"`
	Result []QueryTransactionParams	  `json:"result"`
	Error  *utils.Error                   `json:"error"`
}

type ConfirmTransactionRequestParams struct{
	Taskid int    `json:"taskid"`
	Signedtrx string    `json:"signedtrx"`
}
type ConfirmTransactionRequest struct {
	Id      int                 `json:"id"`
	JsonRpc string              `json:"jsonrpc"`
	Method  string              `json:"method"`
	Params  []ConfirmTransactionRequestParams          `json:"params"`
}

type ConfirmTransactionResponse struct {
	Id     int                            `json:"id"`
	Result bool	  `json:"result"`
	Error  *utils.Error                   `json:"error"`
}

type RejectTransactionRequestParams struct{
	Taskid int    `json:"taskid"`
}
type RejectTransactionRequest struct {
	Id      int                 `json:"id"`
	JsonRpc string              `json:"jsonrpc"`
	Method  string              `json:"method"`
	Params  []RejectTransactionRequestParams          `json:"params"`
}

type RejectTransactionResponse struct {
	Id     int                            `json:"id"`
	Result bool	  `json:"result"`
	Error  *utils.Error                   `json:"error"`
}

func CreateTransaction(id int,message string,c websocket.Connection){
	error_return :=func(res CreateTransactionResponse,error_code int,errArgs ...interface{})[]byte{
		res.Error = utils.MakeError(error_code,errArgs)
		res_data,_ := json.Marshal(res)
		return res_data
	}
	var req CreateTransactionRequest
	err := json.Unmarshal([]byte(message), &req)
	var res CreateTransactionResponse
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

}

func QueryTransaction(id int,message string,c websocket.Connection){
	error_return :=func(res QueryTransactionResponse,error_code int,errArgs ...interface{})[]byte{
		res.Error = utils.MakeError(error_code,errArgs)
		res_data,_ := json.Marshal(res)
		return res_data
	}

	var req QueryTransactionRequest
	err := json.Unmarshal([]byte(message), &req)
	var res QueryTransactionResponse
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

	trxid,err := model.GlobalDBMgr.ServerTaskMgr.GetTrxidByTaskid(req.Params.Taskid)
	if err != nil {
		c.EmitMessage(error_return(res,900000,err.Error()))
		return
	}

	transaction,err := model.GlobalDBMgr.ServerTransactionMgr.GetTransactionById(trxid)
	if err != nil {
		c.EmitMessage(error_return(res,900000,err.Error()))
		return
	}

	coin,err := model.GlobalDBMgr.ServerCoinConfigMgr.GetCoin(transaction.Coinid)
	if err != nil {
		c.EmitMessage(error_return(res,900000,err.Error()))
		return
	}

	res.Result.ServerId = transaction.Serverid
	res.Result.Fromaddress = transaction.Fromaddr
	res.Result.CoinSymbol = coin.Coinsymbol
	res.Result.Signedtrx = strings.SplitAfter(transaction.Signedtrxs,",")
	res.Result.Totalfee = transaction.Fee
	res.Result.WalletId = transaction.Walletid
	res.Result.Createserverid = transaction.Serverid
	res.Result.Status = transaction.State
	res.Result.Signedserverids = strings.SplitAfter(transaction.Serverfirmed,",")

	var to []string
	to = strings.Split(transaction.Todetails,",")
	for i := 0;i < len(to);i++{
        var oneto []string = strings.Split(to[i],":")
        res.Result.Todetail[i].To_address = oneto[0]
        res.Result.Todetail[i].Value = oneto[1]
	}

	res_data,_ := json.Marshal(res)
	c.EmitMessage(res_data)
	return
}

func QueryRelatedTransaction(id int,message string,c websocket.Connection){
	error_return :=func(res QueryRelatedTransactionResponse,error_code int,errArgs ...interface{})[]byte{
		res.Error = utils.MakeError(error_code,errArgs)
		res_data,_ := json.Marshal(res)
		return res_data
	}

	var req QueryRelatedTransactionRequest
	err := json.Unmarshal([]byte(message), &req)
	var res QueryRelatedTransactionResponse
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

	var taskid []int = req.Params.Taskid
	var status []int = req.Params.Status
	transactions,err := model.GlobalDBMgr.ServerTransactionMgr.GetTransactionsByServerIdAndTaskidAndState(req.Params.Serverid,taskid,status)
	if err != nil {
		c.EmitMessage(error_return(res,900000,err.Error()))
		return
	}

	for _,onetransaction := range transactions{
		var res_onetrx QueryTransactionParams
		res_onetrx.ServerId = onetransaction.Serverid
		res_onetrx.Fromaddress = onetransaction.Fromaddr
		coin,err := model.GlobalDBMgr.ServerCoinConfigMgr.GetCoin(onetransaction.Coinid)
		if err != nil {
			c.EmitMessage(error_return(res,900000,err.Error()))
			return
		}
		res_onetrx.CoinSymbol = coin.Coinsymbol
		res_onetrx.Signedtrx = strings.SplitAfter(onetransaction.Signedtrxs,",")
		res_onetrx.Totalfee = onetransaction.Fee
		res_onetrx.WalletId = onetransaction.Walletid
		res_onetrx.Createserverid = onetransaction.Serverid
		res_onetrx.Status = onetransaction.State
		res_onetrx.Signedserverids = strings.SplitAfter(onetransaction.Serverfirmed,",")

		var to []string
		to = strings.Split(onetransaction.Todetails,",")
		for i := 0;i < len(to);i++{
			var oneto []string = strings.Split(to[i],":")
			res_onetrx.Todetail[i].To_address = oneto[0]
			res_onetrx.Todetail[i].Value = oneto[1]
		}

		res.Result = append(res.Result,res_onetrx)
	}

	res.Error = nil
	res_data,_ := json.Marshal(res)
	c.EmitMessage(res_data)
	return
}

func ConfirmTransaction(id int,message string,c websocket.Connection){

}

func RejectTransaction(id int,message string,c websocket.Connection){

}

func OnTransaction(message string,c websocket.Connection){
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

	if method == "create_transaction"{
		CreateTransaction(id,message,c)
	}else if method == "query_transaction"{
		QueryTransaction(id,message,c)
	}else if method == "confirm_transaction"{
		ConfirmTransaction(id,message,c)
	}else if method == "reject_transaction"{
		RejectTransaction(id,message,c)
	}else{
		res.Id = id
		res.Result =nil
		res.Error = utils.MakeError(200000)
		emit_message,_ := json.Marshal(res)
		c.EmitMessage(emit_message)
	}

}