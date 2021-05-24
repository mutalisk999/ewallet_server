package model

import (
	"github.com/kataras/iris/core/errors"
	"sync"
	"time"
)

type tblServerCoinConfig struct {
	Coinid     int       `xorm:"pk INTEGER autoincr"`
	Coinsymbol string    `xorm:"VARCHAR(16) NOT NULL UNIQUE"`
	Ip         string    `xorm:"VARCHAR(64) NOT NULL"`
	Rpcport    int       `xorm:"INT NOT NULL"`
	Rpcuser    string    `xorm:"VARCHAR(64)"`
	Rpcpass    string    `xorm:"VARCHAR(64)"`
	State      int       `xorm:"INT NOT NULL"`
	Createtime time.Time `xorm:"DATETIME"`
	Updatetime time.Time `xorm:"DATETIME"`
}

type tblServerCoinConfigMgr struct {
	TableName string
	Mutex     *sync.Mutex
}

func (t *tblServerCoinConfigMgr) Init() {
	t.TableName = "tbl_server_coin_config"
	t.Mutex = new(sync.Mutex)
}

func (t *tblServerCoinConfigMgr) ListCoins() ([]tblServerCoinConfig, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var coins []tblServerCoinConfig
	err := GetDBEngine().Find(&coins)
	return coins, err

}

func (t *tblServerCoinConfigMgr) InsertCoin(coinSymbol string, ip string, port int, rpcUserName, rpcPassword string, state int) error {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var coin tblServerCoinConfig
	coin.Coinsymbol = coinSymbol
	coin.Ip = ip
	coin.Rpcport = port
	coin.Rpcuser = rpcUserName
	coin.Rpcpass = rpcPassword
	coin.State = state
	coin.Createtime = time.Now()
	coin.Updatetime = time.Now()
	_, err := GetDBEngine().Insert(&coin)
	return err
}

func (t *tblServerCoinConfigMgr) UpdateCoin(coinId int, coinSymbol string, ip string, port int, rpcUserName, rpcPassword string, state int) error {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var coin tblServerCoinConfig
	exist, err := GetDBEngine().Where("coinid=?", coinId).Get(&coin)
	if !exist {
		return errors.New("coinId not found!")
	}
	if err != nil {
		return err
	}
	coin.Coinsymbol = coinSymbol
	coin.Ip = ip
	coin.Rpcport = port
	coin.Rpcuser = rpcUserName
	coin.Rpcpass = rpcPassword
	coin.State = state
	coin.Updatetime = time.Now()
	_, err = GetDBEngine().Where("coinid=?", coinId).Update(&coin)
	return err

}

func (t *tblServerCoinConfigMgr) UpdateCoinState(coinId int, state int) error {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var coin tblServerCoinConfig
	exist, err := GetDBEngine().Where("coinid=?", coinId).Get(&coin)
	if !exist {
		return errors.New("coinId not found!")
	}
	if err != nil {
		return err
	}
	coin.State = state
	coin.Updatetime = time.Now()
	_, err = GetDBEngine().Where("coinid=?", coinId).Update(&coin)
	return err

}

func (t *tblServerCoinConfigMgr) GetCoin(coinId int) (tblServerCoinConfig, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var coin tblServerCoinConfig
	result, err := GetDBEngine().Where("coinid=?", coinId).Get(&coin)
	if !result {
		return coin, errors.New("key not found")
	}
	return coin, err
}

func (t *tblServerCoinConfigMgr) GetCoins(coinIds []int) ([]tblServerCoinConfig, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var coins []tblServerCoinConfig
	err := GetDBEngine().In("coinid", coinIds).Find(&coins)
	if err != nil {
		return nil, err
	}
	return coins, err
}

func (t *tblServerCoinConfigMgr) GetCoinBySymbol(symbol string) (tblServerCoinConfig, error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var coin tblServerCoinConfig
	result, err := GetDBEngine().Where("coinsymbol=?", symbol).Get(&coin)
	if !result {
		return coin, errors.New("symbol not found")
	}
	return coin, err

}
