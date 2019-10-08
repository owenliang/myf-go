package middlewares

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/owenliang/myf-go-cat/cat"
	cat2 "github.com/owenliang/myf-go/client/cat"
	"github.com/owenliang/myf-go/conf"
	"github.com/owenliang/myf-go/util"
)

// CAT监控中间件
func Cat() gin.HandlerFunc {
	// 初始化CAT
	if conf.MyfConf.CatConfig.IsOpen {
		if conf.MyfConf.Debug != 0 {
			cat.DebugOn()
		}
		cat.Init(conf.MyfConf.Domain, cat.Config{CatServerVersion: "V2"})
	}

	// 注册中间件
	return func(context *gin.Context) {
		myfCat := cat2.NewMyfCat()

		// 开启URL事务
		root := cat.NewTransaction("URL", util.ReplaceURI(context.Request.URL.Path))
		myfCat.Append(root)

		// 调用方的信息埋点
		catCallerDomain := context.GetHeader("_catCallerDomain")
		catCallerMethod := context.GetHeader("_catCallerMethod")
		if len(catCallerDomain) != 0 && len(catCallerMethod) != 0 {
			root.LogEvent("Request.from", fmt.Sprintf("%s/%s", catCallerDomain, catCallerMethod))
		}

		context.Set("cat", myfCat)

		// 调用下一个中间件
		context.Next()

		myfCat.FinishAll()
	}
}