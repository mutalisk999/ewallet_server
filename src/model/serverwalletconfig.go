package model

import (
	"sync"
	"time"
	"github.com/kataras/iris/core/errors"
	"strconv"
)

type tblServerWalletConfig struct {
	Walletid     int       `xorm:"pk INTEGER autoincr"`
	Coinid		int    `xorm:"INT"`
	Walletname 	string   	 `xorm:"VARCHAR(64)"`
	Address    string       `xorm:"VARCHAR(64)"`
	Keys    string       `xorm:"TEXT"`
	Needsigcount    int       `xorm:"INT"`
	Keycount		int		  `xorm:"INT"`
	Fee		string		  `xorm:"VARCHAR(64)"`
	Gasprice    string       `xorm:"VARCHAR(64)"`
	Gaslimit    string       `xorm:"VARCHAR(64)"`
	Confirmcount    int       `xorm:"INT"`
	Confirmserverids    string       `xorm:"TEXT"`
	State    int       `xorm:"INT"`
	Createtime time.Time `xorm:"DATETIME"`
	Updatetime time.Time `xorm:"DATETIME"`
}



type tblServerWalletConfigMgr struct {
	TableName string
	Mutex     *sync.Mutex
}

func (t *tblServerWalletConfigMgr) Init() {
	t.TableName = "tbl_server_wallet_config"
	t.Mutex = new(sync.Mutex)
}

func (t *tblServerWalletConfigMgr) CreateWallet(coinid int,walletname string,address string,keys string,needsigcount int,keycount int,
	fee string,gasprice ,gaslimit string,confirmserverid int) (int,error){
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var info tblServerWalletConfig
	info.Address = address
	info.Coinid = coinid
	ok,err := GetDBEngine().Get(&info)
	if ok{
		return -1,errors.New("already exist same coinid and address!")
	}
	if err!=nil{
		return -1,err
	}
	info.Walletname = walletname
	info.Keys = keys
	info.Needsigcount = needsigcount
	info.Keycount = keycount
	info.Fee = fee
	info.Gasprice= gasprice
	info.Gaslimit = gaslimit
	info.Confirmcount = 1
	info.Confirmserverids = strconv.Itoa(confirmserverid)
	info.Createtime = time.Now()
	info.Updatetime = time.Now()
	info.State = 0
	if needsigcount ==1{
		info.State=1
	}
	count,err := GetDBEngine().Insert(&info)
	if count ==0{
		return -1,errors.New("insert wallet failed!")
	}
	if err!=nil{
		return -1 ,err
	}
	return info.Walletid, nil
}

func (t *tblServerWalletConfigMgr) GetWalletId(coinid int,address string)(int,error){
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var info tblServerWalletConfig
	ok,err:=GetDBEngine().Where("coinid=? and address=?",coinid,address).Get(&info)
	if err!=nil{
		return -1,err
	}
	if!ok{
		return -1,errors.New("not found wallet!")
	}
	return info.Walletid,nil
}


func (t *tblServerWalletConfigMgr) ConfirmWallet(walletid int,serverid int) (int,error){
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var info tblServerWalletConfig
	info.Walletid = walletid
	ok,err := GetDBEngine().Get(&info)
	if !ok{
		return -1,errors.New("wallet not exist!")
	}
	if err!=nil{
		return -1,err
	}
	if info.State ==1{
		return -1,errors.New("wallet already confirmed!")
	}
	if info.State ==2{
		return -1,errors.New("wallet already discard!")
	}

	info.Confirmcount = info.Confirmcount+1
	info.Confirmserverids = info.Confirmserverids+","+strconv.Itoa(serverid)
	info.Updatetime = time.Now()
	if info.Needsigcount ==info.Confirmcount{
		info.State=1
	}

	count,err := GetDBEngine().Where("walletid=?",walletid).Update(&info)
	if count ==0{
		return -1,errors.New("update wallet failed!")
	}
	if err!=nil{
		return -1 ,err
	}
	return info.Walletid, nil
}


func (t *tblServerWalletConfigMgr) GetRelateWallets(serverid int) ([]tblServerWalletConfig,error){
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var infos []tblServerWalletConfig
	err:=GetDBEngine().Where("keys like ?","%"+strconv.Itoa(serverid)+":%").Find(&infos)
	return infos,err
}


func (t *tblServerWalletConfigMgr) GetUnConfirmedRelateWallets(serverid int) ([]tblServerWalletConfig,error){
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var infos []tblServerWalletConfig
	err:=GetDBEngine().Where("state = 0 and keys like ?","%"+strconv.Itoa(serverid)+":%").Find(&infos)
	return infos,err
}


func (t *tblServerWalletConfigMgr) GetWalletByName(servername string) (*tblServerWalletConfig,error){
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var info tblServerWalletConfig
	exist,err:=GetDBEngine().Where("walletname=?",servername).Get(&info)
	if err!=nil{
		return nil,err
	}
	if exist{
		return &info,nil
	}
	return nil,nil
}



func (t *tblServerWalletConfigMgr) QueryWallet(coinsymbol []string,ids []int) ([]tblServerWalletConfig,error){
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var info []tblServerWalletConfig
	query_sql := GetDBEngine().Where("1=1")
	if len(coinsymbol)>0{
		query_sql = query_sql.In("coinsymbol",coinsymbol)
	}
	if len(ids)>0{
		query_sql = query_sql.In("walletid",ids)
	}

	err:=query_sql.Find(&info)
	if err!=nil{
		return make([]tblServerWalletConfig,0),err
	}

	return info,nil
}