package model

import (
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
)

type DBMgr struct {
	DBEngine              *xorm.Engine
	ServerInfoMgr           *tblServerInfoMgr
	ServerPubkeyPoolMgr         *tblServerPubkeyPoolMgr
	ServerCoinConfigMgr   *tblServerCoinConfigMgr
	ServerTransactionMgr         *tblServerTransactionMgr
	ServerWalletConfigMgr       *tblServerWalletConfigMgr
	ServerPendingTransactionMgr *tblServerPendingTransactionMgr
	ServerTaskMgr        *tblServerTaskMgr
	ServerOperationLogMgr       *tblServerOperationLogMgr
}

var GlobalDBMgr *DBMgr

func GetDBEngine() *xorm.Engine {
	return GlobalDBMgr.DBEngine
}

func InitDB(dbType string, dbSource string) error {
	var err error
	GlobalDBMgr = new(DBMgr)
	GlobalDBMgr.DBEngine, err = xorm.NewEngine(dbType, dbSource)
	if err != nil {
		return err
	}
	GlobalDBMgr.DBEngine.SetTableMapper(core.SnakeMapper{})
	GlobalDBMgr.DBEngine.SetColumnMapper(core.SnakeMapper{})

	GlobalDBMgr.ServerInfoMgr = new(tblServerInfoMgr)
	GlobalDBMgr.ServerInfoMgr.Init()

	GlobalDBMgr.ServerPubkeyPoolMgr = new(tblServerPubkeyPoolMgr)
	GlobalDBMgr.ServerPubkeyPoolMgr.Init()

	GlobalDBMgr.ServerCoinConfigMgr = new(tblServerCoinConfigMgr)
	GlobalDBMgr.ServerCoinConfigMgr.Init()

	GlobalDBMgr.ServerTransactionMgr = new(tblServerTransactionMgr)
	GlobalDBMgr.ServerTransactionMgr.Init()

	GlobalDBMgr.ServerWalletConfigMgr = new(tblServerWalletConfigMgr)
	GlobalDBMgr.ServerWalletConfigMgr.Init()

	GlobalDBMgr.ServerPendingTransactionMgr = new(tblServerPendingTransactionMgr)
	GlobalDBMgr.ServerPendingTransactionMgr.Init()

	GlobalDBMgr.ServerTaskMgr = new(tblServerTaskMgr)
	GlobalDBMgr.ServerTaskMgr.Init()

	GlobalDBMgr.ServerOperationLogMgr = new(tblServerOperationLogMgr)
	GlobalDBMgr.ServerOperationLogMgr.Init()


	return nil
}
