package middlewares

import (
	"github.com/gin-gonic/gin"
	metadata2 "github.com/owenliang/myf-go/app/middlewares/metadata"
)

func MetadataMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		metadata := &metadata2.MetaData{
			Gin: context,
		}
		context.Set("metadata", metadata)
	}
}