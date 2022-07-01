package core

import (
	"fmt"
	"sync"
	"time"
)

/*
 * 算法解释
 * SnowFlake的结构如下(每部分用-分开):<br>
 * 0 - 0000000000 0000000000 0000000000 0000000000 0 - 00000 - 00000 - 000000000000 <br>
 * 1位标识，由于long基本类型在Java中是带符号的，最高位是符号位，正数是0，负数是1，所以id一般是正数，最高位是0<br>
 * 10位的数据机器位，可以部署在1024个节点，包括5位datacenterId和5位workerId<br>
 * 41位时间截(毫秒级)，注意，41位时间截不是存储当前时间的时间截，而是存储时间截的差值（当前时间截 - 开始时间截)
 * 得到的值），这里的的开始时间截，一般是我们的id生成器开始使用的时间，由我们程序来指定的（如下的epoch属性）。
 * 41位的时间截，可以使用69年，年T = (1L << 41) / (1000L * 60 * 60 * 24 * 365) = 69<br>
 * 12位序列，毫秒内的计数，12位的计数顺序号支持每个节点每毫秒(同一机器，同一时间截)产生4096个ID序号<br>
 * 加起来刚好64位，为一个Long型。<br>
 * SnowFlake的优点是，整体上按照时间自增排序，并且整个分布式系统内不会产生ID碰撞(由数据中心ID和机器ID作区分)，并且效率较高，经测试，SnowFlake每秒能够产生26万ID左右。
 */
const (
	//开始时间戳 2021-7-1
	epoch int64 = 1625068800000
	// 数据标识id所占的位数
	datacenterIdBits int64 = 10
	// 支持的最大数据标识id，结果是31
	maxDatacenterId int64 = -1 ^ (-1 << datacenterIdBits)
	//序列在id中占的位数
	sequenceBits int64 = 12
	//时间截向左移12位自身占据41个字节
	timestampLeftShift int64 = 41
	// 数据标识id向左移17位(41+12)
	datacenterIdShift int64  = sequenceBits + timestampLeftShift
	// 生成序列的掩码，这里为4095 (0b111111111111=0xfff=4095)
	sequenceMask int64 = -1 ^ (-1 << sequenceBits)
)

type SnowflakeIdWorker struct {
	mutex sync.Mutex // 添加互斥锁 确保并发安全
	lastTimestamp int64 // 上次生成ID的时间截 41bit
	datacenterId int64 //数据中心ID(0~1024) 10bit
	sequence int64 // 毫秒内序列(0~4095) 12bit
}

var (
	gSnowWorker map[string]*SnowflakeIdWorker = nil
)

//获取数据的单例模式，然后获取ID记录
func GetSnow(dId int64) *SnowflakeIdWorker {
	str := fmt.Sprintf("table@%d", dId)
	if gSnowWorker == nil {//数据为空的情况
		gSnowWorker = make(map[string]*SnowflakeIdWorker)
	}
	if worker, ok := gSnowWorker[str]; ok {
		return worker
	}
	worker, _ := NewCreateWorker(dId)
	gSnowWorker[str] = worker
	return worker
}

/*
 * 创建SnowflakeIdWorker
 * datacenterId 数据中心ID (0~1024) 标记table
 */
func NewCreateWorker(dId int64)(*SnowflakeIdWorker, error){
	if dId < 0 || dId > maxDatacenterId {
		dId = dId % maxDatacenterId
	}
	// 生成一个新节点
	return &SnowflakeIdWorker{
		lastTimestamp: 0,
		datacenterId: dId,
		sequence: 0,
	}, nil
}

//创建生成一个唯一的识别ID
func (w *SnowflakeIdWorker) GetId() int64 {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	nowTime := time.Now().UnixNano() / 1e6
	if  w.lastTimestamp == nowTime {
		w.sequence = (w.sequence + 1) & sequenceMask
		if w.sequence == 0 {
			// 阻塞到下一个毫秒，直到获得新的时间戳
			for nowTime <= w.lastTimestamp {
				nowTime = time.Now().UnixNano() / 1e6
			}
		}
	} else {//让随机的出现基数和偶数 散列
		w.sequence = 0
		if nowTime % 2 != 0 {
			w.sequence = 1
		}
	}
	w.lastTimestamp = nowTime
	ID := int64((w.datacenterId << datacenterIdShift) | ((nowTime - epoch) << sequenceBits) | w.sequence)
	return ID
}
