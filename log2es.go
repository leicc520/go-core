package core

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"time"

	"github.com/leicc520/go-orm"
	"github.com/leicc520/go-orm/log"
	"github.com/olivere/elastic/v7"
	"gopkg.in/yaml.v2"
)

type Log2ElasticSt struct {
	Host string     `yaml:"host"`
	User string     `yaml:"user"`
	Password string `yaml:"password"`
	Index string    `yaml:"index"`
	RealIndex string
}

const ES_SRV_CONFIG = "es.srv.config"
var esEnv string
var esCli *elastic.Client = nil
var esInstance *Log2ElasticSt = nil
var log2esMapping = `{
    "mappings": {
        "properties": {
             "env": {
                "type": "keyword"
            },
            "level": {
                "type": "keyword"
            },
            "token": {
                "type": "keyword"
            },
            "agent": {
                "type": "text"
            },
			"url": {
                "type": "keyword"
            },
 	        "request": {
                "type": "text"
            },
            "info": {
                "type": "text"
            },
            "stack": {
                "type": "text"
            },
            "time": {
                "type": "date",
                "format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd||epoch_millis"
            }
        }
    }
}`

//单例模式获取es引擎
func GetEsInstance() *Log2ElasticSt {
	if esInstance != nil || esEnv == "NoSupport" {
		return esInstance
	}
	defer func() {//异常捕获不支持的情况
		if err := recover(); err != nil {
			esEnv = "NoSupport"
			log.Write(log.ERROR, err)
		}
	}()
	esInstance = &Log2ElasticSt{}
	data := NewMicRegSrv("").Config(ES_SRV_CONFIG)
	//把yaml形式的字符串解析成struct类型
	if err := yaml.Unmarshal([]byte(data), esInstance); err != nil {
		return nil
	}
	return esInstance
}

//启动的时候连接es，如果es连接失败的情况做逻辑处理
func (s *Log2ElasticSt) Init() *elastic.Client {
	var err error
	if esCli != nil || len(s.Host) < 3 || len(s.User) < 1 || len(s.Password) < 1 {
		return esCli//未开启es的支持的情况 || es已经完成初始化的情况
	}
	esCli, err = elastic.NewClient(elastic.SetSniff(false),
		elastic.SetBasicAuth(s.User, s.Password),
		elastic.SetURL(s.Host))
	if err != nil {
		log.Write(log.ERROR, "ES初始化出现异常:", err)
		esCli = nil
	}
	esEnv  = os.Getenv("DCENV")
	if len(esEnv) < 1 {
		esEnv = "prod"
	}
	return esCli
}

//检测索引是否存在，不存在的话创建新的索引-mapping
func (s *Log2ElasticSt) createIndexIfNotExists(es *elastic.Client) bool {
	realIndex := time.Now().Format("200601")
	if s.RealIndex == realIndex {//说明已经创建过了
		return true
	}
	s.RealIndex = realIndex
	exists, err := es.IndexExists(s.Index+s.RealIndex).Do(context.Background())
	if err != nil {
		log.Write(log.ERROR, "ES检测索引异常:", err)
		return false
	}
	if exists {//索引已经存在的情况
		return true
	}
	_, err = es.CreateIndex(s.Index+s.RealIndex).BodyString(log2esMapping).Do(context.Background())
	if err != nil {
		log.Write(log.ERROR, "ES创建索引异常:", err)
		return false
	}
	return true
}

//上报日志数据到es处理
func (s *Log2ElasticSt) Write2ES(mask int8, agent, url, request, token, info, stack string)  {
	es := s.Init() //检测，并主动创建索引
	if es == nil || !s.createIndexIfNotExists(esCli) {
		return
	}
	ckey  := fmt.Sprintf("es@%x", md5.Sum([]byte(info+url)))
	mcache:= orm.GetMCache()
	exists:= mcache.Get(ckey)
	if exists != nil {//数据为空的情况
		return //已经上报错误了直接跳过
	}
	mcache.Set(ckey, 1, 10) //10秒过期
	dtime:= time.Now().Format("2006-01-02 15:04:05")
	data := map[string]interface{}{"env":esEnv, "level":log.LevelName(mask), "token":token, "agent":agent,
		"url":url, "request":request, "info":info, "stack":stack, "time":dtime}
	_, err := es.Index().Index(s.Index+s.RealIndex).BodyJson(data).Do(context.Background())
	if err != nil {
		log.Write(log.ERROR, "es写入错误日志异常", err)
	}
}



