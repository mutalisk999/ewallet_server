package coin

import (
	"bufio"
	"bytes"
	"config"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Messer4/bchaddr"
	"github.com/kataras/iris/core/errors"
	"github.com/mutalisk999/bitcoin-lib/src/blob"
	"github.com/mutalisk999/bitcoin-lib/src/keyid"
	"github.com/mutalisk999/bitcoin-lib/src/pubkey"
	"github.com/mutalisk999/bitcoin-lib/src/script"
	"github.com/mutalisk999/bitcoin-lib/src/serialize"
	"github.com/mutalisk999/bitcoin-lib/src/transaction"
	"github.com/mutalisk999/bitcoin-lib/src/utility"
	"github.com/ybbus/jsonrpc"
	"io"
	"sort"
	"strconv"
)

type BCHAgent struct {
	ServerUrl string
}

func (agent *BCHAgent) Type() string {
	return "BCHAgent"
}

func (agent *BCHAgent) CoinType() string {
	return "BCH"
}

func (agent *BCHAgent) Init(urlstr string) {
	agent.ServerUrl = urlstr
}

func (agent *BCHAgent) DoHttpJsonRpcCallType1(method string, args ...interface{}) (*jsonrpc.RPCResponse, error) {
	rpcClient := jsonrpc.NewClient(agent.ServerUrl)
	rpcResponse, err := rpcClient.Call(method, args)
	if err != nil {
		return nil, err
	}
	return rpcResponse, nil
}

func (agent *BCHAgent) GetBalanceByAddressRPC(addr string) (string, error) {
	supportCoin, ok := config.GlobalSupportCoinMgr[agent.CoinType()]
	if !ok {
		return "", errors.New("not support coin")
	}
	nPrec := supportCoin.Precision

	res, err := agent.DoHttpJsonRpcCallType1("listunspent", 0, 99999999, []string{addr})
	if err != nil {
		return "0", err
	}
	if res.Error != nil {
		return "0", errors.New(res.Error.Message)
	}

	sum := int64(0)
	for _, i := range res.Result.([]interface{}) {
		out := i.(map[string]interface{})

		amountva, ok := out["amount"]
		if ok == false {
			continue
		}
		amount, err := amountva.(json.Number).Float64()
		if err != nil {
			continue
		}
		amountStr := strconv.FormatFloat(amount, 'f', nPrec, 64)
		amountPrec, err := ToPrecisionAmount(amountStr, nPrec)
		if err != nil {
			continue
		}

		sum += amountPrec
	}
	return FromPrecisionAmount(sum, nPrec), nil
}

func (agent *BCHAgent) GetUtxosByAddressRPC(addr string) ([]UTXODetail, error) {
	supportCoin, ok := config.GlobalSupportCoinMgr[agent.CoinType()]
	if !ok {
		return nil, errors.New("not support coin")
	}
	nPrec := supportCoin.Precision

	res, err := agent.DoHttpJsonRpcCallType1("listunspent", 0, 99999999, []string{addr})
	if err != nil {
		return nil, err
	}
	if res.Error != nil {
		return nil, errors.New(res.Error.Message)
	}

	var utxos UTXOsDetail
	for _, i := range res.Result.([]interface{}) {
		var utxo UTXODetail
		utxo.Address = addr
		out := i.(map[string]interface{})

		amount, ok := out["amount"]
		if ok == false {
			continue
		}
		txid, ok := out["txid"]
		if ok == false {
			continue
		}
		vout, ok := out["vout"]
		if ok == false {
			continue
		}
		scriptPubKey, ok := out["scriptPubKey"]
		if ok == false {
			continue
		}
		confirmations, ok := out["confirmations"]
		if ok == false {
			continue
		}

		amountValue, err := amount.(json.Number).Float64()
		if err != nil {
			continue
		}
		amountStr := strconv.FormatFloat(amountValue, 'f', nPrec, 64)
		amountPrec, err := ToPrecisionAmount(amountStr, nPrec)
		if err != nil {
			continue
		}

		if amountPrec == 0 {
			continue
		}
		utxo.Amount = amountPrec

		txidValue := txid.(string)
		utxo.TxId = txidValue

		i64, err := vout.(json.Number).Int64()
		if err != nil {
			continue
		}
		utxo.Vout = int(i64)

		scriptPubKeyValue := scriptPubKey.(string)
		utxo.ScriptPubKey = scriptPubKeyValue

		i64, err = confirmations.(json.Number).Int64()
		if err != nil {
			continue
		}
		utxo.Confirmations = int(i64)

		utxos = append(utxos, utxo)
	}

	// sort by confirmations desc
	sort.Sort(utxos)

	return utxos, nil
}

func (agent *BCHAgent) ImportAddressRPC(address string) error {

	res, err := agent.DoHttpJsonRpcCallType1("importaddress", address, "", false)
	if err != nil {
		return err
	}
	if res.Error != nil {
		return errors.New(res.Error.Message)
	}
	return nil
}

func (agent *BCHAgent) BroadcastTransactionRPC(rawtrx string) (string, error) {
	res, err := agent.DoHttpJsonRpcCallType1("sendrawtransaction", rawtrx)
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

func (agent *BCHAgent) IsTransactionConfirmedRPC(trxId string) (bool, error) {
	res, err := agent.DoHttpJsonRpcCallType1("gettransaction", trxId)
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
	cfm, err := resmap["confirmations"].(json.Number).Int64()
	if err != nil {
		return false, err
	}

	coin, ok := config.GlobalSupportCoinMgr[agent.CoinType()]
	if !ok {
		return false, errors.New("not support coin")
	}
	if cfm >= int64(coin.ConfirmCount) {
		return true, nil
	}
	return false, nil

}

func (agent *BCHAgent) IsAddressValidRPC(address string) (bool, error) {
	res, err := agent.DoHttpJsonRpcCallType1("validateaddress", address)
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
	isValid := resmap["isvalid"].(bool)
	if err != nil {
		return false, err
	}
	return isValid, nil

}

func BCHGetUnCompressPubKey(pubKeyBytes []byte) ([]byte, error) {
	if len(pubKeyBytes) != 64 {
		return nil, errors.New("invalid pubKeyBytes size")
	}

	pubkeyUnCompress := make([]byte, 65, 65)
	pubkeyUnCompress[0] = 0x4
	copy(pubkeyUnCompress[1:], pubKeyBytes[0:64])

	return pubkeyUnCompress, nil
}

func BCHGetCompressPubKey(pubKeyBytes []byte) ([]byte, error) {
	if len(pubKeyBytes) != 64 {
		return nil, errors.New("invalid pubKeyBytes size")
	}

	pubkeyCompress := make([]byte, 33, 33)
	if pubKeyBytes[63]%2 == 0 {
		pubkeyCompress[0] = 0x2
	} else {
		pubkeyCompress[0] = 0x3
	}
	copy(pubkeyCompress[1:], pubKeyBytes[0:32])

	fmt.Println("pubkeyCompress:", hex.EncodeToString(pubkeyCompress))

	return pubkeyCompress, nil
}

func BCHCalcAddressByPubKey(pubKeyStr string) (string, error) {
	pubKeyBytes, err := hex.DecodeString(pubKeyStr)
	if err != nil {
		return "", err
	}

	pubkeyCompress, err := BCHGetCompressPubKey(pubKeyBytes)
	if err != nil {
		return "", err
	}

	pubKey := new(pubkey.PubKey)
	pubKey.SetPubKeyData(pubkeyCompress)

	keyIdBytes, err := pubKey.CalcKeyIDBytes()
	if err != nil {
		return "", err
	}
	keyId := new(keyid.KeyID)
	keyId.SetKeyIDData(keyIdBytes)

	var version byte
	if config.IsTestEnvironment {
		version = 111
	} else {
		version = 0
	}
	addrStr, err := keyId.ToBase58Address(version)
	if err != nil {
		return "", err
	}
	addrStr, err = bchaddr.ToCashAddress(addrStr, false)
	if err != nil {
		return "", err
	}
	return addrStr, nil
}

func BCHGetRedeemScriptByPubKeys(needCount int, pubKeyStrList []string) (string, error) {
	if needCount <= 0 || needCount > 16 {
		return "", errors.New("BCHGetRedeemScriptByPubKeys error: invalid needCount")
	}
	if len(pubKeyStrList) == 0 || len(pubKeyStrList) > 16 {
		return "", errors.New("BCHGetRedeemScriptByPubKeys error: invalid pubKeyStrList size")
	}
	if needCount > len(pubKeyStrList) {
		return "", errors.New("BCHGetRedeemScriptByPubKeys error: needCount greater than pubKeyStrList size")
	}

	bytesBuf := bytes.NewBuffer([]byte{})
	bufWriter := io.Writer(bytesBuf)
	err := serialize.PackUint8(bufWriter, uint8(needCount+0x50))
	if err != nil {
		return "", err
	}
	for _, pubKeyStr := range pubKeyStrList {
		pubKeyBytes, err := hex.DecodeString(pubKeyStr)
		if err != nil {
			return "", err
		}

		pubKeyCpsBytes, err := BCHGetCompressPubKey(pubKeyBytes)
		if err != nil {
			return "", err
		}

		pubKey := new(pubkey.PubKey)
		pubKey.SetPubKeyData(pubKeyCpsBytes)

		err = pubKey.Pack(bufWriter)
		if err != nil {
			return "", err
		}
	}
	err = serialize.PackUint8(bufWriter, uint8(len(pubKeyStrList)+0x50))
	if err != nil {
		return "", err
	}
	err = serialize.PackUint8(bufWriter, uint8(script.OP_CHECKMULTISIG))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytesBuf.Bytes()), nil
}

func BCHGetMultiSignAddressByRedeemScript(redeemScriptStr string) (string, error) {
	redeemScript, err := hex.DecodeString(redeemScriptStr)
	if err != nil {
		return "", err
	}

	scriptIdBytes := utility.Hash160(redeemScript)
	keyId := new(keyid.KeyID)
	keyId.SetKeyIDData(scriptIdBytes)

	var version byte
	if config.IsTestEnvironment {
		version = 196
	} else {
		version = 5
	}

	addrStr, err := keyId.ToBase58Address(version)
	if err != nil {
		return "", err
	}
	addrStr, err = bchaddr.ToCashAddress(addrStr, false)
	if err != nil {
		return "", err
	}
	return addrStr, nil
}

func BCHGetMultiSignScriptByRedeemScript(redeemScriptStr string) ([]byte, error) {
	redeemScript, err := hex.DecodeString(redeemScriptStr)
	if err != nil {
		return nil, err
	}
	scriptIdBytes := utility.Hash160(redeemScript)

	bytesBuf := bytes.NewBuffer([]byte{})
	bufWriter := io.Writer(bytesBuf)
	err = serialize.PackUint8(bufWriter, script.OP_HASH160)
	if err != nil {
		return nil, err
	}
	multiSigScript := new(script.Script)
	multiSigScript.SetScriptBytes(scriptIdBytes)
	err = multiSigScript.Pack(bufWriter)
	if err != nil {
		return nil, err
	}
	err = serialize.PackUint8(bufWriter, script.OP_EQUAL)
	if err != nil {
		return nil, err
	}

	return bytesBuf.Bytes(), nil
}

func (agent *BCHAgent) BuildTrxInPutsOutPutsRPC(addrFromStr string, addrToStr string, amountTransferStr string, feeRateStr string) (string, InPuts, OutPuts, error) {
	supportCoin, ok := config.GlobalSupportCoinMgr[agent.CoinType()]
	if !ok {
		return "0", nil, nil, errors.New("not support coin")
	}
	nPrec := supportCoin.Precision

	balanceStr, err := agent.GetBalanceByAddressRPC(addrFromStr)
	if err != nil {
		return "0", nil, nil, err
	}
	balance, err := ToPrecisionAmount(balanceStr, nPrec)
	if err != nil {
		return "0", nil, nil, err
	}
	amountTransfer, err := ToPrecisionAmount(amountTransferStr, nPrec)
	if err != nil {
		return "0", nil, nil, err
	}
	feeRate, err := ToPrecisionAmount(feeRateStr, nPrec)
	if err != nil {
		return "0", nil, nil, err
	}

	if balance <= amountTransfer {
		return "0", nil, nil, errors.New("not enough balance")
	}

	utxos, err := agent.GetUtxosByAddressRPC(addrFromStr)
	if err != nil {
		return "0", nil, nil, err
	}

	spentBalance := int64(0)
	change := int64(0)
	feeCost := int64(0)
	trxBytes := 0
	balanceOk := false
	inputs := make([]InPut, 0)
	outputs := make(map[string]string)

	// trx size 100k
	for _, utxo := range utxos {
		// ignore dust
		if utxo.Amount <= 546 {
			continue
		}
		inputs = append(inputs, InPut{TxId: utxo.TxId, Vout: utxo.Vout, ScriptPubKey: utxo.ScriptPubKey})
		spentBalance = spentBalance + utxo.Amount
		trxBytes = len(inputs)*180 + 40 + 40
		if trxBytes > 100*1000 {
			return "0", nil, nil, errors.New("too large trx size")
		}
		feeCost = int64(float64(trxBytes) / 1000.0 * float64(feeRate))
		if spentBalance >= amountTransfer+feeCost {
			balanceOk = true
			break
		}
	}
	if balanceOk != true {
		return "0", nil, nil, errors.New("not enough balance")
	}

	if addrToStr == addrFromStr {
		outputs[addrToStr] = fmt.Sprintf(FromPrecisionAmount(spentBalance-feeCost, nPrec))
	} else {
		outputs[addrToStr] = fmt.Sprintf(FromPrecisionAmount(amountTransfer, nPrec))
		change = spentBalance - amountTransfer - feeCost
	}

	if change > 546 {
		outputs[addrFromStr] = fmt.Sprintf(FromPrecisionAmount(change, nPrec))
	}

	return FromPrecisionAmount(feeCost, nPrec), inputs, outputs, nil
}

func (agent *BCHAgent) BuildTrxFromUtxosRPC(utxos []UTXODetail, addrFromStr string, addrToStr string, amountTransferStr string, feeRateStr string) (string, InPuts, OutPuts, error) {
	supportCoin, ok := config.GlobalSupportCoinMgr[agent.CoinType()]
	if !ok {
		return "0", nil, nil, errors.New("not support coin")
	}
	nPrec := supportCoin.Precision

	balanceStr := CalcBalanceFromUtxos(utxos, addrFromStr, nPrec)

	balance, err := ToPrecisionAmount(balanceStr, nPrec)
	if err != nil {
		return "0", nil, nil, err
	}
	amountTransfer, err := ToPrecisionAmount(amountTransferStr, nPrec)
	if err != nil {
		return "0", nil, nil, err
	}
	feeRate, err := ToPrecisionAmount(feeRateStr, nPrec)
	if err != nil {
		return "0", nil, nil, err
	}

	if balance <= amountTransfer {
		return "0", nil, nil, errors.New("not enough balance")
	}

	spentBalance := int64(0)
	change := int64(0)
	feeCost := int64(0)
	trxBytes := 0
	balanceOk := false
	inputs := make([]InPut, 0)
	outputs := make(map[string]string)

	// trx size 100k
	for _, utxo := range utxos {
		if utxo.Address != addrFromStr {
			continue
		}
		// ignore dust
		if utxo.Amount <= 546 {
			continue
		}
		inputs = append(inputs, InPut{TxId: utxo.TxId, Vout: utxo.Vout, ScriptPubKey: utxo.ScriptPubKey})
		spentBalance = spentBalance + utxo.Amount
		trxBytes = len(inputs)*180 + 40 + 40
		if trxBytes > 100*1000 {
			return "0", nil, nil, errors.New("too large trx size")
		}
		feeCost = int64(float64(trxBytes) / 1000.0 * float64(feeRate))
		if spentBalance >= amountTransfer+feeCost {
			balanceOk = true
			break
		}
	}
	if balanceOk != true {
		return "0", nil, nil, errors.New("not enough balance")
	}

	if addrToStr == addrFromStr {
		outputs[addrToStr] = fmt.Sprintf(FromPrecisionAmount(spentBalance-feeCost, nPrec))
	} else {
		outputs[addrToStr] = fmt.Sprintf(FromPrecisionAmount(amountTransfer, nPrec))
		change = spentBalance - amountTransfer - feeCost
	}

	if change > 546 {
		outputs[addrFromStr] = fmt.Sprintf(FromPrecisionAmount(change, nPrec))
	}

	return FromPrecisionAmount(feeCost, nPrec), inputs, outputs, nil
}

func (agent *BCHAgent) CreateRawTransactionRPC(inputs InPuts, outputs OutPuts) (string, error) {
	res, err := agent.DoHttpJsonRpcCallType1("createrawtransaction", inputs, outputs)
	if err != nil {
		return "", err
	}
	if res.Error != nil {
		return "", errors.New(res.Error.Message)
	}
	return res.Result.(string), nil
}

func BCHUnPackRawTransaction(rawTrx string) (*transaction.Transaction, error) {
	Blob := new(blob.Byteblob)
	err := Blob.SetHex(rawTrx)
	if err != nil {
		return nil, err
	}
	bytesBuf := bytes.NewBuffer(Blob.GetData())
	bufReader := io.Reader(bytesBuf)
	trx := new(transaction.Transaction)
	err = trx.UnPack(bufReader)
	if err != nil {
		return nil, err
	}
	return trx, nil
}

func BCHPackRawTransaction(trxSig transaction.Transaction) (string, error) {
	bytesBuf := bytes.NewBuffer([]byte{})
	bufWriter := io.Writer(bytesBuf)
	err := trxSig.Pack(bufWriter)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytesBuf.Bytes()), nil
}

func BCHCombineSignatureAndPubKey(signature []byte, pubKey []byte) []byte {
	scriptSig := make([]byte, 0, 1+len(signature)+1+len(pubKey))
	scriptSig = append(scriptSig, byte(len(signature)))
	scriptSig = append(scriptSig, signature...)
	scriptSig = append(scriptSig, byte(len(pubKey)))
	scriptSig = append(scriptSig, pubKey...)
	fmt.Println("scriptSig ", hex.EncodeToString(scriptSig))
	return scriptSig
}

func BCHCombineSignatureAndRedeemScript(signature []byte, redeemScriptBytes []byte) ([]byte, error) {
	bytesBuf := bytes.NewBuffer([]byte{})
	bufWriter := io.Writer(bytesBuf)
	err := serialize.PackUint8(bufWriter, script.OP_0)
	if err != nil {
		return nil, err
	}
	signatureScript := new(script.Script)
	signatureScript.SetScriptBytes(signature)
	err = signatureScript.Pack(bufWriter)
	if err != nil {
		return nil, err
	}

	if len(redeemScriptBytes) < int(script.OP_PUSHDATA1) {
	} else {
		opPushData := uint8(0)
		if len(redeemScriptBytes) <= 0xff {
			opPushData = script.OP_PUSHDATA1
		} else if len(redeemScriptBytes) <= 0xffff {
			opPushData = script.OP_PUSHDATA2
		} else {
			opPushData = script.OP_PUSHDATA4
		}
		err = serialize.PackUint8(bufWriter, opPushData)
		if err != nil {
			return nil, err
		}
	}

	redeemScrip := new(script.Script)
	redeemScrip.SetScriptBytes(redeemScriptBytes)
	err = redeemScrip.Pack(bufWriter)
	if err != nil {
		return nil, err
	}
	return bytesBuf.Bytes(), nil
}

func (agent *BCHAgent) SignRawTransactionRPC(rawTrx string, pubKeyStr string, keyIndex uint16, utxos []UTXODetail) (string, error) {
	pubKeyBytes, err := hex.DecodeString(pubKeyStr)
	if err != nil {
		return "", err
	}
	pubkeyCompress, err := BCHGetCompressPubKey(pubKeyBytes)
	if err != nil {
		return "", err
	}

	trx, err := BCHUnPackRawTransaction(rawTrx)
	if err != nil {
		return "", err
	}

	signedDataList := make([][]byte, len(trx.Vin))

	// add scriptPubKey
	for i := 0; i < len(trx.Vin); i++ {
		vinFound := false
		var scriptCode script.Script
		var amount int64
		for j := 0; j < len(utxos); j++ {
			if trx.Vin[i].PrevOut.Hash.GetHex() == utxos[j].TxId {
				vinFound = true
				scriptPubKey, err := hex.DecodeString(utxos[j].ScriptPubKey)
				if err != nil {
					return "", errors.New("invalid ScriptPubKey")
				}
				scriptCode.SetScriptBytes(scriptPubKey)
				amount = utxos[j].Amount
				fmt.Println("scriptPubKey len", len(scriptPubKey))
				fmt.Println("amount", amount)
				break
			}
		}
		if vinFound != true {
			return "", errors.New("can not found valid utxo for rawTrx")
		}

		bytesBuf := bytes.NewBuffer([]byte{})
		bufWriter := io.Writer(bytesBuf)

		// version
		err := serialize.PackInt32(bufWriter, trx.Version)
		if err != nil {
			return "", err
		}

		// hashPrevout
		hashBlob, err := BCHSigGetPrevoutHash(*trx)
		if err != nil {
			return "", err
		}
		err = hashBlob.Pack(bufWriter, hashBlob.GetDataSize())
		if err != nil {
			return "", err
		}

		// hashSequence
		hashBlob, err = BCHSigGetSequenceHash(*trx)
		if err != nil {
			return "", err
		}
		err = hashBlob.Pack(bufWriter, hashBlob.GetDataSize())
		if err != nil {
			return "", err
		}

		// prevout
		err = trx.Vin[i].PrevOut.Pack(bufWriter)
		if err != nil {
			return "", err
		}

		// script
		err = scriptCode.Pack(bufWriter)
		if err != nil {
			return "", err
		}

		// amount
		err = serialize.PackInt64(bufWriter, amount)
		if err != nil {
			return "", err
		}

		// sequence
		err = serialize.PackUint32(bufWriter, trx.Vin[i].Sequence)
		if err != nil {
			return "", err
		}

		// hashOutPuts
		hashBlob, err = BCHSigGetOutputsHash(*trx)
		if err != nil {
			return "", err
		}
		err = hashBlob.Pack(bufWriter, hashBlob.GetDataSize())
		if err != nil {
			return "", err
		}

		// lockTime
		err = serialize.PackUint32(bufWriter, trx.LockTime)
		if err != nil {
			return "", err
		}

		// hashType
		nHashType := uint32(0x1 | 0x40)
		err = serialize.PackUint32(bufWriter, nHashType)
		if err != nil {
			return "", err
		}

		rawTrxBytes := bytesBuf.Bytes()

		hashBytes := utility.Sha256(utility.Sha256(rawTrxBytes))

		fmt.Println("rawTrxBytes:", hex.EncodeToString(rawTrxBytes))
		fmt.Println("hashBytes:", hex.EncodeToString(hashBytes))

		// signature
		var rsBytes []byte

		for {
			rsBytes, err = CoinSignTrx('1', hashBytes, keyIndex)
			if err != nil {
				return "", err
			}
			if len(rsBytes) != 64 {
				return "", errors.New("invalid r/s lens")
			}
			if rsBytes[32] < 128 {
				break
			}
		}

		verifyOk, err := CoinVerifyTrx('1', keyIndex, hashBytes, rsBytes)
		if err != nil {
			return "", err
		}
		if !verifyOk {
			return "", errors.New("verify signature error")
		}
		fmt.Println("rsHex:", hex.EncodeToString(rsBytes))

		rBytes := rsBytes[0:32]
		sBytes := rsBytes[32:64]

		// serialize r,s to der encoding
		signedData, err := SerializeDerEncoding(rBytes, sBytes)
		if err != nil {
			return "", err
		}
		fmt.Println("signedData:", hex.EncodeToString(signedData))

		// append SIGHASH_ALL
		signedData = append(signedData, 0x1|0x40)

		scriptSig := BCHCombineSignatureAndPubKey(signedData, pubkeyCompress)

		signedDataList[i] = scriptSig
	}

	for i := 0; i < len(trx.Vin); i++ {
		trx.Vin[i].ScriptSig.SetScriptBytes(signedDataList[i])
	}

	trxSigStr, err := BCHPackRawTransaction(*trx)
	if err != nil {
		return "", err
	}

	return trxSigStr, nil
	return "", nil
}

func BCHSigGetPrevoutHash(trx transaction.Transaction) (blob.Baseblob, error) {
	bytesBuf := bytes.NewBuffer([]byte{})
	bufWriter := io.Writer(bytesBuf)
	for i := 0; i < len(trx.Vin); i++ {
		err := trx.Vin[i].PrevOut.Pack(bufWriter)
		if err != nil {
			return blob.Baseblob{}, err
		}
	}
	var hashBlob blob.Baseblob
	hashBlob.SetData(utility.Sha256(utility.Sha256(bytesBuf.Bytes())))
	return hashBlob, nil
}

func BCHSigGetSequenceHash(trx transaction.Transaction) (blob.Baseblob, error) {
	bytesBuf := bytes.NewBuffer([]byte{})
	bufWriter := io.Writer(bytesBuf)
	for i := 0; i < len(trx.Vin); i++ {
		err := serialize.PackUint32(bufWriter, trx.Vin[i].Sequence)
		if err != nil {
			return blob.Baseblob{}, err
		}
	}
	var hashBlob blob.Baseblob
	hashBlob.SetData(utility.Sha256(utility.Sha256(bytesBuf.Bytes())))
	return hashBlob, nil
}

func BCHSigGetOutputsHash(trx transaction.Transaction) (blob.Baseblob, error) {
	bytesBuf := bytes.NewBuffer([]byte{})
	bufWriter := io.Writer(bytesBuf)
	for i := 0; i < len(trx.Vout); i++ {
		err := trx.Vout[i].Pack(bufWriter)
		if err != nil {
			return blob.Baseblob{}, err
		}
	}
	var hashBlob blob.Baseblob
	hashBlob.SetData(utility.Sha256(utility.Sha256(bytesBuf.Bytes())))
	return hashBlob, nil
}

func (agent *BCHAgent) MultiSignRawTransactionRPC(rawTrx string, redeemScriptStr string, keyIndex uint16, utxos []UTXODetail) (string, error) {
	var redeemScript script.Script

	redeemScriptBytes, err := hex.DecodeString(redeemScriptStr)
	if err != nil {
		return "", errors.New("invalid redeemScript")
	}
	redeemScript.SetScriptBytes(redeemScriptBytes)

	trx, err := BCHUnPackRawTransaction(rawTrx)
	if err != nil {
		return "", err
	}
	signedDataList := make([][]byte, len(trx.Vin))

	// add scriptPubKey
	for i := 0; i < len(trx.Vin); i++ {
		vinFound := false
		var amount int64
		for j := 0; j < len(utxos); j++ {
			if trx.Vin[i].PrevOut.Hash.GetHex() == utxos[j].TxId {
				vinFound = true
				amount = utxos[j].Amount
				fmt.Println("amount", amount)
				break
			}
		}
		if vinFound != true {
			return "", errors.New("can not found valid utxo for rawTrx")
		}

		bytesBuf := bytes.NewBuffer([]byte{})
		bufWriter := io.Writer(bytesBuf)

		// version
		err := serialize.PackInt32(bufWriter, trx.Version)
		if err != nil {
			return "", err
		}

		// hashPrevout
		hashBlob, err := BCHSigGetPrevoutHash(*trx)
		if err != nil {
			return "", err
		}
		err = hashBlob.Pack(bufWriter, hashBlob.GetDataSize())
		if err != nil {
			return "", err
		}

		// hashSequence
		hashBlob, err = BCHSigGetSequenceHash(*trx)
		if err != nil {
			return "", err
		}
		err = hashBlob.Pack(bufWriter, hashBlob.GetDataSize())
		if err != nil {
			return "", err
		}

		// prevout
		err = trx.Vin[i].PrevOut.Pack(bufWriter)
		if err != nil {
			return "", err
		}

		// script
		err = redeemScript.Pack(bufWriter)
		if err != nil {
			return "", err
		}

		// amount
		err = serialize.PackInt64(bufWriter, amount)
		if err != nil {
			return "", err
		}

		// sequence
		err = serialize.PackUint32(bufWriter, trx.Vin[i].Sequence)
		if err != nil {
			return "", err
		}

		// hashOutPuts
		hashBlob, err = BCHSigGetOutputsHash(*trx)
		if err != nil {
			return "", err
		}
		err = hashBlob.Pack(bufWriter, hashBlob.GetDataSize())
		if err != nil {
			return "", err
		}

		// lockTime
		err = serialize.PackUint32(bufWriter, trx.LockTime)
		if err != nil {
			return "", err
		}

		// hashType
		nHashType := uint32(0x1 | 0x40)
		err = serialize.PackUint32(bufWriter, nHashType)
		if err != nil {
			return "", err
		}

		rawTrxBytes := bytesBuf.Bytes()
		hashBytes := utility.Sha256(utility.Sha256(rawTrxBytes))

		fmt.Println("rawTrxBytes:", hex.EncodeToString(rawTrxBytes))
		fmt.Println("hashBytes:", hex.EncodeToString(hashBytes))

		// signature
		var rsBytes []byte

		for {
			rsBytes, err = CoinSignTrx('1', hashBytes, keyIndex)
			if err != nil {
				return "", err
			}
			if len(rsBytes) != 64 {
				return "", errors.New("invalid r/s lens")
			}
			if rsBytes[32] < 128 {
				break
			}
		}

		fmt.Println("hashBytes", hex.EncodeToString(hashBytes))
		fmt.Println("rsBytes", hex.EncodeToString(rsBytes))

		verifyOk, err := CoinVerifyTrx('1', keyIndex, hashBytes, rsBytes)
		if err != nil {
			return "", err
		}
		if !verifyOk {
			return "", errors.New("verify signature error")
		}
		fmt.Println("rsHex:", hex.EncodeToString(rsBytes))

		rBytes := rsBytes[0:32]
		sBytes := rsBytes[32:64]

		// serialize r,s to der encoding
		signedData, err := SerializeDerEncoding(rBytes, sBytes)
		if err != nil {
			return "", err
		}
		fmt.Println("signedData:", hex.EncodeToString(signedData))

		// append SIGHASH_ALL
		signedData = append(signedData, 0x1|0x40)

		scriptSig, err := BCHCombineSignatureAndRedeemScript(signedData, redeemScriptBytes)
		if err != nil {
			return "", err
		}
		signedDataList[i] = scriptSig
	}

	for i := 0; i < len(trx.Vin); i++ {
		trx.Vin[i].ScriptSig.SetScriptBytes(signedDataList[i])
	}

	trxSigStr, err := BCHPackRawTransaction(*trx)
	if err != nil {
		return "", err
	}

	return trxSigStr, nil
}

func (agent *BCHAgent) MultiVerifySignRawTransactionRPC(signedRawTrx string, pubKeyStr string, utxos []UTXODetail) error {
	trx, err := BCHUnPackRawTransaction(signedRawTrx)
	if err != nil {
		return err
	}

	signedDataList := make([]script.Script, 0, len(trx.Vin))
	for i := 0; i < len(trx.Vin); i++ {
		signedDataList = append(signedDataList, trx.Vin[i].ScriptSig)
	}

	for i := 0; i < len(trx.Vin); i++ {
		trxTemp, err := BCHUnPackRawTransaction(signedRawTrx)
		if err != nil {
			return err
		}
		// set empty script
		for i := 0; i < len(trx.Vin); i++ {
			emptyScript := new(script.Script)
			trxTemp.Vin[i].ScriptSig.SetScriptBytes(emptyScript.GetScriptBytes())
		}

		bytesBuf := bytes.NewBuffer(signedDataList[i].GetScriptBytes())
		bufReader := io.Reader(bytesBuf)
		u8, err := serialize.UnPackUint8(bufReader)
		if err != nil {
			return err
		}
		if u8 != 0 {
			return errors.New("invalid multisig script, not started with 0x0")
		}

		signatureScript := new(script.Script)
		err = signatureScript.UnPack(bufReader)
		if err != nil {
			return err
		}

		signatureScriptBytes := signatureScript.GetScriptBytes()
		if signatureScriptBytes[len(signatureScriptBytes)-1] != 0x1|0x40 {
			return errors.New("invalid signature, not ended with 0x1|0x40[ALL|FORKID]")
		}
		signatureScriptBytes = signatureScriptBytes[:len(signatureScriptBytes)-1]

		// skip oppushdata
		tmpBufReader := bufio.NewReader(bufReader)
		opPushDataBytes, err := tmpBufReader.Peek(1)
		if err != nil {
			return err
		}

		if opPushDataBytes[0] == script.OP_PUSHDATA1 || opPushDataBytes[0] == script.OP_PUSHDATA2 || opPushDataBytes[0] == script.OP_PUSHDATA4 {
			bufReader = io.Reader(tmpBufReader)
			_, err = serialize.UnPackUint8(bufReader)
			if err != nil {
				return err
			}
		}

		redeemScript := new(script.Script)
		err = redeemScript.UnPack(bufReader)
		if err != nil {
			return err
		}

		// get compress address
		pubKeyBytes, err := hex.DecodeString(pubKeyStr)
		if err != nil {
			return err
		}
		compressPubKey, err := BCHGetCompressPubKey(pubKeyBytes)
		if err != nil {
			return err
		}
		solverOk, whichType, PubKeys := script.Solver(*redeemScript)
		if solverOk != true || whichType != script.TX_MULTISIG || len(PubKeys) == 0 {
			return errors.New("invalid redeemscript, solver error")
		}

		// is compressPubKey in PubKeys
		isInPubKeys := false
		for _, pubKey := range PubKeys {
			if 0 == bytes.Compare(compressPubKey, pubKey) {
				isInPubKeys = true
				break
			}
		}
		if !isInPubKeys {
			return errors.New("arguments pubKeyStr of function MultiVerifySignRawTransaction is no match to redeemscript")
		}

		// prepare signature data
		vinFound := false
		var amount int64
		for j := 0; j < len(utxos); j++ {
			if trx.Vin[i].PrevOut.Hash.GetHex() == utxos[j].TxId {
				vinFound = true
				amount = utxos[j].Amount
				fmt.Println("amount", amount)
				break
			}
		}
		if vinFound != true {
			return errors.New("can not found valid utxo for rawTrx")
		}

		bytesBuf = bytes.NewBuffer([]byte{})
		bufWriter := io.Writer(bytesBuf)

		// version
		err = serialize.PackInt32(bufWriter, trx.Version)
		if err != nil {
			return err
		}

		// hashPrevout
		hashBlob, err := BCHSigGetPrevoutHash(*trx)
		if err != nil {
			return err
		}
		err = hashBlob.Pack(bufWriter, hashBlob.GetDataSize())
		if err != nil {
			return err
		}

		// hashSequence
		hashBlob, err = BCHSigGetSequenceHash(*trx)
		if err != nil {
			return err
		}
		err = hashBlob.Pack(bufWriter, hashBlob.GetDataSize())
		if err != nil {
			return err
		}

		// prevout
		err = trx.Vin[i].PrevOut.Pack(bufWriter)
		if err != nil {
			return err
		}

		// script
		err = redeemScript.Pack(bufWriter)
		if err != nil {
			return err
		}

		// amount
		err = serialize.PackInt64(bufWriter, amount)
		if err != nil {
			return err
		}

		// sequence
		err = serialize.PackUint32(bufWriter, trx.Vin[i].Sequence)
		if err != nil {
			return err
		}

		// hashOutPuts
		hashBlob, err = BCHSigGetOutputsHash(*trx)
		if err != nil {
			return err
		}
		err = hashBlob.Pack(bufWriter, hashBlob.GetDataSize())
		if err != nil {
			return err
		}

		// lockTime
		err = serialize.PackUint32(bufWriter, trx.LockTime)
		if err != nil {
			return err
		}

		// hashType
		nHashType := uint32(0x1 | 0x40)
		err = serialize.PackUint32(bufWriter, nHashType)
		if err != nil {
			return err
		}

		rawTrxBytes := bytesBuf.Bytes()
		hashBytes := utility.Sha256(utility.Sha256(rawTrxBytes))

		fmt.Println("rawTrxBytes:", hex.EncodeToString(rawTrxBytes))
		fmt.Println("hashBytes:", hex.EncodeToString(hashBytes))
		fmt.Println("signatureScriptBytes", hex.EncodeToString(signatureScriptBytes))

		// do not verify by crypto device
		//verifyOk, err := CoinVerifyTrxWithOutsidePubkey('2', pubKeyBytes, hashBytes, signatureScriptBytes)

		verifyOk, err := CoinVerifyTrx2(compressPubKey, hashBytes, signatureScriptBytes)
		if err != nil {
			return err
		}
		if !verifyOk {
			return errors.New("verify signature error")
		}
	}

	return nil
}

func (agent *BCHAgent) CombineRawTransactionRPC(signedRawTrxs []string) (string, error) {
	trxs := make([]*transaction.Transaction, 0)
	for _, signedRawTrx := range signedRawTrxs {
		trx, err := OMNIUnPackRawTransaction(signedRawTrx)
		if err != nil {
			return "", err
		}
		trxs = append(trxs, trx)
	}

	vinCount := 0
	if len(trxs) == 0 {
		return "", errors.New("transaction count is 0")
	} else if len(trxs) == 1 {
		return signedRawTrxs[0], nil
	} else {
		vinCount = len(trxs[0].Vin)
		for i := 1; i < len(trxs); i++ {
			if vinCount != len(trxs[i].Vin) {
				return "", errors.New("signedRawTrxs with different vins")
			}
		}
	}

	tempTrx, err := OMNIUnPackRawTransaction(signedRawTrxs[0])
	if err != nil {
		return "", err
	}

	validRedeemScript := false
	redeemScript := new(script.Script)
	for i := 0; i < vinCount; i++ {
		signatureScripts := make([]*script.Script, 0)
		for j := 0; j < len(trxs); j++ {
			bytesBuf := bytes.NewBuffer(trxs[j].Vin[i].ScriptSig.GetScriptBytes())
			bufReader := io.Reader(bytesBuf)
			u8, err := serialize.UnPackUint8(bufReader)
			if err != nil {
				return "", err
			}
			if u8 != 0 {
				return "", errors.New("invalid multisig script, not started with 0x0")
			}

			signatureScript := new(script.Script)
			err = signatureScript.UnPack(bufReader)
			if err != nil {
				return "", err
			}

			signatureScriptBytes := signatureScript.GetScriptBytes()
			if signatureScriptBytes[len(signatureScriptBytes)-1] != 0x1|0x40 {
				return "", errors.New("invalid signature, not ended with 0x1|0x40[ALL|FORKID]")
			}
			signatureScriptBytes = signatureScriptBytes[:len(signatureScriptBytes)-1]

			// skip oppushdata
			tmpBufReader := bufio.NewReader(bufReader)
			opPushDataBytes, err := tmpBufReader.Peek(1)
			if err != nil {
				return "", err
			}

			if opPushDataBytes[0] == script.OP_PUSHDATA1 || opPushDataBytes[0] == script.OP_PUSHDATA2 || opPushDataBytes[0] == script.OP_PUSHDATA4 {
				bufReader = io.Reader(tmpBufReader)
				_, err = serialize.UnPackUint8(bufReader)
				if err != nil {
					return "", err
				}
			}

			if !validRedeemScript {
				err = redeemScript.UnPack(bufReader)
				if err != nil {
					return "", err
				}
				validRedeemScript = true
			}

			signatureScripts = append(signatureScripts, signatureScript)
		}

		// combine signature
		bytesBuf := bytes.NewBuffer([]byte{})
		bufWriter := io.Writer(bytesBuf)
		err := serialize.PackUint8(bufWriter, 0)
		if err != nil {
			return "", err
		}

		for _, signatureScript := range signatureScripts {
			err = signatureScript.Pack(bufWriter)
			if err != nil {
				return "", err
			}
		}

		// oppushdata
		opPushData := uint8(0)
		if redeemScript.GetScriptLength() < int(script.OP_PUSHDATA1) {
			opPushData = uint8(0)
		} else if redeemScript.GetScriptLength() >= int(script.OP_PUSHDATA1) && redeemScript.GetScriptLength() <= 0xff {
			opPushData = script.OP_PUSHDATA1
		} else if redeemScript.GetScriptLength() > 0xff && redeemScript.GetScriptLength() <= 0xffff {
			opPushData = script.OP_PUSHDATA2
		} else if redeemScript.GetScriptLength() > 0xffff {
			opPushData = script.OP_PUSHDATA4
		}

		if opPushData != uint8(0) {
			err := serialize.PackUint8(bufWriter, opPushData)
			if err != nil {
				return "", err
			}
		}

		// redeemscript
		err = redeemScript.Pack(bufWriter)
		if err != nil {
			return "", err
		}

		tempTrx.Vin[i].ScriptSig.SetScriptBytes(bytesBuf.Bytes())
	}

	bytesBuf := bytes.NewBuffer([]byte{})
	bufWriter := io.Writer(bytesBuf)
	err = tempTrx.Pack(bufWriter)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytesBuf.Bytes()), nil
}
