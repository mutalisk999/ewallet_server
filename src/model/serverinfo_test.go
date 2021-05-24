package model

import (
	"testing"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
)

func TestTblServerInfoMgr_CreateServerInfo(t *testing.T) {
	InitDB("mysql","root:123456@tcp(192.168.1.107:3306)/ewallet_server?charset=utf8")
	AA,ERR :=GlobalDBMgr.ServerWalletConfigMgr.GetWalletByName("ABC")
	fmt.Println(AA,ERR)

}