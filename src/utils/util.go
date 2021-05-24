package utils

import (
	"bytes"
	"math/rand"
	"encoding/json"
)

var JsonId  = 0

func RandInt(min int , max int) int {
	return min + rand.Intn(max-min)
}
func RandomString (l int ) string {
	var result   bytes.Buffer
	var temp string
	for i:=0 ; i<l ;  {
		temp = string(RandInt(65,90))
		result.WriteString(temp)
		i++

	}
	return result.String()
}

func GetJsonId() (int){
	JsonId++
	return JsonId
}


type JsonRpcRequest struct {
	Id      int           `json:"id"`
	JsonRpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type JsonRpcResponse struct {
	Id     int          `json:"id"`
	Result *interface{} `json:"result"`
	Error  *Error       `json:"error"`
}


func DecodeRequest(message string) (int,string,error) {
	var jsonRpcRequest JsonRpcRequest
	err := json.Unmarshal([]byte(message), &jsonRpcRequest)
	if err != nil {
		return 0, "",err
	}
	return jsonRpcRequest.Id, jsonRpcRequest.Method, nil

}
