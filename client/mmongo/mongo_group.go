package mmongo

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoGroup struct {
	Instance   *MongoConn
}

// 新建连接
func newMongoGroup(groupConf *MongoGroupConf) (mongoGroup *MongoGroup, err error) {
	// mongo实例名字必须设置
	if len(groupConf.Name) <= 0 {
		err = ERR_MONGO_NAME_NOT_FOUND
		return
	}

	mongoGroup = &MongoGroup{}
	var mongoConn *MongoConn
	if mongoConn, err = newMongoConnection(groupConf, groupConf.Instances); err != nil {
		return
	}

	mongoGroup.Instance = mongoConn
	return
}

// 选择连接
func (mongoGroup *MongoGroup) CurrentDB() (currentDB *mongo.Database, err error) {
	currentDB, err =  mongoGroup.Instance.CurrentDB()
	return
}