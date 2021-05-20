/**
 * package: delay
 * name:  producer
 * author: zhaiyujin
 * date: 2021/5/17 17:56
 * description:
 */
package gqueue

import (
	"encoding/json"
	"errors"
	uuid "github.com/satori/go.uuid"
	"time"
)

const (
	listSuffix, zsetSuffix = ":list", ":zset"
)

type Message struct {
	ID        string `json:"id"`
	Body      string `json:"body"`
	Timestamp int64  `json:"timestamp"`
	DelayTime int64  `json:"delayTime"`
}

func NewMessage(id string, body string) *Message {
	if id == "" {
		id = uuid.NewV4().String()
	}
	return &Message{
		ID:        id,
		Body:      body,
		Timestamp: time.Now().Unix(),
		DelayTime: time.Now().Unix(),
	}
}

type producer struct {
	redis RedisCmd
}

func NewProducer(cache RedisCmd) *producer {
	return &producer{redis: cache}
}

func (p *producer) Publish(topicName string, body string) error {
	msg := NewMessage("", body)
	sendData, _ := json.Marshal(msg)
	return p.redis.RPush(topicName+listSuffix, string(sendData))

}
func (p *producer) PublishDelay(topicName string, body string, delay time.Duration) error {
	msg := NewMessage("", body)
	if delay <= 0 {
		return errors.New("delay need great than zero")
	}
	tm := time.Now().Add(delay)
	msg.DelayTime = tm.Unix()
	member, _ := json.Marshal(msg)

	return p.redis.ZAdd(topicName+zsetSuffix, ZAdd{
		Score:  float64(tm.Unix()),
		Member: string(member),
	})

}
