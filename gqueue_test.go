/**
 * package: gqueue
 * name:  gqueue_test
 * author: zhaiyujin
 * date: 2021/5/18 10:15
 * description:
 */
package gqueue

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

//队列消费
func TestPublishDelay(t *testing.T) {
	//延迟30s

	producer := NewProducer(NewRedis())

	for i := 0; i <= 10; i++ {
		req, _ := json.Marshal(MyMsg{
			Name: "ouyang",
			Age:  i,
		})
		err := producer.PublishDelay("zdd", string(req), time.Second*10) //nolint:errcheck
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println("success")
	}

	ctx, cancel := context.WithCancel(context.Background()) //nolint:govet
	stopCh := make(chan os.Signal)
	signal.Notify(stopCh, os.Interrupt)
	consumer := NewMQConsumer(ctx, NewRedis(), "zdd", NewRateLimitPeriod(time.Second))
	consumer.SetHandler(&MyHandler{})
	<-stopCh
	cancel()
	fmt.Println("stop server")
}

//延迟队列
func TestPublish(t *testing.T) {

	producer := NewProducer(NewRedis())
	req, _ := json.Marshal(MyMsg{
		Name: "ouyang",
		Age:  40,
	})

	err := producer.Publish("list", string(req)) //nolint:errcheck
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("success")

	ctx, cancel := context.WithCancel(context.Background()) //nolint:govet
	stopCh := make(chan os.Signal)
	signal.Notify(stopCh, os.Interrupt)
	consumer := NewMQConsumer(ctx, NewRedis(), "list", NewRateLimitPeriod(time.Second))
	consumer.SetHandler(&MyHandler{})
	<-stopCh
	cancel()
	fmt.Println("stop server")
}

func TestLpush(t *testing.T) {
	for i := 0; i < 10; i++ {
		err := NewProducer(NewRedis()).Publish("queue", string(rune(i))) //nolint:errcheck
		if err != nil {
			fmt.Println(err.Error())
		}
	}

}

func TestNs(t *testing.T) {
	ch := make(chan os.Signal)
	//如果不指定要监听的信号，那么默认是所有信号
	signal.Notify(ch, syscall.SIGKILL)

	//ch将一直阻塞在这里，因为它将收不到任何信号
	//所以下面的exit输出也无法执行
	<-ch
	fmt.Println("exit")

}
func TestConsumerLpop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background()) //nolint:govet
	consumer := NewMQConsumer(ctx, NewRedis(), "queue", NewRateLimitPeriod(time.Second))
	consumer.SetHandler(&MyHandler{})

	stopCh := make(chan os.Signal)
	signal.Notify(stopCh, os.Interrupt)
	<-stopCh
	cancel()
	fmt.Println("stop server")

}

func TestRpushData(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background()) //nolint:govet
	consumer := NewMQConsumer(ctx, NewRedis(), "queue")
	consumer.SetHandler(&MyHandler{})
	go func() {
		ticker := time.NewTicker(time.Second / 10)
		producer := NewProducer(NewRedis())
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("stop produce...")
				return
			case <-ticker.C:
				msg := &MyMsg{
					Name: fmt.Sprintf("name_%d", rand.Int()),
					Age:  rand.Intn(20),
				}

				body, _ := json.Marshal(msg)

				body2, _ := json.Marshal(Message{
					ID:        "",
					Body:      string(body),
					Timestamp: 111111,
					DelayTime: 111111,
				})
				if err := producer.Publish("queue", string(body2)); err != nil {
					panic(err)
				}
				/*	if err := producer.Publish("queue", body, time.Second); err != nil {
					panic(err)
				}*/
			}

		}
	}()

	stopCh := make(chan os.Signal)
	signal.Notify(stopCh, os.Interrupt)
	<-stopCh
	cancel()
	fmt.Println("stop server")
}

type MyMsg struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type QueueHandler struct {
}

func (q *QueueHandler) HandleMessage(m *Message) {
	fmt.Printf("receive msg: %v \n", m)
}

type MyHandler struct{}

func (*MyHandler) HandleMessage(m *Message) {
	revMsg := &MyMsg{}

	if err := json.Unmarshal([]byte(m.Body), revMsg); err != nil {
		fmt.Printf("handle message error: %#v \n", err)
		return
	}
	fmt.Printf("receive msg: %#v \n", *revMsg)
}
