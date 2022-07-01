package core

import (
	"github.com/go-redis/redis"
	"time"
)

const AUTOUNLOCKTIME = time.Second*30

type RdsLocker struct {
	state bool
	RedisSt *redis.Client
}

//给指定的key 上锁，记得defer自动解锁
func (l *RdsLocker) Lock(ckey string) bool {
	result := l.RedisSt.SetNX(ckey, 1, AUTOUNLOCKTIME)
	l.state = result.Val()
	return l.state
}

//解锁处理逻辑
func (l *RdsLocker) UnLock(ckey string)  {
	if l.state {//上锁成功的情况解锁
		l.RedisSt.Del(ckey)
	}
}
