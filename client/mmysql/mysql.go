package mmysql

import "errors"

const (
	DEFAULT_MAX_CONN int = 10 //默认最大连接数
	DEFAULT_IDLE_CONN int = 5 //默认闲置连接数
)

// 单例管理器
var MyfDB *DBManager

// 单实例配置
type DBConnConf struct {
	Host string `toml:"host"` // host
	Port int    `toml:"port"` // 端口
}

// Master 或者 Slave配置
type DBSubGroupConf struct {
	MaxConn   int          `toml:"maxConn"`  // 最大连接数
	IdleConn  int          `toml:"idleConn"` // 最大保持连接数
	IdleTime  int          `toml:"idleTime"` // 空闲回收时间
	Instances []DBConnConf `toml:"Instance"` // 实例列表
}

// Mysql主从配置
type DBGroupConf struct {
	Name     string `toml:"name"`
	Database string `toml:"database"` // 数据库名称
	Username string `toml:"username"` // 用户名
	Password string `toml:"password"` // 密码

	Master *DBSubGroupConf `toml:"Master"` // 主库配置
	Slaves *DBSubGroupConf `toml:"Slave"`  // 从库配置
}

type DBConf struct {
	GroupConfList []DBGroupConf `toml:"Group"`
}

// 错误码
var (
	ERR_DB_GROUP_NOT_FOUND   = errors.New("此DB不存在")
	ERR_DB_CONN_NOT_FOUND    = errors.New("没有可用DB连接")
	ERR_QUERY_RESULT_INVALID = errors.New("result传参类型必须是*[]*ElemType")
	ERR_RECURSION_TX         = errors.New("嵌套开启了事务")
	ERR_INVALID_TX           = errors.New("非事务不能提交或回滚")
)
