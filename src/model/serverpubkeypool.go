package model

import (
	"sync"
	"time"
	"github.com/kataras/iris/core/errors"
	"strconv"
	"fmt"
)

type tblServerPubkeyPool struct {
	Serverid     int       `xorm:"INT"`
	Keyindex   int   `xorm:"INT"`
	Pubkey	string   	 `xorm:"VARCHAR(256) "`
	Isused    int       `xorm:"TINYINT"`
	Createtime time.Time `xorm:"DATETIME"`
	Updatetime time.Time `xorm:"DATETIME"`
}

type tblServerPubkeyPoolMgr struct {
	TableName string
	Mutex     *sync.Mutex
}

func (t *tblServerPubkeyPoolMgr) Init() {
	t.TableName = "tbl_server_pubkey_pool"
	t.Mutex = new(sync.Mutex)
}

func (t *tblServerPubkeyPoolMgr) InsertOnePubkey(serverid , keyindex int,pubkey string) (bool,error){
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var onekey tblServerPubkeyPool
	onekey.Serverid=serverid
	onekey.Pubkey = pubkey
	onekey.Keyindex = keyindex
	count,err := GetDBEngine().FindAndCount(&onekey)
	if count==0 || err!=nil{
		//insert
		onekey.Isused =0
		onekey.Createtime = time.Now()
		onekey.Updatetime = time.Now()
		count,err := GetDBEngine().Insert(&onekey)
		if err!=nil{
			return false,err
		}
		if count!=1 {
			return false,errors.New("insert failed!")
		}
		return true,nil

	}
	return true,errors.New("pubkey already exists")
}

func (t *tblServerPubkeyPoolMgr) InsertManyPubkey(datas []map[string]interface{}) (bool,error){
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	allkeys := make([]tblServerPubkeyPool,0,len(datas))
	for _,one_data := range datas{
		one_key := tblServerPubkeyPool{}
		one_key.Serverid= one_data["serverid"].(int)
		one_key.Pubkey = one_data["pubkey"].(string)
		one_key.Keyindex = one_data["keyindex"].(int)
		count,_ := GetDBEngine().FindAndCount(&one_key)
		if count>0 {
			return false,errors.New("already have this key in database "+one_data["pubkey"].(string))
		}
		one_key.Createtime = time.Now()
		one_key.Updatetime = time.Now()
		allkeys = append(allkeys, one_key)
	}
	count,err:=GetDBEngine().Insert(&allkeys)
	if err!=nil{
		return false,err
	}
	if int(count) != len(datas){
		return false,errors.New("insert count is wrong,insert count:"+strconv.Itoa(int(count))+" expect count:"+strconv.Itoa( len(datas)))
	}
	return true,nil
}



func (t *tblServerPubkeyPoolMgr) CheckFullPubkeys(datas []map[string]interface{}) (bool,error){
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	bak_keys := make(map[string]string)
	serverid := -1
	for _,one_data := range datas{
		one_key := tblServerPubkeyPool{}
		one_key.Serverid= one_data["serverid"].(int)
		if serverid == -1{
			serverid = one_key.Serverid
		}

		one_key.Serverid= one_data["serverid"].(int)
		if serverid != one_key.Serverid{
			return false,errors.New("different server id")
		}
		one_key.Pubkey = one_data["pubkey"].(string)
		one_key.Keyindex = one_data["keyindex"].(int)
		count,_ := GetDBEngine().FindAndCount(&one_key)
		bak_keys[strconv.Itoa(one_key.Serverid)+","+ strconv.Itoa(one_key.Keyindex)] = ""
		if count==0 {
			return false,errors.New("already have this key in database "+one_data["pubkey"].(string))
		}
	}
	var allkeys []tblServerPubkeyPool
	db_count,err:=GetDBEngine().Where("serverid=?",serverid).FindAndCount(&allkeys)
	if err!=nil{
		return false,err
	}
	if int(db_count) != len(bak_keys){
		return false,errors.New("pukey count is wrong,db count:"+strconv.Itoa(int(db_count))+" input count:"+strconv.Itoa( len(datas)))
	}
	return true,nil
}

func (t *tblServerPubkeyPoolMgr) QueryAndUseNewPubkey(serverid int) (*tblServerPubkeyPool,error){
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var pubkeypool tblServerPubkeyPool
	server_info,err:=GlobalDBMgr.ServerInfoMgr.GetServerInfo(serverid)
	if err!=nil{
		return nil,err
	}
	startindex := server_info.Serverstartindex
	//query unused key
	ok,err:=GetDBEngine().Where("serverid=? and keyindex>? and isused=0",server_info,startindex).Asc("keyindex").Limit(1,0).Get(&pubkeypool)
	if err!=nil{
		return nil,err
	}

	if !ok {
		return nil,errors.New("query failed")
	}
	fmt.Println(pubkeypool)
	pubkeypool.Isused=1
	pubkeypool.Updatetime=time.Now()
	count,err := GetDBEngine().Where("serverid=? and keyindex=?").Update(&pubkeypool)
	if count ==0 &&err!=nil{
		return nil,err
	}
	return &pubkeypool,nil
}


func (t *tblServerPubkeyPoolMgr) GetPubkeyByIdIndex(keyindex,serverid int) (*tblServerPubkeyPool,error){
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	var oneKey tblServerPubkeyPool
	exist,err:=GetDBEngine().Where("serverid=? and keyindex=?",serverid,keyindex).Get(&oneKey)
	if err!=nil{
		return nil,err
	}
	if !exist{
		return nil,errors.New("key not found!")
	}
	return &oneKey,nil
}
func (t *tblServerPubkeyPoolMgr) GetAllPubkey() ([]tblServerPubkeyPool,error){
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	var keys []tblServerPubkeyPool
	err:=GetDBEngine().Find(&keys)
	if err!=nil{
		return make([]tblServerPubkeyPool,0),err
	}

	return keys,nil
}