package model

import (
"sync"
"time"
)

type tblServerOperatorLog struct {
	Logid      int       `xorm:"pk INTEGER autoincr"`
	Serverid     int       `xorm:"INT NOT NULL"`
	Optype     int       `xorm:"INT NOT NULL"`
	Content    string    `xorm:"TEXT NOT NULL"`
	Createtime time.Time `xorm:"created"`
}

type tblServerOperationLogMgr struct {
	TableName string
	Mutex     *sync.Mutex
}

func (t *tblServerOperationLogMgr) Init() {
	t.TableName = "tbl_server_operation_log"
	t.Mutex = new(sync.Mutex)
}

func (t *tblServerOperationLogMgr) GetServerOperatorLogs(serverId []int, opType []int, opTime [2]string, offSet int, limit int) (int, []tblServerOperatorLog, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	dbSession := GetDBEngine().Where("")
	if serverId != nil && len(serverId) != 0 {
		dbSession = dbSession.In("serverid", serverId)
	}
	if opType != nil && len(opType) != 0 {
		dbSession = dbSession.In("optype", opType)
	}
	if opTime[0] != "" {
		dbSession = dbSession.And("createtime > ?", opTime[0])
	}
	if opTime[1] != "" {
		dbSession = dbSession.And("createtime < ?", opTime[1])
	}
	var log tblServerOperatorLog
	total, err := dbSession.Count(&log)
	if err != nil {
		return 0, nil, err
	}

	dbSession2 := GetDBEngine().Where("")
	if serverId != nil && len(serverId) != 0 {
		dbSession2 = dbSession2.In("serverid", serverId)
	}
	if opType != nil && len(opType) != 0 {
		dbSession2 = dbSession2.In("optype", opType)
	}
	if opTime[0] != "" {
		dbSession2 = dbSession2.And("createtime > ?", opTime[0])
	}
	if opTime[1] != "" {
		dbSession2 = dbSession2.And("createtime < ?", opTime[1])
	}
	opLogs := make([]tblServerOperatorLog, 0)
	dbSession2.Limit(limit, offSet).Desc("createtime").Find(&opLogs)
	return int(total), opLogs, nil
}

func (t *tblServerOperationLogMgr) NewServerOperatorLog(serverId int, opType int, content string) (int, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	var log tblServerOperatorLog
	log.Serverid = serverId
	log.Optype = opType
	log.Content = content
	log.Createtime = time.Now()
	_, err := GetDBEngine().Insert(&log)
	if err != nil {
		return 0, err
	}
	return log.Logid, nil
}
