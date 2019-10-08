package middlewares

import (
	"context"
	cronContext "github.com/owenliang/myf-go/cron/context"
	"time"
)

func WithTimeout(duration time.Duration) cronContext.HandleFunc {
	return func(ctx *cronContext.Context) {
		// 超时熔断
		timeoutCtx, cancelFunc := context.WithTimeout(ctx.Context, duration)
		defer cancelFunc()

		ctx.Context = timeoutCtx

		ctx.Next()
	}
}