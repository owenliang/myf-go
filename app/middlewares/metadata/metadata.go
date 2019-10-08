package metadata

import "github.com/gin-gonic/gin"

// 请求链路元数据
type MetaData struct {
	Gin *gin.Context
}
