package model

import (
"errors"
"sync"
"time"
)

type tblServerTransaction struct {
	Trxid         int    `xorm:"pk INTEGER autoincr"`
	Rawtrxid      string `xorm:"VARCHAR(128)"`
	Walletid      int    `xorm:"INT NOT NULL"`
	Coinid        int    `xorm:"INT NOT NULL"`
	Contractaddr  string `xorm:"VARCHAR(128)"`
	Serverid      int    `xorm:"INT NOT NULL"`
	Fromaddr      string `xorm:"VARCHAR(128) NOT NULL"`
	Todetails      string  `xorm:"TEXT"`
	Fee        string `xorm:"VARCHAR(128) NOT NULL"`
	Trxtime       time.Time
	Needconfirm   int    `xorm:"INT NOT NULL"`
	Confirmed     int    `xorm:"INT NOT NULL"`
	Serverfirmed  string  `xorm:"TEXT"`
	Signedtrxs    string  `xorm:"TEXT"`
	State         int    `xorm:"INT NOT NULL"`
	Createtime    time.Time
	Updatetime   time.Time `xorm:"DATETIME"`
}

type tblServerTransactionMgr struct {
	TableName string
	Mutex     *sync.Mutex
}

func (t *tblServerTransactionMgr) Init() {
	t.TableName = "tbl_server_transaction"
	t.Mutex = new(sync.Mutex)
}

func (t *tblServerTransactionMgr) NewTransaction(walletId int, coinId int, contractAddr string, serverid int, from string, to string,
	fee string, needConfirm int, serverfirmed string, signedtrxs string) (int, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	var transaction tblServerTransaction
	transaction.Walletid = walletId
	transaction.Coinid = coinId
	transaction.Contractaddr = contractAddr
	transaction.Serverid = serverid
	transaction.Fromaddr = from
	transaction.Todetails = to
	transaction.Fee = fee
	transaction.Trxtime = time.Now()
	transaction.Needconfirm = needConfirm
	transaction.Confirmed = 0
	transaction.Serverfirmed = serverfirmed
	transaction.Signedtrxs = signedtrxs
	transaction.State = 0
	transaction.Createtime = time.Now()
	//transaction.Updatetime = time.Now()

	_, err := GetDBEngine().Insert(&transaction)
	if err != nil {
		return 0, err
	}
	return transaction.Trxid, nil
}

func (t *tblServerTransactionMgr) CreateNewTransaction(transaction *tblServerTransaction) (int, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	/*var transaction tblServerTransaction
	transaction.Walletid = walletId
	transaction.Coinid = coinId
	transaction.Contractaddr = contractAddr
	transaction.Serverid = serverid
	transaction.Fromaddr = from
	transaction.Todetails = to
	transaction.Fee = fee
	transaction.Trxtime = time.Now()
	transaction.Needconfirm = needConfirm
	transaction.Confirmed = 0
	transaction.Serverfirmed = serverfirmed
	transaction.Signedtrxs = signedtrxs
	transaction.State = 0
	transaction.Createtime = time.Now()
	//transaction.Updatetime = time.Now()*/

	_, err := GetDBEngine().Insert(&transaction)
	if err != nil {
		return 0, err
	}
	return transaction.Trxid, nil
}

func (t *tblServerTransactionMgr) GetTransactionById(trxId int) (tblServerTransaction, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	var trx tblServerTransaction
	result, err := GetDBEngine().Where("trxid=?", trxId).Get(&trx)
	if err != nil {
		return trx, err
	}
	if result {
		return trx, nil
	}
	return trx, errors.New("no find transaction")
}

func (t *tblServerTransactionMgr) UpdateTransaction(transaction tblServerTransaction) error {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	_, err := GetDBEngine().Id(transaction.Trxid).Update(transaction)
	if err != nil {
		return err
	}
	return nil
}

func (t *tblServerTransactionMgr) UpdateTransactionState(trxId int, state int) error {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	var transaction tblServerTransaction
	transaction.State = state
	_, err := GetDBEngine().Id(trxId).Cols("state").Update(&transaction)
	if err != nil {
		return err
	}
	return nil
}

func (t *tblServerTransactionMgr) ListTransactions() ([]tblServerTransaction, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	var trxs []tblServerTransaction
	err := GetDBEngine().Find(&trxs)
	return trxs, err
}

func (t *tblServerTransactionMgr) GetTransactionsByServerId(serverId int) ([]tblServerTransaction, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	var trxs []tblServerTransaction
	err := GetDBEngine().Where("serverid=?", serverId).Find(&trxs)
	return trxs, err
}

func (t *tblServerTransactionMgr) GetTransactionsByServerIdAndTaskid(serverId int,taskid string) ([]tblServerTransaction, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	var trxs []tblServerTransaction
	err := GetDBEngine().Where("serverid=? and taskid in(?)", serverId,taskid).Find(&trxs)
	return trxs, err
}

func (t *tblServerTransactionMgr) GetTransactionsByServerIdAndState(serverId int,state string) ([]tblServerTransaction, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	var trxs []tblServerTransaction
	err := GetDBEngine().Where("serverid=? and state in(?)", serverId,state).Find(&trxs)
	return trxs, err
}

func (t *tblServerTransactionMgr) GetTransactionsByServerIdAndTaskidAndState(serverId int,taskid []int,state []int) ([]tblServerTransaction, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	dbSession := GetDBEngine().Where("serverid=?",serverId)

	if taskid != nil && len(taskid) != 0 {
		dbSession = dbSession.In("taskid", taskid)
	}
	if state != nil && len(state) != 0 {
		dbSession = dbSession.In("state", state)
	}

	trxs := make([]tblServerTransaction, 0)
	dbSession.Find(&trxs)
	return trxs, nil
}

func (t *tblServerTransactionMgr) DeleteOnChainTrx(trxId int) error {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	var trx tblServerTransaction
	trx.Trxid = trxId
	_, err := GetDBEngine().Where("trxid=? and state=?", trxId, 2).Delete(&trx)
	if err != nil {
		return err
	}
	return nil
}

func (t *tblServerTransactionMgr) DeleteCaccelTrx(trxId int) error {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	var trx tblServerTransaction
	trx.Trxid = trxId
	_, err := GetDBEngine().Where("trxid=? and state=?", trxId, 3).Delete(&trx)
	if err != nil {
		return err
	}
	return nil
}
