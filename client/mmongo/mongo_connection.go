package mmongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/url"
	"strings"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type MongoConn struct {
	client *mongo.Client
	// 引用配置， 用于cat埋点
	groupConf    *MongoGroupConf
	connConf     []MongoConnConf
}

// 建立mongo连接
func newMongoConnection(groupConf *MongoGroupConf, connConf []MongoConnConf) (mongoConn *MongoConn, err error) {
	initOptions(groupConf, connConf)
	// 拼接参数
	hostList := make([]string, 0)
	for i := 0; i < len(connConf); i++ {
		hostList = append(hostList, fmt.Sprintf("%s:%d", connConf[i].Host, connConf[i].Port))
	}
	// mongodb://[username:password@]host1[:port1][,host2[:port2],...[,hostN[:portN]]][/[database][?options]]
	DSN := fmt.Sprintf("mongodb://%s:%s@%s/%s", groupConf.Username, url.PathEscape(groupConf.Password), strings.Join(hostList, ","), groupConf.Database)

	option := options.Client()
	option.SetMaxPoolSize(groupConf.MaxPoolSize)
	option.SetMinPoolSize(groupConf.MinPoolSize)
	option.SetConnectTimeout(time.Duration(groupConf.ConnectTimeout) * time.Millisecond)

	var client *mongo.Client
	client, err = mongo.NewClient(options.Client().ApplyURI(DSN))
	if err != nil {
		return
	}

	// 尝试连接
	err = client.Connect(context.TODO())
	if err != nil {
		return
	}

	// 确认连接正常建立
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return
	}

	mongoConn = &MongoConn{
		client: client,
		groupConf: groupConf,
	}
	return
}

// 初始化默认值
func initOptions(groupConf *MongoGroupConf, connConf []MongoConnConf)  {
	if groupConf.MaxPoolSize <= 0 {
		groupConf.MaxPoolSize = DEFAULT_MAX_POOL_SIZE
	}
	if groupConf.MinPoolSize <= 0 {
		groupConf.MinPoolSize = DEFAULT_MIN_POOL_SIZE
	}

	if groupConf.ConnectTimeout <= 0 {
		groupConf.ConnectTimeout = DEFAULT_CONNECT_TIMEOUT
	}

}

func (mongoConn *MongoConn) CurrentDB() (currentDB *mongo.Database, err error) {
	currentDB = mongoConn.client.Database(mongoConn.groupConf.Database)
	return
}
