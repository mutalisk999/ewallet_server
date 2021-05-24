package utils

import (
	"fmt"
	"github.com/kataras/iris"
	"github.com/kataras/iris/core/errors"
)

var GlobalError map[int]string

type Error struct {
	ErrCode int    `json:"code"`
	ErrMsg  string `json:"message"`
}

func SetInternalError(ctx iris.Context, errorStr string) {
	ctx.Values().Set("error", errorStr)
	ctx.StatusCode(iris.StatusInternalServerError)
}

func FormatSysError(err error) *Error {
	sysError := new(Error)
	if err == nil {
		sysError.ErrCode = 0
		sysError.ErrMsg = ""
	} else {
		sysError.ErrCode = 999999
		sysError.ErrMsg = err.Error()
	}
	return sysError
}

func GetErrorString(err *Error) string {
	if err == nil {
		return "no error"
	}
	return fmt.Sprintf("errcode: %d, error msg: %s", err.ErrCode, err.ErrMsg)
}

func InitGlobalError() {
	GlobalError = make(map[int]string)

	GlobalError[100000] = "invalid json rpc request"
	GlobalError[100001] = "invalid json rpc params"
	GlobalError[100002] = "connection not login!"
	GlobalError[200000] = "not support json rpc function [%s] for path [%s]"

	GlobalError[300001] = "db error, dbname [%s], optype [%s], detail [%s]"

	GlobalError[500000] = "needcount [%v] and keydeatil count [%v] isn't same"
	GlobalError[500001] = "walletname [%s] already exist!"

	GlobalError[600000] = "[%s] error hex string format"
	GlobalError[600001] = "verify login signedmessage error"

	GlobalError[900000] = "system error,details [%s]"
}

func InvalidCoinSymbol(coinSymbol string) error {
	return errors.New(fmt.Sprintf("Invaild Coin Symbol [%s]", coinSymbol))
}

func MakeError(errCode int, errArgs ...interface{}) *Error {
	err := new(Error)
	err.ErrCode = errCode
	format, ok := GlobalError[errCode]
	if ok {
		err.ErrMsg = fmt.Sprintf(format, errArgs...)
	}
	return err
}
