package context

import (
	"golang.org/x/net/context"
)

// Cron的执行上下文, 类比于gin.Context

type Context struct {
	context.Context
	Name string // 任务名
	Spec string // 任务cron表达式
	Chain []HandleFunc// 中间件调用链
	ChainIndex int // 执行到chain的哪一环
	KeyValues map[string]interface{} // 存储KV
}

type HandleFunc func(*Context)	// Cron处理函数

// 执行下一个中间件
func (ctx *Context) Next() {
	curIndex := ctx.ChainIndex
	if curIndex >= len(ctx.Chain) {
		return
	}
	ctx.ChainIndex++
	ctx.Chain[curIndex](ctx)
}

func (ctx *Context) Set(key string, value interface{}) {
	ctx.KeyValues[key] = value
}

func (ctx *Context) Get(key string) (value interface{}, exist bool) {
	value, exist = ctx.KeyValues[key]
	return
}

// 复写context.Context的Value方法, 提供键值能力
func (ctx *Context) Value(key interface{}) (value interface{}) {
	if strKey, ok := key.(string); ok {
		value, _ = ctx.KeyValues[strKey]
	}
	return
}