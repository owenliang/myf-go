package middlewares

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	cat2 "github.com/owenliang/myf-go/client/cat"
	"github.com/owenliang/myf-go/conf"
	"os"
	"runtime"
)

// 调用栈
func stack(skip int) string {
	buf := new(bytes.Buffer)

	for i := skip; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
	}
	return buf.String()
}

// 崩溃日志中间件
func Recovery() gin.HandlerFunc {
	// 注册中间件
	return func(context *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				if cat, exists := context.Get("cat"); exists {
					if tran := cat.(*cat2.MyfCat).Top(); tran != nil {
						s := stack(3)
						tran.LogEvent("Error", "Recovery", s)
						tran.SetStatus(s)
					}
				}
				// 调试输出到命令行
				if conf.MyfConf.Debug != 0 {
					fmt.Fprintln(os.Stderr, err, stack(3))
				}
			}
		}()

		// 调用下一个中间件
		context.Next()
	}
}
