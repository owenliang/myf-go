package mmysql

// Mysql总管理器

type DBManager struct {
	GroupMap map[string]*DBGroup
}

func NewMysql(config *DBConf) (dbMgr *DBManager, err error) {
	dbMgr = &DBManager{
		GroupMap: make(map[string]*DBGroup),
	}
	MyfDB = dbMgr // 单例

	if config == nil || config.GroupConfList == nil {
		return
	}

	// 按Name索引每一个DBGroup
	groupConfList := config.GroupConfList
	for i := 0; i < len(groupConfList); i++ {
		var dbGroup *DBGroup
		if dbGroup, err = newDBGroup(&groupConfList[i]); err != nil {
			return
		}
		dbMgr.GroupMap[groupConfList[i].Name] = dbGroup
	}
	return
}

// 根据name获取mysql组
func (dbMgr *DBManager) Instance(name string) (dbGroup *DBGroup, err error) {
	var existed bool
	if dbGroup, existed = dbMgr.GroupMap[name]; !existed {
		err = ERR_DB_GROUP_NOT_FOUND
	}
	return
}
