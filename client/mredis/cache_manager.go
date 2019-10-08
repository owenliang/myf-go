package mredis

import (
	"context"
	"github.com/go-redis/redis"
)

// CACHE总管理器
type CACHEManager struct {
	GroupMap map[string]*CACHEGroup
}

func NewRedis(config *CACHEConf) (cacheMgr *CACHEManager, err error) {
	cacheMgr = &CACHEManager{
		GroupMap: make(map[string]*CACHEGroup),
	}

	if config == nil || config.GroupConfList == nil {
		return
	}

	// 按Name索引每一个CACHEGroup
	groupConfList := config.GroupConfList
	for i := 0; i < len(groupConfList); i++ {
		var cacheGroup *CACHEGroup
		if cacheGroup, err = newCACHEGroup(&groupConfList[i]); err != nil {
			return
		}
		cacheMgr.GroupMap[groupConfList[i].Name] = cacheGroup
	}

	MyfRedis = cacheMgr
	return
}

// 根据name获取mysql组
func (cacheMgr *CACHEManager) Instance(ctx context.Context, name string, accessType ACC_TYPE) (client *redis.Client, err error) {
	cacheGroup, existed := cacheMgr.GroupMap[name]
	if !existed {
		err = ERR_CACHE_GROUP_NOT_FOUND
		return
	}

	//调用group 选择具体是用那个connection
	client, err = cacheGroup.ChooseConn(ctx, accessType)
	return
}
