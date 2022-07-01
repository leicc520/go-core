package core

import (
	"context"
	
	"github.com/leicc520/go-orm/log"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/apache/rocketmq-client-go/v2/rlog"
)

/***********************************************************************
	topic专题可能需要运维rocketMQ的人员手动创建
 */
type ProducerMQHandler func(topic string)
type ConsumerMQHandler func(*primitive.MessageExt)
type RocketMQ struct {
	NameSrv []string
	SyncProducer rocketmq.Producer
	PushConsumer rocketmq.PushConsumer
}

//创建一个执行实例
func NewRocketMQ(namesrv ...string) *RocketMQ {
	rlog.SetLogLevel("error")
	return &RocketMQ{NameSrv: namesrv}
}

//释放类库的资源信息
func (r *RocketMQ) Release()  {
	if r.SyncProducer != nil {//生产者
		r.SyncProducer.Shutdown()
	}
	if r.PushConsumer != nil {//生产者
		r.PushConsumer.Shutdown()
	}
}

/**
发送一条消息到队列当中
1s 5s 10s 30s 1m 2m 3m 4m 5m 6m 7m 8m 9m 10m 20m 30m 1h 2h
 **********************************************************************/
func (r *RocketMQ) Send(topic string, message []byte) (*primitive.SendResult, error) {
	msg := primitive.NewMessage(topic, message)
	if res, err := r.SyncProducer.SendSync(context.Background(), msg); err != nil {
		log.Write(log.ERROR, "rocketmq send message error topic{", topic, "}", err)
		return nil, err
	} else {
		log.Write(log.INFO, "rocketmq send message ok topic{", topic, "}:", res.MsgID, ":", res.OffsetMsgID)
		return res, err
	}
}

//执行生产则业务处理逻辑
func (r *RocketMQ) ProducerStart(group string) {
	r.SyncProducer, _ = rocketmq.NewProducer(
		producer.WithNsResolver(primitive.NewPassthroughResolver(r.NameSrv)),
		producer.WithGroupName(group),
		producer.WithRetry(2),
	)
	if err := r.SyncProducer.Start(); err != nil {
		log.Write(log.ERROR, "start rocketMQ producer error ", err)
		panic(err) //启动生产者服务的处理逻辑
	}
}

//拉取消费消息Push的处理逻辑 独立协程中执行拉取消息
func (r *RocketMQ) ConsumerStart(group, topic string, handler ConsumerMQHandler) {
	r.PushConsumer, _ = rocketmq.NewPushConsumer(
		consumer.WithGroupName(group),
		consumer.WithNsResolver(primitive.NewPassthroughResolver(r.NameSrv)),
	)
	//defer pushConsumer.Shutdown() //释放资源
	err := r.PushConsumer.Subscribe(topic, consumer.MessageSelector{}, func(ctx context.Context,
		messages ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for idx := range messages {//数据信息
			log.Write(log.ERROR, "subscribe push message", string(messages[idx].Body))
			handler(messages[idx]) //请求数据
		}
		return consumer.ConsumeSuccess, nil
	})
	if err != nil {
		log.Write(log.ERROR, "start push consumer start ", err)
		panic(err) //启动生产者服务的处理逻辑
	}
	if err = r.PushConsumer.Start(); err != nil {
		log.Write(log.ERROR, "start push consumer start ", err)
		panic(err) //启动生产者服务的处理逻辑
	}
}
