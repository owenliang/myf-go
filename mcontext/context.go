package mcontext

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/owenliang/myf-go/client/cat"
	context2 "github.com/owenliang/myf-go/cron/context"
)

// Myf框架请求上下文
type MyfContext struct {
	context.Context
	Cat *cat.MyfCat
	Gin *gin.Context
	Cron *context2.Context
}

// Myf框架controller定义
type MyfHandleFunc func(c *MyfContext)