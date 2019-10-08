package app

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/owenliang/myf-go/app/middlewares"
	"github.com/owenliang/myf-go/client/cat"
	"github.com/owenliang/myf-go/client/mhttp"
	"github.com/owenliang/myf-go/client/mmongo"
	"github.com/owenliang/myf-go/client/mmysql"
	"github.com/owenliang/myf-go/client/mredis"
	"github.com/owenliang/myf-go/conf"
	"github.com/owenliang/myf-go/mcontext"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

// 框架实例
type MyfApp struct {
	Gin *gin.Engine
}

// 创建APP
func New() (app *MyfApp, err error) {
	// 加载配置
	if err = conf.InitConf(); err != nil {
		return
	}

	// 创建APP
	app = &MyfApp{}

	// 创建Gin
	app.initGin()

	// 创建客户端
	if err = app.initClients(); err != nil {
		return
	}

	// 注册中间件
	app.registerMiddleware()

	return
}

// 启动APP
func (app *MyfApp) Run() (err error) {
	srv := &http.Server{
		Addr:         conf.MyfConf.AppConfig.Addr,
		Handler:      app.Gin,
		ReadTimeout:  time.Duration(conf.MyfConf.AppConfig.ReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(conf.MyfConf.AppConfig.WriteTimeout) * time.Millisecond,
	}

	//启动server服务
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			errMsg := fmt.Sprintf("启动server失败：%+v, %+v ", srv, err)
			panic(errMsg)
		}
	}()

	//等待优雅退出命令
	app.waitGraceExit(srv)
	return
}

// 闭包myf上下文
func (app *MyfApp) WithMyfContext(myfHandle mcontext.MyfHandleFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 请求超时控制
		timeoutCtx, cancelFunc := context.WithTimeout(c, time.Duration(conf.MyfConf.AppConfig.HandleTimeout)*time.Millisecond)
		defer cancelFunc()

		// Myf上下文
		myfCtx := mcontext.MyfContext{
			Context: timeoutCtx,
			Gin:            c,
			Cat:            c.Value("cat").(*cat.MyfCat),
		}

		// 回调用户
		myfHandle(&myfCtx)
	}
}

// 初始化Gin
func (app *MyfApp) initGin() {
	if conf.MyfConf.Debug != 0 {
		gin.SetMode(gin.DebugMode) // 调试模式
	} else {
		gin.SetMode(gin.ReleaseMode) // 生产模式
	}
	app.Gin = gin.New()
}

// 注册中间件
func (app *MyfApp) registerMiddleware() {
	// 日志
	app.Gin.Use(middlewares.Logger())
	// Metadata
	app.Gin.Use(middlewares.MetadataMiddleware())
	// CAT
	app.Gin.Use(middlewares.Cat())
	// Recovery崩溃日志(一定要放在最后一位, 否则panic会跳过其下的middleware)
	app.Gin.Use(middlewares.Recovery())
}

// 初始化各种客户端
func (app *MyfApp) initClients() (err error) {
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

// 等待优雅退出
func (app *MyfApp) waitGraceExit(server *http.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			fmt.Fprintf(os.Stdout, "收到信号: %s, 服务正在退出... \n", s.String())
			server.Close()
			return
		case syscall.SIGHUP:
		default:
		}
	}
}

// 初始化运行环境
func init() {
	runtime.GOMAXPROCS(runtime.NumCPU()) // 用满所有核心
}
