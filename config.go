package core

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/leicc520/go-orm"
	"github.com/leicc520/go-orm/cache"
	logv2 "github.com/leicc520/go-orm/log"
	"gopkg.in/yaml.v2"
)

type Config struct {
	ConfAddr string	//配置的加载地址 或者 文件配置
	App AppConfigSt	   		   `yaml:"app"`
	Logger logv2.LogFileSt	   `yaml:"logger"`
	Redis  string 			   `yaml:"redis"`
	DiSrv  string 			   `yaml:"disrv"`
	CacheDir string            `yaml:"cachedir"`
	Cache  cache.CacheConfigSt `yaml:"cache"`
}

const DBPREFIX = "go.config.db."

//判断是否是目录结构
func IsDir(dir string) bool {
	ss, err := os.Stat(dir)
	if err == nil && ss.IsDir() {
		return true
	}
	return false
}

//加载配置文件数据信息
func (c *Config) Load(config interface{}) *Config {
	workdir := os.Getenv("WORKDIR")
	if len(workdir) > 0 && IsDir(workdir)  {
		os.Chdir(workdir)
	}
	file, err := os.Stat(c.ConfAddr)
	if err == nil && file.Mode().IsRegular() {
		c.LoadFile(c.ConfAddr, config)
	} else {
		c.LoadAddr(c.ConfAddr, config)
	}
	InitJwtSecret(c.App.Jwt) //初始化jwt
	if len(c.CacheDir) > 0 {
		cache.GFileCacheDir = c.CacheDir
	}
	workdir, err = os.Getwd()
	logv2.Write(-1, "workdir {"+workdir+"} cachedir {"+cache.GFileCacheDir+"}", err)
	return c
}

//加载配置 数据资料信息
func (c *Config)LoadFile(confFile string, config interface{}) *Config {
	if confFile == "" {
		confFile = "config/default.yml"
	}
	if file, err:=os.Stat(confFile); err != nil || !file.Mode().IsRegular() {
		log.Fatalln("load Config File Failed: "+err.Error())
	}
	data, _ := ioutil.ReadFile(confFile)
	//把yaml形式的字符串解析成struct类型 先子类初始化
	if err := yaml.Unmarshal(data, config); err != nil {
		log.Fatalln("load Config child Parse Failed:"+err.Error())
	}
	//把yaml形式的字符串解析成struct类型 父类加载初始化
	if err := yaml.Unmarshal(data, c); err != nil {
		log.Fatalln("load Config parent Parse Failed:"+err.Error())
	}
	return 	c
}

//加载配置 通过配置加载数据
func (c *Config)LoadAddr(srvAddr string, config interface{}) *Config {
	data := NewMicRegSrv(srvAddr).Config(c.App.Name)
	//把yaml形式的字符串解析成struct类型
	if err := yaml.Unmarshal([]byte(data), config); err != nil {
		log.Fatalln("load Config child Parse Failed:"+err.Error())
	}
	//把yaml形式的字符串解析成struct类型
	if err := yaml.Unmarshal([]byte(data), c); err != nil {
		log.Fatalln("load Config parent Parse Failed:"+err.Error())
	}
	return 	c
}

//通过配置名称加载配置，然后解析到config配置当中
func (c *Config)LoadConfig(srvAddr, confName string, config interface{}) error {
	data := NewMicRegSrv(srvAddr).Config(confName)
	//把yaml形式的字符串解析成struct类型
	if err := yaml.Unmarshal([]byte(data), config); err != nil {
		return err
	}
	return nil
}

//通过远程配置服务器加载
func (c *Config)LoadDBRemote(name string, srvAddr string) {
	dbName   := DBPREFIX+name
	data     := NewMicRegSrv(srvAddr).Config(dbName)
	dbSlice  := make([]orm.DbConfig, 0)
	//把yaml形式的字符串解析成struct类型
	if err := yaml.Unmarshal([]byte(data), &dbSlice); err != nil {
		log.Fatalln("load Config {"+dbName+"} Parse Failed:"+err.Error())
	}
	for idx := 0; idx < len(dbSlice); idx++ {
		orm.InitDBPoolSt().Set(dbSlice[idx].SKey, &dbSlice[idx])
	}
}

