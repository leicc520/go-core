package core

import (
	"crypto/md5"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/leicc520/go-orm"
)

type HashSt struct {
	JKey []byte
}

const DOTSEGS = ","

var DefaultHash HashSt

func init()  {
	DefaultHash.SetJKey([]byte{44,56,77,90,12,123,65,77,58,96,87,80,25,66,99})
}

//设置密钥数据信息
func (c *HashSt) SetJKey(jkey []byte)  {
	c.JKey = jkey
}

//数据的加密处理逻辑 涉及压缩+抑或加密
func (c *HashSt) Encrypt(str string) string {
	str    += DOTSEGS+strconv.FormatInt(time.Now().Unix(), 10)
	str     = orm.SwapStringCrypt(str)
	buffer := append(c.JKey, []byte(str)...)
	buffer  = append(buffer, c.JKey...)
	md5str := fmt.Sprintf("%x", md5.Sum(buffer))
	return md5str[28:31]+str+md5str[0:3]
}

//解密算法处理逻辑
func (c *HashSt) Decrypt(str string, expire int64) string {
	if len(str) < 8 {//长度不足的情况
		return ""
	}
	ostr   := str[3:len(str)-3]
	buffer := append(c.JKey, []byte(ostr)...)
	buffer  = append(buffer, c.JKey...)
	md5str := fmt.Sprintf("%x", md5.Sum(buffer))
	if md5str[0:3] == str[len(str)-3:] && md5str[28:31] == str[0:3] {
		ostr  = orm.SwapStringCrypt(ostr)
		astr := strings.SplitN(ostr, DOTSEGS, 2)
		if len(astr) == 2 && expire == 0 {//数据不为空的情况
			return astr[0]
		}
		stime, err := strconv.ParseInt(astr[1], 10, 64)
		if err != nil || stime > expire {
			return astr[0]
		}
	}
	return ""
}
