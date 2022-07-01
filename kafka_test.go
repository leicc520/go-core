package core

import (
	"fmt"
	"github.com/Shopify/sarama"
	"testing"
	"time"
)

/**
kafka 相同的group 的话 可以重复取到消息
 */
func TestKafka(t *testing.T) {
	mqv := NewKafkaMQ("192.168.138.128:9092")
	go mqv.ConsumerStart("demo1", "test", func(msg *sarama.ConsumerMessage) {
		fmt.Println("routing 1")
		fmt.Println(msg.Offset, msg.Partition, msg.Topic, string(msg.Value), string(msg.Key))
	})
	mq := NewKafkaMQ("192.168.138.128:9092")
	go mq.ConsumerStart("demo", "test", func(msg *sarama.ConsumerMessage) {
		fmt.Println("routing 2")
		fmt.Println(msg.Offset, msg.Partition, msg.Topic, string(msg.Value), string(msg.Key))
	})
	mq.ProducerStart()
	idx := 0
	for  {
		msg := fmt.Sprintf("demo test msg %d", idx)
		mq.Send("test", []byte(msg))
		time.Sleep(time.Second*1)
	}
	t.Log("end.")
}
