package conf

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/owenliang/myf-go/client/mhttp"
	"github.com/owenliang/myf-go/client/mmongo"
	"github.com/owenliang/myf-go/client/mmysql"
	"github.com/owenliang/myf-go/client/mredis"
)

var (
	confPath *string
	MyfConf  *MyfConfig
)

// 框架配置实例
type MyfConfig struct {
	Domain           string            `toml:"domain"`     // 域名
	Debug            int               `toml:"debug"`      // 调试模式
	AppConfig *AppConfig               `toml:"App"`        // App配置
	CronConfig *CronConfig             `toml:"Cron"`       // Cron配置
	CatConfig        *CatConfig        `toml:"Cat"`        // Cat配置
	Redis            *mredis.CACHEConf `toml:"Redis"`      //redis配置
	Mysql            *mmysql.DBConf    `toml:"Mysql"`      // mysql配置
	HttpClientConfig *mhttp.HttpConfig `toml:"HttpClient"` //http-client配置
	Mongo            *mmongo.MongoConf `toml:"Mongo"`      //mongo配置
}

// 默认配置
func defaultConf() *MyfConfig {
	return &MyfConfig{
		Debug:            0,
		Domain:           "",
		AppConfig: DefaultAppConfig(),
		CronConfig: DefaultCronConfig(),
		CatConfig:        nil,
		Redis:            nil,
		Mysql:            nil,
		HttpClientConfig: nil,
	}
}

// 初始化配置
func InitConf() (err error) {
	MyfConf = defaultConf()
	if _, err = toml.DecodeFile(*confPath, MyfConf); err != nil {
		return
	}
	return
}

func init() {
	confPath = flag.String("conf", "./app.toml", "框架配置文件路径")
}
