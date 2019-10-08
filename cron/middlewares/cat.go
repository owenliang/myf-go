package middlewares

import (
	"fmt"
	"github.com/owenliang/myf-go-cat/cat"
	cat2 "github.com/owenliang/myf-go/client/cat"
	"github.com/owenliang/myf-go/conf"
	cronContext "github.com/owenliang/myf-go/cron/context"
)

func Cat() cronContext.HandleFunc {
	// 初始化CAT
	if conf.MyfConf.CatConfig.IsOpen {
		if conf.MyfConf.Debug != 0 {
			cat.DebugOn()
		}
		cat.Init(conf.MyfConf.Domain, cat.Config{CatServerVersion: "V2"})
	}

	return func(context *cronContext.Context) {
		myfCat := cat2.NewMyfCat()

		// 开启URL事务
		root := cat.NewTransaction("JOB", fmt.Sprintf("%s(%s)", context.Name, context.Spec))
		myfCat.Append(root)

		context.Set("cat", myfCat)

		// 调用下一个中间件
		context.Next()

		myfCat.FinishAll()
	}
}
