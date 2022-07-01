package core

import (
	"github.com/gin-gonic/gin"
)

type HttpError struct {
	Code int     `json:"code"`
	Msg string   `json:"msg"`
	Debug string `json:"debug"`
}

func (self *HttpError) Error() string {
	return self.Msg
}

//出错的模式情况
func (self *HttpError) SetDebug(msg string) *HttpError {
	self.Debug = msg
	return self
}

func (self *HttpError) ToMap() map[string]interface{} {
	data := map[string]interface{}{"code":self.Code, "debug":self.Debug, "msg":self.Msg}
	return data
}

type HttpView struct {
	HttpError
	ctx *gin.Context
	Datas interface{} `json:"datas"`
}

//new 创建一个对外的执行示例
func NewHttpView(ctx *gin.Context) *HttpView {
	view := &HttpView{ctx: ctx}
	view.Msg, view.Code = "OK", 0
	return view
}

//出错的模式情况
func (c *HttpView)ErrorDisplay(code int, msg string) {
	c.Msg, c.Code = msg, code
	c.ctx.JSON(200, c)
}

//json 数据模式格式化输出
func (c *HttpView)JsonDisplay(datas interface{}) {
	c.Datas = datas
	c.ctx.JSON(200, c)
}
