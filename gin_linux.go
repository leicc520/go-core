package core

import (
	"strings"
	
	"github.com/fvbock/endless"
	"github.com/leicc520/go-orm"
	"github.com/leicc520/go-orm/log"
)

//启动执行APP业务处理逻辑
func (app *Application) Start() {
	if len(app.handler) > 0 {
		for _, handle := range app.handler {
			handle(app.app)
		}
	}
	WritePidFile(app.config.Name) //写入进程pid数据资料信息
 	httpStr, wsStr, isSsl := "", "", false
	isSsl = strings.HasPrefix(strings.ToLower(app.config.Ssl), "on")
	if isSsl && orm.FileExists(app.config.KeyFile) && orm.FileExists(app.config.KeyFile) {
		httpStr = "https://"+app.config.Host
		if len(app.config.ImSeg) > 1 {
			wsStr = "wss://"+app.config.Host+app.config.ImSeg
		}
	} else {//配置阐述不对的情况
		isSsl   = false
		httpStr = "http://"+app.config.Host
		if len(app.config.ImSeg) > 1 {
			wsStr = "ws://"+app.config.Host+app.config.ImSeg
		}
	}
	log.Write(-1, "=======================start app linux=====================")
	log.Write(-1, "===http server "+httpStr)
	if len(wsStr) > 1 {
		log.Write(-1, "===websocket server "+wsStr)
	}
	log.Write(-1, "===========================================================")
	endSrv := endless.NewServer(app.config.Host, app.app)
	if isSsl {//针对https 热更新的处理逻辑
		if err := endSrv.ListenAndServeTLS(app.config.CertFile, app.config.KeyFile); err != nil {
			log.Write(-1, "start app failed:"+err.Error())
		}
	} else {//针对http 热更新的处理逻辑
		if err := endSrv.ListenAndServe(); err != nil {
			log.Write(-1, "start app failed:"+err.Error())
		}
	}
}
