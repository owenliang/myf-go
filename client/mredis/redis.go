package mredis

import (
	"errors"
	"time"
)

type ACC_TYPE string

const (
	MASTER ACC_TYPE = "w"
	SLAVE  ACC_TYPE = "r"
	DEFAULT_POOL_SIZE int = 10 //默认最大连接数
	DEFAULT_MIN_IDLE_CONN int = 5 //默认闲置连接数
)

// 单例管理器
var MyfRedis *CACHEManager

// 单实例配置
type CACHEConnConf struct {
	Host string `toml:"host"` // host+端口
	Port int    `toml:"port"` // 端口
}

// Master 或者 Slave配置
type CACHESubGroupConf struct {
	DialTimeout  time.Duration `toml: "dialTimeout"`  // 连接超时时间
	ReadTimeout  time.Duration `toml: "readTimeout"`  //  读-超时
	WriteTimeout time.Duration `toml: "writeTimeout"` // 写-超时

	PoolSize     int           `toml: "poolSize"`     //每个CPU上的默认最大连接总数, 默认是10
	MinIdleConns int           `toml: "minIdleConns"` //最小空闲连接数
	PoolTimeout  time.Duration `toml: "poolTimeout"`  //所有的连接打满后，请求超时时间 Default is ReadTimeout + 1 second
	IdleTimeout  time.Duration `toml: "idleTimeout"`  //闲置请求超时时间， Default is 5 minutes. -1 disables idle timeout check.

	Instances []CACHEConnConf `toml:"Instance"` // 实例列表
}

// Redis主从配置
type CACHEGroupConf struct {
	Name     string `toml:"name"`
	Password string `toml:"password"` // 密码

	Master *CACHESubGroupConf `toml:"Master"` // 主库配置
	Slaves *CACHESubGroupConf `toml:"Slave"`  // 从库配置

}

//多redis实例，配置
type CACHEConf struct {
	GroupConfList []CACHEGroupConf `toml:"Group"`
}

// 错误码
var (
	ERR_CACHE_NAME_NOT_FOUND   = errors.New("redis数据库名称不能为空")
	ERR_CACHE_GROUP_NOT_FOUND  = errors.New("此redis实例不存在")
	ERR_CACHE_CONN_NOT_FOUND   = errors.New("没有可用redis实例连接")
)
