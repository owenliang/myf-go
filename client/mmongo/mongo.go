package mmongo

import (
	"errors"
)

// 单例管理器
var MyfMongo *MongoManager

const (
	DEFAULT_MAX_POOL_SIZE uint64 = 100 //默认最大连接数
	DEFAULT_MIN_POOL_SIZE uint64 = 10 //默认闲置连接
	//ConnectTimeout
	DEFAULT_CONNECT_TIMEOUT int = 10000 //默认连接超时 10s
)

// 单实例配置
type MongoConnConf struct {
	Host string `toml:"host"` // host
	Port int    `toml:"port"` // 端口
}

// Mongo主从配置
type MongoGroupConf struct {
	Name     string `toml:"name"`
	Database string `toml:"database"` // 数据库名称
	Username string `toml:"username"` // 用户名
	Password string `toml:"password"` // 密码

	MaxPoolSize uint64 `toml:"maxPoolSize"` // 最大连接数
	MinPoolSize uint64 `toml:"minPoolSize"` // 最小连接数
	ConnectTimeout int `toml:"connectTimeout"` //连接超时

	Instances []MongoConnConf `toml:"Instance"` // 实例列表
}

type MongoConf struct {
	GroupConfList []MongoGroupConf `toml:"Group"`
}

// 错误码
var (
	ERR_MONGO_NAME_NOT_FOUND   = errors.New("mongo名称不能为空")
	ERR_MONGO_GROUP_NOT_FOUND  = errors.New("此mongo实例不存在")
	ERR_MONGO_CONN_NOT_FOUND   = errors.New("没有可用mongo实例连接")
)

