package mmongo

import (
	"go.mongodb.org/mongo-driver/mongo"
)

// mongo总管理器
type MongoManager struct {
	GroupMap map[string]*MongoGroup
}

// 生成Mongo实例
func NewMongo(config *MongoConf) (mongoMgr *MongoManager, err error) {
	mongoMgr = &MongoManager{
		GroupMap: make(map[string]*MongoGroup),
	}

	if config == nil || config.GroupConfList == nil {
		return
	}

	// 按Name索引每一个mongoGroup
	groupConfList := config.GroupConfList
	for i := 0; i < len(groupConfList); i++ {
		var mongoGroup *MongoGroup

		if mongoGroup, err = newMongoGroup(&groupConfList[i]); err != nil {
			return
		}
		mongoMgr.GroupMap[groupConfList[i].Name] = mongoGroup
	}

	MyfMongo = mongoMgr
	return
}

// 根据name获取mongo组
func (mongoMgr *MongoManager) Instance(name string) (client *mongo.Database, err error) {
	mongoGroup, existed := mongoMgr.GroupMap[name]
	if !existed {
		err = ERR_MONGO_GROUP_NOT_FOUND
		return
	}

	//调用group 选择具体是用那个connection
	client, err = mongoGroup.CurrentDB()
	return
}