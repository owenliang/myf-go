package mredis

import (
	"context"
	"github.com/go-redis/redis"
	"sync/atomic"
)

// Mysql主从组

type CACHEGroup struct {
	mCounter uint64
	Master   []*CACHEConn

	sCounter uint64
	Slave    []*CACHEConn
}

func newCACHEGroup(groupConf *CACHEGroupConf) (cacheGroup *CACHEGroup, err error) {
	//redis实例名字必须设置
	if len(groupConf.Name) <= 0 {
		err = ERR_CACHE_NAME_NOT_FOUND
		return
	}

	masterConf := groupConf.Master
	slaveConf := groupConf.Slaves

	cacheGroup = &CACHEGroup{
		Master: make([]*CACHEConn, 0),
		Slave:  make([]*CACHEConn, 0),
	}

	var cacheConn *CACHEConn

	// master连接池
	if masterConf != nil && masterConf.Instances != nil {
		for i := 0; i < len(masterConf.Instances); i++ {
			if cacheConn, err = newCACHEConnection(groupConf, masterConf, &masterConf.Instances[i]); err != nil {
				return
			}
			cacheGroup.Master = append(cacheGroup.Master, cacheConn)
		}
	}

	// slave连接池
	if slaveConf != nil && slaveConf.Instances != nil {
		for i := 0; i < len(slaveConf.Instances); i++ {
			if cacheConn, err = newCACHEConnection(groupConf, slaveConf, &slaveConf.Instances[i]); err != nil {
				return
			}
			cacheGroup.Slave = append(cacheGroup.Slave, cacheConn)
		}
	}
	return
}

// 选择连接
func (cacheGroup *CACHEGroup) ChooseConn(ctx context.Context, accessType ACC_TYPE) (currentClient *redis.Client, err error) {
	var currentConn *CACHEConn

	// 从库
	if accessType == SLAVE && len(cacheGroup.Slave) > 0 {
		counter := atomic.AddUint64(&cacheGroup.sCounter, 1)
		currentConn = cacheGroup.Slave[counter%uint64(len(cacheGroup.Slave))]
	} else if len(cacheGroup.Master) != 0 {
		counter := atomic.AddUint64(&cacheGroup.mCounter, 1)
		currentConn = cacheGroup.Master[counter%uint64(len(cacheGroup.Master))]
	}

	if currentConn == nil {
		err = ERR_CACHE_CONN_NOT_FOUND
		return
	}

	// 包装请求上下文
	currentClient, err = currentConn.NewCurrentClient(ctx)
	return
}
