package coin

import (
	"bytes"
	"config"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ybbus/jsonrpc"
	"math/big"
	"strconv"
	"strings"
)

type ERC20Agent struct {
	ServerUrl string
}

func (agent *ERC20Agent) Init(urlstr string) {
	agent.ServerUrl = urlstr
}
func (agent *ERC20Agent) DoHttpJsonRpcCallType1(method string, args ...interface{}) (*jsonrpc.RPCResponse, error) {
	rpcClient := jsonrpc.NewClient(agent.ServerUrl)
	rpcResponse, err := rpcClient.Call(method, args)
	if err != nil {
		return nil, err
	}
	return rpcResponse, nil
}

func (agent *ERC20Agent) CreateERC20Transaction(coinSymbol, from, to, gas, gasprice, value, data string, keyindex uint16) (string, error) {
	nonce, err := agent.GetTransactionCount(from, "pending")
	if err != nil {
		return "", err
	}
	to_address := common.HexToAddress(config.GlobalSupportCoinMgr[coinSymbol].ContractAddress)
	AccountNonce := uint64(nonce)
	Price := big.NewInt(0)
	Price.SetString(gasprice, 10)
	GasLimit := big.NewInt(0)
	GasLimit.SetString(gas, 10)
	Amount := big.NewInt(0)
	var Payload []byte
	if data == "" {
		Value := big.NewInt(0)
		amountStr := ConvertStringToBigNumber(coinSymbol, value)
		Value.SetString(amountStr, 10)
		valueData := math.PaddedBigBytes(Value, 32)
		//valueObj := math.HexOrDecimal256(*Value)
		//valueStr,err := valueObj.MarshalText()
		//if err !=nil{
		//	fmt.Println(err)
		//	return "",err
		//}
		if strings.HasPrefix(to, "0x") {
			to = to[2:]
		}
		tmp_input := "0xa9059cbb" + "000000000000000000000000" + to + hex.EncodeToString(valueData)

		Payload, err = hex.DecodeString(tmp_input[2:])
		if err != nil {
			fmt.Println("payload", err)
			return "", err
		}
		fmt.Println(value)
		est_gas, err := agent.EstimateGas(config.GlobalSupportCoinMgr[coinSymbol].ContractAddress, "0x0", tmp_input)
		if err != nil {
			fmt.Println("111", err)
			return "", err
		}
		if est_gas.Cmp(GasLimit) > 0 {
			fmt.Println("est fee large than input!", est_gas, GasLimit)
			GasLimit = est_gas.Add(est_gas, big.NewInt(10000))
		}
	} else {
		Payload, err = hex.DecodeString(data)
	}

	//Payload,err := hex.DecodeString(data)
	//types.NewEIP155Signer() if change to formal chain  must use EIP155 and set chainid
	var signer = types.HomesteadSigner{}
	my_trx := types.NewTransaction(AccountNonce, to_address, Amount, GasLimit.Uint64(), Price, Payload)
	hash_data := signer.Hash(my_trx)
	success := 0
	for {
		fmt.Println(hash_data)
		res, err := CoinSignTrx('1', hash_data.Bytes(), keyindex)
		fmt.Println(res)
		sigdata := make([]byte, 0, 65)
		sigdata = append(sigdata, res...)
		for i := 0; i < 1; i++ {
			sigdata = append(sigdata, byte(i))
			my_trx, err = my_trx.WithSignature(signer, sigdata)
			if err != nil {
				fmt.Println("withsignature", err.Error())
			}
			sig_from, err := signer.Sender(my_trx)
			if err != nil {
				fmt.Println(err.Error())
			}
			if sig_from.Hex() == from {
				success = 1
				break
			}
			fmt.Println("from:", sig_from.Hex())
		}
		if success > 0 {
			break
		}

	}

	raw_writer := new(bytes.Buffer)
	my_trx.EncodeRLP(raw_writer)
	fmt.Println("sign_data: ", hex.EncodeToString(raw_writer.Bytes()))

	return "0x" + hex.EncodeToString(raw_writer.Bytes()), nil
}

//cal call contract cost
func (agent *ERC20Agent) EstimateGas(to string, value string, data string) (*big.Int, error) {
	if to == "" {
		to = "0x0"
	}
	res, err := agent.DoHttpJsonRpcCallType1("eth_estimateGas", map[string]string{"to": to, "value": value, "data": data})
	if err != nil {
		return big.NewInt(0), err
	}
	esti_fee := big.NewInt(0)

	esti_fee_str, err := res.GetString()
	fmt.Println(esti_fee_str)
	if err != nil {
		return big.NewInt(0), err
	}
	esti_fee.SetString(esti_fee_str[2:], 16)

	return esti_fee, nil
}

func (agent *ERC20Agent) GetBalanceByAddress(addr string) (*big.Int, error) {

	res, err := agent.DoHttpJsonRpcCallType1("eth_getBalance", addr, "latest")
	if err != nil {
		return big.NewInt(0), err
	}
	balance := big.NewInt(0)

	balance_str, err := res.GetString()
	if err != nil {
		return big.NewInt(0), err
	}
	balance.SetString(balance_str[2:], 16)

	return balance, nil
}

func (agent *ERC20Agent) GetERC20BalanceByAddress(coinSymbol, addr string) (*big.Int, *big.Int, error) {

	input := "0x70a08231" + "000000000000000000000000" + addr[2:]
	contract_address := config.GlobalSupportCoinMgr[coinSymbol].ContractAddress
	eth_call_obj := map[string]string{"to": contract_address, "data": input}

	blockdata, err := agent.DoHttpJsonRpcCallType1("eth_call", eth_call_obj, "latest")
	if err != nil {
		fmt.Println(err)
		return big.NewInt(0), big.NewInt(0), err
	}
	fmt.Println(blockdata)
	balance_str, _ := blockdata.GetString()
	fmt.Println(balance_str)
	erc20_balance := big.NewInt(0)
	erc20_balance.SetString(balance_str[2:], 16)

	balance, err := agent.GetBalanceByAddress(addr)
	if err != nil {
		return erc20_balance, big.NewInt(0), err
	}

	return erc20_balance, balance, nil
}

func (agent *ERC20Agent) GetTransactionRealCost(coinSymbol, trxId string) (string, error) {

	res, err := agent.DoHttpJsonRpcCallType1("eth_getTransactionReceipt", trxId)
	if err != nil {
		return "0", err
	}
	gasUsed := big.NewInt(0)
	gasUsed_str, exist := res.Result.(map[string]interface{})["gasUsed"].(string)
	if exist != true {
		return "0", err
	}
	gasUsed.SetString(gasUsed_str[2:], 16)

	res, err = agent.DoHttpJsonRpcCallType1("eth_getTransactionByHash", trxId)
	if err != nil {
		return "0", err
	}
	gasPrice := big.NewInt(0)
	gasPrice_str, exist := res.Result.(map[string]interface{})["gasPrice"].(string)
	if exist != true {
		return "0", err
	}
	gasPrice.SetString(gasPrice_str[2:], 16)
	totalCost := big.NewInt(0)
	totalCost.Mul(gasPrice, gasUsed)
	return ConvertBigNumberToString(coinSymbol, totalCost), nil
}

//tag latest pending
func (agent *ERC20Agent) GetTransactionCount(addr string, tag string) (int64, error) {
	res, err := agent.DoHttpJsonRpcCallType1("eth_getTransactionCount", addr, tag)
	fmt.Println(res)
	if err != nil {
		return 0, err
	}
	trxC, err := res.GetString()
	if err != nil {
		return 0, nil
	}
	txCount, err := strconv.ParseInt(trxC[2:], 16, 64)
	if err != nil {
		return 0, nil
	}

	return txCount, nil
}

func (agent *ERC20Agent) BroadcastTransaction(rawtrx string) (string, error) {
	res, err := agent.DoHttpJsonRpcCallType1("eth_sendRawTransaction", rawtrx)
	if err != nil {
		return "", err
	}
	if res.Error != nil {
		return "", errors.New(res.Error.Message)
	}
	txid, err := res.GetString()
	if err != nil {
		return "", nil
	}
	return txid, err
}

func (agent *ERC20Agent) IsTransactionConfirmed(rawtrx string) (bool, error) {
	res, err := agent.DoHttpJsonRpcCallType1("eth_getTransactionByHash", rawtrx)
	if err != nil {
		return false, err
	}
	if res.Error != nil {
		return false, errors.New(res.Error.Message)
	}
	resmap, ok := res.Result.(map[string]interface{})
	if ok == false {
		return false, errors.New("parse response error")
	}
	cbn, ok := resmap["blockNumber"].(string)
	if err != nil {
		return false, err
	}
	var confirm_height int64
	if cbn == "" {
		confirm_height = 0
	} else {
		confirm_height, err = strconv.ParseInt(cbn[2:], 16, 32)
		if err != nil {
			return false, err
		}
	}

	if confirm_height > 0 {
		res, err := agent.DoHttpJsonRpcCallType1("eth_blockNumber")
		if err != nil {
			return false, err
		}
		if res.Error != nil {
			return false, errors.New(res.Error.Message)
		}
		curBlcokNumber, ok := res.Result.(string)
		if ok == false {
			return false, errors.New("parse response error")
		}
		height_number, err := strconv.ParseInt(curBlcokNumber[2:], 16, 32)
		if err != nil {
			return false, err
		}
		if height_number-confirm_height > int64(config.GlobalSupportCoinMgr["ETH"].ConfirmCount) {
			return true, nil
		}
	}
	return false, nil

}
