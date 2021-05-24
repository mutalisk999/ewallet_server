package model

import (
	"sync"
	"time"
	"github.com/kataras/iris/core/errors"
)

type tblServerTask struct {
	Taskid      int       `xorm:"pk INTEGER autoincr"`
	Serverid     int       `xorm:"INT NOT NULL"`
	Type     int       `xorm:"INT NOT NULL"`
	Walletid     int       `xorm:"INT NOT NULL"`
	Trxid     int       `xorm:"INT NOT NULL"`
	State     int       `xorm:"INT NOT NULL"`
	Createtime time.Time `xorm:"created"`
	Updatetime   time.Time `xorm:"DATETIME"`
}

type tblServerTaskMgr struct {
	TableName string
	Mutex     *sync.Mutex
}

func (t *tblServerTaskMgr) Init() {
	t.TableName = "tbl_server_task"
	t.Mutex = new(sync.Mutex)
}

func (t *tblServerTaskMgr) NewServerTaskMgr(serverId int, pushType int, walletId int, trxId int, state int) (int, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	var task tblServerTask
	task.Serverid = serverId
	task.Type = pushType
	task.Walletid = walletId
	task.Trxid = trxId
	task.State = state
	task.Createtime = time.Now()
	_, err := GetDBEngine().Insert(&task)
	if err != nil {
		return 0, err
	}
	return task.Taskid, nil
}

func (t *tblServerTaskMgr) ListNoDealServerTasks() ([]tblServerTask, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	pushtasks := make([]tblServerTask, 0)
	err := GetDBEngine().Where("state = ?", 0).
		Desc("createtime").Find(&pushtasks)
	if err != nil {
		return nil, err
	}
	return pushtasks, nil
}

func (t *tblServerTaskMgr) ListDealedServerTasks() ([]tblServerTask, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	pushtasks := make([]tblServerTask, 0)
	err := GetDBEngine().Where("state = ?", 1).
		Desc("createtime").Find(&pushtasks)
	if err != nil {
		return nil, err
	}
	return pushtasks, nil
}

func (t *tblServerTaskMgr) GetTrxidByTaskid(taskid int) (int, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var task tblServerTask
	exist, err := GetDBEngine().Where("taskid=?", taskid).Get(&task)
	if err != nil {
		return taskid, err
	}
	if !exist {

		return taskid, errors.New("servertask Not Found")
	}

	return task.Taskid, err
}

func (t *tblServerTaskMgr) GetServerTasksCount(serverId int, state int) (int, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	var task tblServerTask
	total, err := GetDBEngine().Where("serverid = ?", serverId).And("state = ?", state).Count(&task)
	if err != nil {
		return 0, err
	}
	return int(total), nil
}

func (t *tblServerTaskMgr) UpdateServerTaskStateByServerId(serverId int,taskId int, state int) error {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	var task tblServerTask
	task.State = state
	task.Taskid = taskId
	task.State = state
	_, err := GetDBEngine().Where("serverid=? and taskid=?", serverId, taskId).Update(&task)
	if err != nil {
		return err
	}
	return nil
}

func (t *tblServerTaskMgr) DeleteDealedServerTask(serverId int,taskId int) error {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	var task tblServerTask
	task.Serverid = serverId
	task.Taskid = taskId
	_, err := GetDBEngine().Where("serverid=? and taskid=? and state=?", serverId, taskId,1).Delete(&task)
	if err != nil {
		return err
	}
	return nil
}

func (t *tblServerTaskMgr) GetServerTasks(serverId []int, pushType []int, walletId []int, trxId []int, state []int, offSet int, limit int) (int, []tblServerTask, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	dbSession := GetDBEngine().Where("")
	if serverId != nil && len(serverId) != 0 {
		dbSession = dbSession.In("serverid", serverId)
	}
	if pushType != nil && len(pushType) != 0 {
		dbSession = dbSession.In("type", pushType)
	}
	if walletId != nil && len(walletId) != 0 {
		dbSession = dbSession.In("walletid", walletId)
	}
	if trxId != nil && len(trxId) != 0 {
		dbSession = dbSession.In("trxid", trxId)
	}
	if state != nil && len(state) != 0 {
		dbSession = dbSession.In("state", state)
	}
	var task tblServerTask
	total, err := dbSession.Count(&task)
	if err != nil {
		return 0, nil, err
	}

	dbSession2 := GetDBEngine().Where("")
	if serverId != nil && len(serverId) != 0 {
		dbSession2 = dbSession2.In("serverid", serverId)
	}
	if pushType != nil && len(pushType) != 0 {
		dbSession2 = dbSession2.In("type", pushType)
	}
	if walletId != nil && len(walletId) != 0 {
		dbSession2 = dbSession2.In("walletid", walletId)
	}
	if trxId != nil && len(trxId) != 0 {
		dbSession2 = dbSession2.In("trxid", trxId)
	}
	if state != nil && len(state) != 0 {
		dbSession2 = dbSession2.In("state", state)
	}
	tasks := make([]tblServerTask, 0)
	dbSession2.Limit(limit, offSet).Desc("createtime").Find(&tasks)
	return int(total), tasks, nil
}