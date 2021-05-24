package model

import (
	"sync"
	"time"
)

type tblServerPendingTransaction struct {
	Id      int       `xorm:"pk INTEGER autoincr"`
	Trxid     int       `xorm:"INT NOT NULL"`
	Coinid     int       `xorm:"INT NOT NULL"`
	Vintrxid    string `xorm:"VARCHAR(128)"`
	Vinvout     int       `xorm:"INT NOT NULL"`
	Fromaddress string `xorm:"VARCHAR(128)"`
	Balance     int       `xorm:"INT NOT NULL"`
	Createtime time.Time `xorm:"created"`
	Updatetime   time.Time `xorm:"DATETIME"`
}

type tblServerPendingTransactionMgr struct {
	TableName string
	Mutex     *sync.Mutex
}

func (t *tblServerPendingTransactionMgr) Init() {
	t.TableName = "tbl_server_pending_transaction"
	t.Mutex = new(sync.Mutex)
}

func (t *tblServerPendingTransactionMgr) NewServerPendingTransactionMgr(trxid int, coinid int, vintrxid string, vinvout int, fromaddr string, balance int) (int, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	var pendingtrx tblServerPendingTransaction
	pendingtrx.Trxid = trxid
	pendingtrx.Coinid = coinid
	pendingtrx.Vintrxid = vintrxid
	pendingtrx.Vinvout = vinvout
	pendingtrx.Fromaddress = fromaddr
	pendingtrx.Balance = balance
	pendingtrx.Createtime = time.Now()
	_, err := GetDBEngine().Insert(&pendingtrx)
	if err != nil {
		return 0, err
	}
	return pendingtrx.Id, nil
}

func (t *tblServerPendingTransactionMgr) GeServerPendingTransactions(trxId []int, coinId []int, vinTrxid []string, vinVout []int,
	fromAddres []string, Balance []int, offSet int, limit int) (int, []tblServerPendingTransaction, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	dbSession := GetDBEngine().Where("")
	if trxId != nil && len(trxId) != 0 {
		dbSession = dbSession.In("trxid", trxId)
	}
	if coinId != nil && len(coinId) != 0 {
		dbSession = dbSession.In("coinid", coinId)
	}
	if vinTrxid != nil && len(vinTrxid) != 0 {
		dbSession = dbSession.In("vintrxid", vinTrxid)
	}
	if vinVout != nil && len(vinVout) != 0 {
		dbSession = dbSession.In("vinvout", vinVout)
	}
	if fromAddres != nil && len(fromAddres) != 0 {
		dbSession = dbSession.In("fromaddress", fromAddres)
	}
	if Balance != nil && len(Balance) != 0 {
		dbSession = dbSession.In("balance", Balance)
	}
	var pendingtrx tblServerPendingTransaction
	total, err := dbSession.Count(&pendingtrx)
	if err != nil {
		return 0, nil, err
	}

	dbSession2 := GetDBEngine().Where("")
	if trxId != nil && len(trxId) != 0 {
		dbSession2 = dbSession2.In("trxid", trxId)
	}
	if coinId != nil && len(coinId) != 0 {
		dbSession2 = dbSession2.In("coinid", coinId)
	}
	if vinTrxid != nil && len(vinTrxid) != 0 {
		dbSession2 = dbSession2.In("vintrxid", vinTrxid)
	}
	if vinVout != nil && len(vinVout) != 0 {
		dbSession2 = dbSession2.In("vinvout", vinVout)
	}
	if fromAddres != nil && len(fromAddres) != 0 {
		dbSession2 = dbSession2.In("fromaddress", fromAddres)
	}
	if Balance != nil && len(Balance) != 0 {
		dbSession2 = dbSession2.In("balance", Balance)
	}
	pendingtrxs := make([]tblServerPendingTransaction, 0)
	dbSession2.Limit(limit, offSet).Desc("createtime").Find(&pendingtrxs)
	return int(total), pendingtrxs, nil
}

func (t *tblServerPendingTransactionMgr) DeleteServerPendingTrx(trxId int) error {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var pendtrx tblServerPendingTransaction
	pendtrx.Trxid = trxId
	_, err := GetDBEngine().Where("trxid=?", trxId).Delete(pendtrx)
	return err
}