package model

import (
	"sync"
	"time"
	"fmt"
	"github.com/kataras/iris/core/errors"
)


type tblServerInfo struct {
	Serverid     int       `xorm:"pk INTEGER autoincr"`
	Servername  string    `xorm:"VARCHAR(64) "`
	Serverstartindex	int   	 `xorm:"INT"`
	Serverstatus    int       `xorm:"INT"`
	Createtime time.Time `xorm:"DATETIME"`
	Updatetime time.Time `xorm:"DATETIME"`
}




type tblServerInfoMgr struct {
	TableName string
	Mutex     *sync.Mutex
}

func (t *tblServerInfoMgr) Init() {
	t.TableName = "tbl_server_info"
	t.Mutex = new(sync.Mutex)
}

func (t *tblServerInfoMgr) CreateServerInfo(servername string,startindex int) (int, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var info tblServerInfo
	info.Servername = servername
	info.Serverstartindex = startindex
	count,err := GetDBEngine().FindAndCount(&info)
	if count==0 || err!=nil{
		//insert
		info.Serverstatus =0
		info.Createtime = time.Now()
		info.Updatetime = time.Now()
		count,err := GetDBEngine().Insert(&info)
		if err!=nil{
			fmt.Println(err.Error())
			return -1,err
		}
		if count!=1 {
			fmt.Println("insert failed")
			return -1,nil
		}
		fmt.Println(info.Serverid)
		return info.Serverid,nil

	}
	return info.Serverid, errors.New("already register same server info!")
}



func (t *tblServerInfoMgr) GetServerInfo(serverid int) (*tblServerInfo, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var info tblServerInfo
	info.Serverid = serverid
	count,err := GetDBEngine().FindAndCount(&info)
	if count==0 || err!=nil{

		return nil,err

	}
	return &info, err
}

func (t *tblServerInfoMgr) GetServerInfoByServerID(serverId int) (tblServerInfo, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var acct tblServerInfo
	exist, err := GetDBEngine().Where("serverid=?", serverId).Get(&acct)
	if err != nil {
		return acct, err
	}
	if !exist {

		return acct, errors.New("serverinfo Not Found")
	}

	return acct, err
}

func (t *tblServerInfoMgr) GetServerInfoByServerName(servername string) (*tblServerInfo, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var info tblServerInfo
	info.Servername = servername
	count,err := GetDBEngine().FindAndCount(&info)
	if count==0 || err!=nil{

		return nil,err

	}
	return &info, err
}

func (t *tblServerInfoMgr) GetActiveServerInfos() ([]tblServerInfo, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var info []tblServerInfo
	count,err := GetDBEngine().Where("serverstatus=1").FindAndCount(&info)
	if count==0 || err!=nil{

		return make([]tblServerInfo,0),err

	}
	return info, err
}

func (t *tblServerInfoMgr) CheckGetActiveServerInfo(serverid int) (*tblServerInfo, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var info tblServerInfo
	info.Serverid = serverid
	count,err := GetDBEngine().Where("serverstatus=1").FindAndCount(&info)
	if count==0 || err!=nil{

		return nil,err

	}
	return &info, err
}

