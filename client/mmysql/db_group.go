package mmysql

import (
	"context"
	"sync/atomic"
)

// Mysql主从组

type DBGroup struct {
	mCounter uint64
	Master   []*DBConn

	sCounter uint64
	Slave    []*DBConn
}

func newDBGroup(groupConf *DBGroupConf) (dbGroup *DBGroup, err error) {
	masterConf := groupConf.Master
	slaveConf := groupConf.Slaves

	dbGroup = &DBGroup{
		Master: make([]*DBConn, 0),
		Slave:  make([]*DBConn, 0),
	}

	var dbConn *DBConn

	// master连接池
	if masterConf != nil && masterConf.Instances != nil {
		for i := 0; i < len(masterConf.Instances); i++ {
			if dbConn, err = newDBConnection(groupConf, masterConf, &masterConf.Instances[i]); err != nil {
				return
			}
			dbGroup.Master = append(dbGroup.Master, dbConn)
		}
	}
	// slave连接池
	if slaveConf != nil && slaveConf.Instances != nil {
		for i := 0; i < len(slaveConf.Instances); i++ {
			if dbConn, err = newDBConnection(groupConf, slaveConf, &slaveConf.Instances[i]); err != nil {
				return
			}
			dbGroup.Slave = append(dbGroup.Slave, dbConn)
		}
	}
	return
}

// 选择连接
func (dbGroup *DBGroup) chooseConn(forceMaster bool) (conn *DBConn) {
	if !forceMaster && len(dbGroup.Slave) != 0 { // 选择slave
		counter := atomic.AddUint64(&dbGroup.sCounter, 1)
		return dbGroup.Slave[counter%uint64(len(dbGroup.Slave))]
	} else if len(dbGroup.Master) != 0 { // 选择master
		counter := atomic.AddUint64(&dbGroup.mCounter, 1)
		return dbGroup.Master[counter%uint64(len(dbGroup.Master))]
	}
	return
}

// SQL查询
func (dbGroup *DBGroup) Query(context context.Context, result interface{}, sql string, values ...interface{}) (err error) {
	dbConn := dbGroup.chooseConn(false)
	if dbConn == nil {
		err = ERR_DB_CONN_NOT_FOUND
		return
	}
	err = dbConn.Query(context, result, sql, values...)
	return
}

// SQL写入
func (dbGroup *DBGroup) Exec(context context.Context, sql string, values ...interface{}) (result int64, err error) {
	dbConn := dbGroup.chooseConn(true)
	if dbConn == nil {
		err = ERR_DB_CONN_NOT_FOUND
		return
	}
	result, err = dbConn.Exec(context, sql, values...)
	return
}

// 开启事务
func (dbGroup *DBGroup) Begin(context context.Context) (tx *DBConn, err error) {
	dbConn := dbGroup.chooseConn(true)
	if dbConn == nil {
		err = ERR_DB_CONN_NOT_FOUND
		return
	}
	tx, err = dbConn.Begin(context)
	return
}
