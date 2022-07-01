package core

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/leicc520/go-orm"
	"github.com/leicc520/go-orm/log"
	"github.com/gin-gonic/gin"
)

const (
	JwtHeader = "SIGNATURE"
	JwtQuery  = "_s"
	JwtCookie = "_s"
	EncryptMd5= "md5req"
	EncryptKey= "X-KEY"
)

type AppConfigSt struct {
	Host string 		    `yaml:"host"`
	Name string 		    `yaml:"name"`
	Jwt string 			    `yaml:"jwt"`
	Ssl string 			    `yaml:"ssl"`
	Version string 		    `yaml:"version"`
	ImSeg string 		    `yaml:"im"`
	OssUrl string 		    `yaml:"ossUrl"`
	AppName string          `yaml:"appName"`
	BaseUrl string 		    `yaml:"baseUrl"`
	Domain string 		    `yaml:"domain"` //网站的域名
	CertFile string 	    `yaml:"certFile"`
	KeyFile string 		    `yaml:"keyFile"`
	CrossDomain string 	    `yaml:"crossDomain"`
	UpFileDir string 	    `yaml:"upfileDir"`
	UpFileBase string 	    `yaml:"upfileBase"`
}

type AppStartHandler func(c *gin.Engine)
type Application struct {
	app *gin.Engine
	config *AppConfigSt
	handler []AppStartHandler
}

var coConfig *AppConfigSt = nil

//初始化创建一个http服务的情况
func NewApp(config *AppConfigSt) *Application {
	coConfig = config
	app := &Application{app: gin.New(), handler: make([]AppStartHandler, 0), config: config}
	app.app.Use(gin.Logger(), GINRecovery())
	if strings.ToLower(config.CrossDomain) == "on" {
		app.app.Use(GINCors()) //跨域的支持集成
	}
	app.app.GET("/healthz", func(c *gin.Context) {
		c.String(200, config.Version)
	})
	app.app.GET("/errlog", func(c *gin.Context) {
		lStr := "OK"
		if c.Query("s") == "simlife" {
			lStr = log.ErrorLog()
		}
		c.String(200, lStr)
	})
	GinValidatorInit("zh")
	return app
}

//注册预先要执行的业务动作处理逻辑
func (app *Application) RegHandler(handler AppStartHandler) *Application {
	app.handler = append(app.handler, handler)
	return app
}

//启动之后写入pid file文件
func WritePidFile(name string)  {
	pidStr := strconv.FormatInt(int64(os.Getpid()), 10)
	pidFile:= "/run/"+name+".pid"
	if os.Getenv("DCENV") != "" {
		pidFile = "/run/"+name+"-"+os.Getenv("DCENV")+".pid"
	}
	os.WriteFile(pidFile, []byte(pidStr), 0777)
}

/***************************************************************************
 服务的管理放到的linux/windows当中，因为不同系统对优雅启动的支出不一致
 */
func DecryptBind(c *gin.Context, obj interface{}) error {
	skey := c.GetHeader(EncryptKey)
	if len(skey) > 6 {//解密的数据业务处理逻辑
		cryptSt := &Crypt{JKey: []byte(skey)}
		c.Set("cryptSt", cryptSt) //设置数据解码
		req, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Write(log.ERROR, "请求数据解码获取数据的时候异常", err)
			return err
		}
		log.Write(log.INFO, "数据接收:", string(req))
		defer c.Request.Body.Close()
		req = cryptSt.Decrypt(req)
		log.Write(log.INFO, "数据解码:", string(req))
		if req == nil || len(req) < 1 {
			log.Write(log.ERROR, "数据解码:", string(req))
			return errors.New("数据解码失败,无法操作.")
		}
		if err = json.Unmarshal(req, obj); err != nil {
			log.Write(log.ERROR, "数据解码:", string(req), err)
			return err
		}
		if err = ValidateStruct(obj); err != nil {
			log.Write(log.ERROR, "结构校验:", string(req), err)
			return err
		}
		c.Set(EncryptMd5, fmt.Sprintf("%x", md5.Sum(req)))
	} else {//非加密的业务处理逻辑
		if err := c.ShouldBind(&obj); err != nil {
			return  err
		}
	}
	return nil
}

//数据的加密处理逻辑
func EncryptBind(c *gin.Context, obj interface{}) interface{} {
	objSt, ok := c.Get("cryptSt")
	if !ok {
		return obj
	}
	if bstr, err := json.Marshal(obj); err != nil {
		return obj
	} else {//数据加密处理逻辑
		cryptSt := objSt.(*Crypt)
		if str, err := cryptSt.Encrypt(bstr); err == nil {
			return str
		}
	}
	return obj
}

//验证码登录 请求token 数据信息
func JWTACLCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := JWTACLToken(c)
		signUser := JwtUser{} //设置初始化信息
		if err := JwtParse(token, c.Request.UserAgent(), &signUser); err != nil {
			NewHttpView(c).ErrorDisplay(9999, "请求token异常")
			c.Abort()
			return
		}
		c.Set("user", &signUser) //设置请求的用户ID
		c.Next()
	}
}

//获取请求的token数据资料信息
func JWTACLToken(c *gin.Context) string {
	token := c.GetHeader(JwtHeader)
	if len(token) < 3 {
		token, _ = c.Cookie(JwtCookie)
		if len(token) < 3 {
			token = c.Query(JwtQuery)
		}
	}
	return token
}

//获取JWT登录授权校验uid
func JWTACLUserid(c *gin.Context) int64 {
	token := JWTACLToken(c)
	signUser := JwtUser{} //设置初始化信息
	if err := JwtParse(token, c.Request.UserAgent(), &signUser); err != nil {
		return -1
	}
	return signUser.Id
}

//计算接口请求的执行时间 并做业务错误拦截处理
func GINRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		sTime  := time.Now()
		defer func() {
			log.Write(log.DEBUG, c.Request.RequestURI, "执行时间:", time.Since(sTime))
			if err := recover(); err != nil {//执行panic数据恢复处理逻辑
				if o, ok := err.(*HttpError); ok {
					c.JSON(200, o.ToMap())
				} else {//未知的错误情况处理逻辑
					rtStack := orm.RuntimeStack(3)
					errStr, _  := json.Marshal(err)
					log.Write(log.ERROR, "GINRecovery", string(errStr), string(rtStack))
					//通过开启一个协成将奔溃日志收集到es上进行处理
					es := GetEsInstance() //独立获取es对象做业务
					if len(es.Host) > 3 && len(es.User) > 1 && len(es.Password) > 1 {
						c.Request.ParseForm() //请求解析form表单
						go es.Write2ES(log.ERROR, c.Request.UserAgent(), c.Request.RequestURI,
							c.Request.Form.Encode(), JWTACLToken(c), string(errStr), string(rtStack))
					}
					err := &HttpError{Code: 500, Msg: "内部服务错误,拒绝服务"}
					c.JSON(500, err.ToMap())
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}

//HTTP请求跨域的业务处理逻辑
func GINCors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")                                       // 这是允许访问所有域
		c.Header("Access-Control-Allow-Methods", "POST,GET,OPTIONS,PUT,DELETE,UPDATE") //服务器支持的所有跨域请求的方法,为了避免浏览次请求的多次'预检'请求
		c.Header("Access-Control-Allow-Headers", "SIGNATURE, Content-Length, X_Requested_With, X-KEY, Accept, Origin, Host, Accept-Encoding, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma")
		c.Header("P3P", "CP=\"CURa ADMa DEVa PSAo PSDo OUR BUS UNI PUR INT DEM STA PRE COM NAV OTC NOI DSP COR\"")
		c.Header("Access-Control-Allow-Credentials", "false")
		if strings.ToUpper(c.Request.Method) == "OPTIONS" {
			c.AbortWithStatus(202)
			return
		}
		c.Next() //  处理请求
	}
}
