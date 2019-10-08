package cron

import (
	"context"
	"fmt"
	"github.com/owenliang/myf-go/client/mmongo"
	v3cron "github.com/robfig/cron/v3"
	"github.com/owenliang/myf-go/client/cat"
	"github.com/owenliang/myf-go/client/mhttp"
	"github.com/owenliang/myf-go/client/mmysql"
	"github.com/owenliang/myf-go/client/mredis"
	"github.com/owenliang/myf-go/conf"
	cronContext "github.com/owenliang/myf-go/cron/context"
	"github.com/owenliang/myf-go/cron/middlewares"
	"github.com/owenliang/myf-go/mcontext"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

// Cron框架实例
type MyfCron struct {
	cron *v3cron.Cron
	middlewares []cronContext.HandleFunc // 中间件
}

// 创建Cron
func New() (myfCron *MyfCron, err error) {
	// 加载配置
	if err = conf.InitConf(); err != nil {
		return
	}

	myfCron = &MyfCron{
		middlewares: make([]cronContext.HandleFunc, 0),
		cron: v3cron.New(),
	}

	if err = myfCron.initClients(); err != nil {
		return
	}

	// 注册中间件
	myfCron.registerMiddleware()
	return
}

// 启动APP
func (cron *MyfCron) Run() (err error) {
	cron.cron.Start()
	cron.waitGraceExit()
	return
}

// 注册中间件
func (cron *MyfCron) Use(handleFunc cronContext.HandleFunc) {
	cron.middlewares = append(cron.middlewares, handleFunc)
}

// 注册定时任务
func (cron *MyfCron) AddJob(name string, spec string, funcs ...cronContext.HandleFunc) (err error) {
	// 中间件数组
	middlewares := make([]cronContext.HandleFunc, len(cron.middlewares) + len(funcs))
	// 全局中间件
	copy(middlewares, cron.middlewares)
	// 任务级中间件
	copy(middlewares[len(cron.middlewares):], funcs)

	f := func() {
		// 每次任务执行前, 生成context
		ctx := cron.newContext(name, spec, middlewares)
		// 执行chain
		ctx.Next()
	}

	// 跳过尚未结束的任务
	job := v3cron.SkipIfStillRunning(v3cron.DefaultLogger)(newJob(f))
	_, err = cron.cron.AddJob(spec, job)
	return
}

// 生成任务执行的上下文
func (cron *MyfCron) newContext(name string, spec string, middlewares []cronContext.HandleFunc) (ctx *cronContext.Context){
	ctx = &cronContext.Context{
		Name: name,
		Spec: spec,
		Context: context.TODO(),
		Chain: middlewares,
		KeyValues: make(map[string]interface{}, 0),
	}
	return
}

// 闭包myf上下文
func (cron *MyfCron) WithMyfContext(myfHandle mcontext.MyfHandleFunc) cronContext.HandleFunc {
	return func(ctx *cronContext.Context) {
		// Myf上下文
		myfCtx := mcontext.MyfContext {
			Context: ctx,
			Cron: ctx,
			Cat: ctx.Value("cat").(*cat.MyfCat),
		}

		// 回调用户
		myfHandle(&myfCtx)
	}
}

// 等待优雅退出
func (cron *MyfCron) waitGraceExit() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			fmt.Fprintf(os.Stdout, "收到信号: %s, 服务正在退出... \n", s.String())
			select {
			case <- time.NewTimer(time.Duration(conf.MyfConf.CronConfig.WaitGraceExit) * time.Millisecond).C:	// 最多等待N秒
			case <- cron.cron.Stop().Done():	// 等待所有任务结束
			}
			return
		case syscall.SIGHUP:
		default:
		}
	}
}

// 注册中间件
func (cron *MyfCron) registerMiddleware() {
	// cat
	cron.Use(middlewares.Cat())
	// recovery
	cron.Use(middlewares.Recovery())
}

// 初始化各种客户端
func (cron *MyfCron) initClients() (err error) {
	// 拉起MYSQL
	if _, err = mmysql.NewMysql(conf.MyfConf.Mysql); err != nil {
		return
	}
	// 拉起REDIS
	if _, err = mredis.NewRedis(conf.MyfConf.Redis); err != nil {
		return
	}
	// 初始化http-client参数
	if err = mhttp.InitHttpClient(conf.MyfConf.HttpClientConfig); err != nil {
		return
	}
	// 拉起Mongo
	if _, err = mmongo.NewMongo(conf.MyfConf.Mongo); err != nil {
		return
	}
	return
}

// 初始化运行环境
func init() {
	runtime.GOMAXPROCS(runtime.NumCPU()) // 用满所有核心
}
