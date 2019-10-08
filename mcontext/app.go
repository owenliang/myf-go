package mcontext

// 二次封装gin的json返回方法
func (myfctx *MyfContext) JSON(code int, obj interface{}) {
	myfctx.Gin.JSON(code, obj)
}

// 二次封装gin的string返回方法
func (myfctx *MyfContext) String(code int, format string, values ...interface{}) {
	myfctx.Gin.String(code, format, values)
}