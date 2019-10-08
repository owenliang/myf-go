# myf-go 

Myf框架的Golang框架核心库

## 特性

* cat埋点的客户端实现，包括：mysql、redis、http
* 订制的请求context，自动熔断超时请求，提供多种丰富特性
* 原生的gin操作体验

## 初始化项目

例如，新建项目地址：https://github.com/owenliang/myf-go-mvc，则依次执行如下：

```
go mod init github.com/owenliang/myf-go-mvc
go get -u github.com/owenliang/myf-go@最新tag版本
```

## 开发web应用

创建main.go如下：

```
package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/owenliang/myf-go/app"
	"github.com/owenliang/myf-go/mcontext"
	"os"
)

func Test(ctx *mcontext.MyfContext) {
	ctx.JSON(200, &gin.H{
		"error_code": 0,
		"error_msg": "",
		"data": nil,
	})
}

func main() {
	flag.Parse()

	if myfApp, err := app.New(); err == nil {
		myfApp.Gin.GET("/test", myfApp.WithMyfContext(Test))
		myfApp.Run()
	} else {
		fmt.Fprintln(os.Stderr, err)
	}
}
```

编写app.toml：

```
domain = "myf-go.smmyf.com"
debug = 1

# web模式
[App]
listen = "0.0.0.0:8087"

# cat配置
[Cat]
    isOpen = true
```

启动：

```
go run main.go -conf ./app.toml
```

# 开发cron定时任务

```
package main

import (
	"flag"
	"fmt"
	"github.com/owenliang/myf-go/cron"
	"github.com/owenliang/myf-go/cron/middlewares"
	"github.com/owenliang/myf-go/mcontext"
	"time"
)

func task1(myfCtx *mcontext.MyfContext) {
	// <- myfCtx.Done()
}

func main() {
	flag.Parse()

	if myfCron, err := cron.New(); err == nil {
		myfCron.AddJob("mytask", "* * * * *", middlewares.WithTimeout(5 * time.Second), myfCron.WithMyfContext(task1))
		myfCron.Run()
	} else {
		fmt.Println(err)
	}
}
```

编写app.toml：

```
domain = "myf-go.smmyf.com"
debug = 1

# cron模式
[Cron]
waitGraceExit = 5000

# cat配置
[Cat]
    isOpen = true
```

启动：

```
go run main.go -conf ./app.toml
```