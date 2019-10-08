package mredis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/owenliang/myf-go-cat/cat"
	cat2 "github.com/owenliang/myf-go/client/cat"
	"strings"
	"time"
)

// 单个cache连接

type CACHEConn struct {
	client *redis.Client
}

func newCACHEConnection(groupConf *CACHEGroupConf, subGroupConf *CACHESubGroupConf, connConf *CACHEConnConf) (cacheConn *CACHEConn, err error) {
	initOptions(groupConf, subGroupConf, connConf)

	DSN := fmt.Sprintf("%s:%d", connConf.Host, connConf.Port)

	client := redis.NewClient(&redis.Options{
		Addr:     DSN,
		Password: groupConf.Password, //传递密码

		DialTimeout:  subGroupConf.DialTimeout * time.Millisecond,  //闲置重新建立连接数时间 默认5s
		ReadTimeout:  subGroupConf.ReadTimeout * time.Millisecond,  //设置读超时时间,默认3s
		WriteTimeout: subGroupConf.WriteTimeout * time.Millisecond, //设置写超时时间，默认同slavetimeout

		PoolSize:     subGroupConf.PoolSize,
		MinIdleConns: subGroupConf.MinIdleConns,
		PoolTimeout:  subGroupConf.PoolTimeout * time.Millisecond,
		IdleTimeout:  subGroupConf.IdleTimeout * time.Millisecond,
	})

	cacheConn = &CACHEConn{
		client: client,
	}
	return
}

// 初始化默认值
func initOptions(groupConf *CACHEGroupConf, subGroupConf *CACHESubGroupConf, connConf *CACHEConnConf)  {
	if subGroupConf.PoolSize <= 0 {
		subGroupConf.PoolSize = DEFAULT_POOL_SIZE
	}
	if subGroupConf.MinIdleConns <= 0 {
		subGroupConf.MinIdleConns = DEFAULT_MIN_IDLE_CONN
	}

}

func (cacheConn *CACHEConn) NewCurrentClient(ctx context.Context) (currentClient *redis.Client, err error) {
	currentClient = cacheConn.client.WithContext(ctx)

	currentClient.WrapProcess(cacheConn.InjectCtx(currentClient))
	currentClient.WrapProcessPipeline(cacheConn.InjectPipelineCtx(currentClient))
	return
}

// 格式化redis命令用于cat展示
func formatCmds(cmders []redis.Cmder) string {
	var cmdList = make([]string, 0)

	// 对于每一个redis请求
	for _, cmd := range cmders {
		// 拼接命令参数
		var args = make([]string, 0)
		for _, arg := range cmd.Args() {
			args = append(args, fmt.Sprint(arg))
		}
		argsStr := strings.Join(args, " ")
		if err := cmd.Err(); err != nil {
			argsStr = fmt.Sprintf("%s(%s)", argsStr, err.Error())
		}
		cmdList = append(cmdList, argsStr)
	}
	return strings.Join(cmdList, "\n")
}

//将ctx注入到go redis上下文环境
func (cacheConn *CACHEConn) InjectCtx(client *redis.Client) func(old func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
	ctx := client.Context()
	return func(old func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
		return func(cmd redis.Cmder) (err error) {
			if zcat, ok := ctx.Value("cat").(*cat2.MyfCat); ok {
				m := cat.NewTransaction("Cache.Redis.Query", fmt.Sprintf("%s(%s)", cmd.Name(), client.Options().Addr))
				m.LogEvent("Cache.redis.server", client.Options().Addr)

				s := formatCmds([]redis.Cmder{cmd})
				m.AddData(s)

				zcat.Append(m)
				defer zcat.Pop()

				if err = old(cmd); err != nil {
					m.LogEvent("Error", client.Options().Addr+err.Error(), "E_notice")
					m.SetStatus(err.Error())
				}
			} else {
				err = old(cmd)
			}
			return err
		}
	}
}

//将ctx注入go redis的pipeline语句中
func (cacheConn *CACHEConn) InjectPipelineCtx(client *redis.Client) func(oldProcess func([]redis.Cmder) error) func([]redis.Cmder) error {
	ctx := client.Context()
	return func(oldProcess func([]redis.Cmder) error) func([]redis.Cmder) error {
		return func(cmders []redis.Cmder) (err error) {
			//cat埋点
			if zcat, ok := ctx.Value("cat").(*cat2.MyfCat); ok {
				m := cat.NewTransaction("Cache.Redis.Query", fmt.Sprintf("%s(%s)", "pipeline", client.Options().Addr))
				m.LogEvent("Cache.redis.server", client.Options().Addr)

				s := formatCmds(cmders)
				m.AddData(s)

				zcat.Append(m)
				defer zcat.Pop()

				if err = oldProcess(cmders); err != nil {
					m.LogEvent("Error", client.Options().Addr+err.Error(), "E_notice")
					m.SetStatus(err.Error())
				}
			} else {
				err = oldProcess(cmders)
			}
			return err
		}
	}
}
