package core

import (
	"github.com/leicc520/go-orm/log"
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"time"
)

/***********************************************************************
topic专题可能需要运维KafkaMQ的人员手动创建
*/
type ClusterConsumerMQHandler func(msg *sarama.ConsumerMessage)
type KafkaMQ struct {
	NodeSrv []string //:9092 broker地址
	SyncProducer sarama.SyncProducer
	ClusterConsumer *cluster.Consumer
}

//创建一个执行实例
func NewKafkaMQ(nodesrv ...string) *KafkaMQ {
	return &KafkaMQ{NodeSrv: nodesrv}
}

//释放类库的资源信息
func (r *KafkaMQ) Release()  {
	if r.SyncProducer != nil {//生产者
		r.SyncProducer.Close()
	}
	if r.ClusterConsumer != nil {
		r.ClusterConsumer.Close()
	}
}

//发送一条消息到队列当中
func (r *KafkaMQ) Send(topic string, message []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message),
	}
	if _p, _offset, err := r.SyncProducer.SendMessage(msg); err != nil {
		log.Write(log.ERROR, "kafka send message error topic{", topic, "}", err)
		return err
	} else {
		log.Write(log.INFO, "kafka send message ok topic{", topic, "}:",  _p, ":", _offset)
	}
	return nil
}

/**
执行初始化生产者连接 数据资料信息
 ***************************************************************/
func (r *KafkaMQ) ProducerStart()  {
	var err error = nil
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true
	config.Producer.Timeout = 5 * time.Second
	r.SyncProducer, err = sarama.NewSyncProducer(r.NodeSrv, config)
	if err != nil {
		log.Write(log.ERROR, "kafka sync producer start error ",  err)
		panic(err)
	}
}

/**
拉取消费消息Push的处理逻辑 如果出现异常的情况可以这么处理 建议独立携程当中执行业务
把 sarama 版本改成 从 v1.xx.1 --> v1.24.1 就可以用啦 github.com/Shopify/sarama v1.24.1
gomod 的配置改下版本号就可以
github.com/Shopify/sarama v1.24.1
github.com/bsm/sarama-cluster v2.1.15+incompatible
 */
func (r *KafkaMQ) ConsumerStart(group, topic string, handler ClusterConsumerMQHandler) {
	var err error = nil
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	r.ClusterConsumer, err = cluster.NewConsumer(r.NodeSrv, group, []string{topic}, config)
	if err != nil {//失败的情况处理逻辑
		log.Write(log.ERROR, "cluster.NewConsumer error", err)
		panic(err)
	}

	go func() {//消费出错的异常日志
		for err := range r.ClusterConsumer.Errors() {
			log.Write(log.ERROR, "ClusterConsumer.Errors:", group, err)
		}
	}()
	go func() {//监听处理集群通知业务逻辑
		for ntf := range r.ClusterConsumer.Notifications() {
			log.Write(log.DEBUG, "ClusterConsumer.Notifications Rebalanced", group, ntf)
		}
	}()
	var msgStats int = 0
	for {//循环监听处理kafka的消息逻辑
		msg, ok := <-r.ClusterConsumer.Messages()
		if ok {
			handler(msg) //回调到外层业务逻辑的消费处理
			log.Writef(log.INFO, "ClusterConsumer.Messages {%s}:{%s}/%d/%d\t%s\t%s",
				group, msg.Topic, msg.Partition, msg.Offset, msg.Key, msg.Value)
			r.ClusterConsumer.MarkOffset(msg, "")  // mark message as processed
			msgStats++
		}
	}
	log.Writef(log.INFO, "%s consume %d messages", group, msgStats)
}
