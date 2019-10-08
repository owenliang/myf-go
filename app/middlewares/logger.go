package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/owenliang/myf-go/conf"
)

// 日志
func Logger() gin.HandlerFunc {
	// 目前logger仅支持输出到终端，并且仅在debug模式下输出
	if conf.MyfConf.Debug != 0 {
		return gin.Logger()
	} else { // 生产模式不输出日志
		return func(context *gin.Context) {
			context.Next()
		}
	}
}
