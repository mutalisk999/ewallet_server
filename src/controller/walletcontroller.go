package controller


import (


"github.com/kataras/iris/websocket"
"utils"
"fmt"
"encoding/json"
	"model"
	"session"
	"strings"
	"strconv"
)

type WalletCreateParam struct {
	CoinSymbol string     	`json:"coinsymbol"`
	NeedSigCount int		`json:"needsigcount"`
	TotalCount  int			`json:"totalcount"`
	KeyDetail []map[string]int	`json:"keydetail"`
	WalletName string		`json:"walletname"`
	Address string			`json:"address"`
	Fee    string			`json:"fee"`
	GasPrice string			`json:"gasprice"`
	GasLimit string			`json:"gaslimit"`
}

type WalletCreateRequest struct {
	Id      int                 `json:"id"`
	JsonRpc string              `json:"jsonrpc"`
	Method  string              `json:"method"`
	Params  WalletCreateParam `json:"params"`
}

type WalletCreateResParam struct {
	WalletId int `json:"walletid"`
}

type WalletCreateResponse struct {
	Id     int                    `json:"id"`
	Result *WalletCreateResParam `json:"result"`
	Error  *utils.Error           `json:"error"`
}



func OnCreateWallet(id int,message string,c websocket.Connection){
	error_return :=func(res WalletCreateResponse,error_code int,errArgs ...interface{})[]byte{
		res.Error = utils.MakeError(error_code,errArgs)
		res.Result = nil
		res_data,_ := json.Marshal(res)
		return res_data
	}


	var req WalletCreateRequest
	err := json.Unmarshal([]byte(message), &req)
	var res WalletCreateResponse
	res.Id = id

	if err != nil {
		c.EmitMessage(error_return(res,900000,err.Error()))
		return
	}
	sessionValue,exist :=session.GlobalSessionMgr.GetSessionValue(c)
	if !exist{
		c.EmitMessage(error_return(res,100002,err.Error()))
		return
	}

	coinSymbol,err := model.GlobalDBMgr.ServerCoinConfigMgr.GetCoinBySymbol(req.Params.CoinSymbol)
	if err!=nil{
		c.EmitMessage(error_return(res,900000,err.Error()))
		return
	}
	if req.Params.TotalCount != len(req.Params.KeyDetail){
		c.EmitMessage(error_return(res,500000,req.Params.TotalCount,len(req.Params.KeyDetail)))
		return
	}
	wallet_data,err := model.GlobalDBMgr.ServerWalletConfigMgr.GetWalletByName(req.Params.WalletName)
	if err!=nil{
		c.EmitMessage(error_return(res,900000,err.Error()))
		return
	}
	if wallet_data !=nil{
		c.EmitMessage(error_return(res,500001,req.Params.WalletName))
		return
	}
	keys := make([]string,0,len(req.Params.KeyDetail))
	for _,key_pair := range req.Params.KeyDetail{
		//create keys string check key state
		keyindex := key_pair["keyindex"]
		serverid := key_pair["serverid"]
		key,err := model.GlobalDBMgr.ServerPubkeyPoolMgr.GetPubkeyByIdIndex(keyindex,serverid)
		if err!=nil{
			c.EmitMessage(error_return(res,900000,err.Error()))
			return
		}
		if key ==nil{
			c.EmitMessage(error_return(res,500001,req.Params.WalletName))
			return
		}

		keys = append(keys,strconv.Itoa(serverid)+":"+strconv.Itoa(keyindex))
	}
	//Todo: check address and get redeemscript

	walletid,err:=model.GlobalDBMgr.ServerWalletConfigMgr.CreateWallet(coinSymbol.Coinid,req.Params.WalletName,req.Params.Address,strings.Join(keys,","),req.Params.NeedSigCount,
	req.Params.TotalCount,req.Params.Fee,req.Params.GasPrice,req.Params.GasLimit,sessionValue.ServerId)
	if err!=nil{
		c.EmitMessage(error_return(res,900000,err.Error()))
		return
	}
	if walletid ==-1{
		c.EmitMessage(error_return(res,900000,"An unknown error occurred when wallet create"))
		return
	}
	var resParam WalletCreateResParam
	resParam.WalletId = walletid
	res.Error = nil
	res_data,_ := json.Marshal(res)
	c.EmitMessage(res_data)
	return

}


//Query wallet

type QueryWalletParam struct {
	CoinSymbol []string     `json:"coinsymbol"`
	WalletIds	[]int		`json:"walletids"`
}

type QueryWalletRequest struct {
	Id      int                 `json:"id"`
	JsonRpc string              `json:"jsonrpc"`
	Method  string              `json:"method"`
	Params  QueryWalletParam `json:"params"`
}

type QueryWalletResParam struct {
	WalletId int `json:"walletid"`
	CoinSymbol string `json:"coinsymbol"`
	NeedSigCount int `json:"needsigcount"`
	TotalCount int `json:"totalcount"`
	KeyDetail []map[string]int `json:"keydetail"`
	WalletName string `json:"walletname"`
	Address string `json:"address"`
	CreateServerId int `json:"createserverid"`
	Fee string `json:"fee"`
	GasPrice string `json:"gasprice"`
	GasLimit string `json:"gaslimit"`
	Status int `json:"status"`

}

type QueryWalletResponse struct {
	Id     int                    `json:"id"`
	Result []QueryWalletResParam `json:"result"`
	Error  *utils.Error           `json:"error"`
}



func OnQueryWallet(id int,message string,c websocket.Connection){
	error_return :=func(res QueryWalletResponse,error_code int,errArgs ...interface{})[]byte{
		res.Error = utils.MakeError(error_code,errArgs)
		res_data,_ := json.Marshal(res)
		return res_data
	}


	var req QueryWalletRequest
	err := json.Unmarshal([]byte(message), &req)
	var res QueryWalletResponse
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

	wallet_datas,err := model.GlobalDBMgr.ServerWalletConfigMgr.QueryWallet(req.Params.CoinSymbol,req.Params.WalletIds)
	if err != nil {
		c.EmitMessage(error_return(res,900000,err.Error()))
		return
	}

	for _,one_wallet := range wallet_datas{
		var query_data QueryWalletResParam
		query_data.WalletId = one_wallet.Walletid
		one_coin,err:= model.GlobalDBMgr.ServerCoinConfigMgr.GetCoin(one_wallet.Coinid)
		if err != nil {
			c.EmitMessage(error_return(res,900000,err.Error()))
			return
		}
		query_data.CoinSymbol = one_coin.Coinsymbol
		query_data.NeedSigCount = one_wallet.Needsigcount
		query_data.TotalCount = one_wallet.Keycount
		json.Unmarshal([]byte(one_wallet.Keys),&query_data.KeyDetail)
		query_data.WalletName =one_wallet.Walletname
		query_data.Address = one_wallet.Address
		ids := strings.Split(one_wallet.Confirmserverids,",")
		if len(ids) <1{
			c.EmitMessage(error_return(res,900000,"confirm server ids is null.Please system safety!"))
			return
		}
		serverid,_ := strconv.ParseInt(ids[0],32,10)
		query_data.CreateServerId =int(serverid)
		query_data.Fee = one_wallet.Fee
		query_data.GasPrice = one_wallet.Gasprice
		query_data.GasLimit = one_wallet.Gaslimit
		query_data.Status = one_wallet.State
		res.Result = append(res.Result,query_data)
	}
	res_data,_ := json.Marshal(res)
	c.EmitMessage(res_data)
	return

}



func OnWallet(message string,c websocket.Connection){
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

	if method == "create_wallet"{
		OnCreateWallet(id,message,c)
	}else if method == "query_wallet"{
		OnQueryWallet(id,message,c)
	}else {
		res.Id = id
		res.Result =nil
		res.Error = utils.MakeError(200000)
		emit_message,_ := json.Marshal(res)
		c.EmitMessage(emit_message)
	}

}