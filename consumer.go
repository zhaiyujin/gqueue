/**
 * package: delay
 * name:  consumer
 * author: zhaiyujin
 * date: 2021/5/17 18:04
 * description:
 */
package gqueue

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/util/gconv"
	"log"
	"strconv"
	"sync"
	"time"
)

type Handler interface {
	HandleMessage(message *Message)
}

type consumer struct {
	ctx       context.Context
	once      sync.Once
	redis     RedisCmd
	handler   Handler
	options   ConsumerOptions
	topicName string
}

//自定义执行
func (c *consumer) SetHandler(handler Handler) {
	c.handler = handler
	c.once.Do(func() {
		c.startGetList()
		c.startGetDelayList()
	})

}
func NewRateLimitPeriod(d time.Duration) ConsumerOption {
	return func(o *ConsumerOptions) {
		o.RateLimitPeriod = d
	}
}

//消费者配置
type ConsumerOptions struct {
	RateLimitPeriod time.Duration
	UseBLPop        bool
}
type ConsumerOption func(options *ConsumerOptions)

//创建消费者实例
func NewMQConsumer(ctx context.Context, rediscmd RedisCmd, topicName string, opts ...ConsumerOption) *consumer {
	c := &consumer{
		ctx:       ctx,
		redis:     rediscmd,
		topicName: topicName,
	}

	for _, o := range opts {
		o(&c.options)
	}
	//1微秒等于百万分之一秒（10的负6次方秒）
	if c.options.RateLimitPeriod == 0 {
		c.options.RateLimitPeriod = time.Microsecond * 200
	}
	return c
}

func (c *consumer) startGetList() {

	go func() {
		//设置打点器，每隔默认是200s执行
		ticker := time.NewTicker(c.options.RateLimitPeriod)
		//运行终止关闭打点器
		defer func() {
			log.Println("end ticker & stop  get list message.")
			ticker.Stop()
		}()
		//组合list 的key
		topicName := c.topicName + listSuffix
		for {
			select {
			case <-c.ctx.Done():
				log.Printf("context Done msg: %#v \n", c.ctx.Err())
				return
			case <-ticker.C:
				var (
					revBody string
					err     error
					res     []string
				)

				//判断使用lpop还是blpop
				if !c.options.UseBLPop {
					revBody, err = c.redis.LPop(topicName)
				} else {

					res, err = c.redis.BLPop(time.Second, topicName)
					if len(res) >= 2 {
						revBody = res[1]
					}
				}
				if err != nil {
					log.Printf("LPOP error: %#v \n", err)
					continue
				}
				if revBody == "" {
					continue
				}

				msg := &Message{}
				json.Unmarshal([]byte(revBody), msg) //nolint:errcheck
				g.Dump(msg)
				if c.handler != nil {
					c.handler.HandleMessage(msg)
				}
			}
		}

	}()
}

func (c *consumer) startGetDelayList() {
	go func() {
		ticker := time.NewTicker(c.options.RateLimitPeriod)
		defer func() {
			log.Println("stop get delay message.")
			ticker.Stop()
		}()
		topicName := c.topicName + zsetSuffix
		for {
			currentTime := time.Now().Unix()
			select {
			case <-c.ctx.Done():
				log.Printf("context Done msg: %#v \n", c.ctx.Err())
				return
			case <-ticker.C:
				var (
					rev []string
					err error
				)
				/*
					普通方式
					rev, err = c.redis.ZRangeScores(topicName, 0, currentTime)
					//nolint:errcheck
					fmt.Println(rev, topicName, currentTime)
					_, err = c.redis.ZRemRangeByScore(topicName, "0", strconv.FormatInt(currentTime, 10)) //nolint:errcheck*/
				rev, err = c.redis.Send(
					func(r *Redis) {
						r.PipelineZRangeScores(topicName, 0, currentTime)                     //nolint:errcheck
						r.PipelineZRemRangeByScore(topicName, "0", gconv.String(currentTime)) //nolint:errcheck
					},
				)

				if err != nil {
					fmt.Println(err.Error())
				}

				if err != nil {
					log.Printf("zset pip error: %#v \n", err)
					continue
				}

				for _, revBody := range rev {
					msg := &Message{}

					json.Unmarshal([]byte(revBody), msg)
					if c.handler != nil {
						g.Dump("handler")
						c.handler.HandleMessage(msg)
					}
				}

				/*		var valuesCmd *redis.ZSliceCmd
						_, err := c.redisCmd.TxPipelined(func(pip redis.Pipeliner) error {
							valuesCmd = pip.ZRangeWithScores(topicName, 0, currentTime)
							pip.ZRemRangeByScore(topicName, "0", strconv.FormatInt(currentTime, 10))
							return nil
						})
						if err != nil {
							log.Printf("zset pip error: %#v \n", err)
							continue
						}
						rev := valuesCmd.Val()
						for _, revBody := range rev {
							msg := &Message{}
							json.Unmarshal([]byte(revBody.Member.(string)), msg)
							if c.handler != nil {
								c.handler.HandleMessage(msg)
							}
						}*/
			}
		}
	}()
}

// Strval 获取变量的字符串值
// 浮点型 3.0将会转换成字符串3, "3"
// 非数值或字符类型的变量将会被转换成JSON格式字符串
func Strval(value interface{}) string {
	// interface 转 string
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return key
}

/*
func (s *consumer) startGetDelayMessage() {
	go func() {
		ticker := time.NewTicker(s.options.RateLimitPeriod)
		defer func() {
			log.Println("stop get delay message.")
			ticker.Stop()
		}()
		topicName := s.topicName + zsetSuffix
		for {
			currentTime := time.Now().Unix()
			select {
			case <-s.ctx.Done():
				log.Printf("context Done msg: %#v \n", s.ctx.Err())
				return
			case <-ticker.C:
				var valuesCmd *redis.ZSliceCmd
				_, err := s.redisCmd.TxPipelined(func(pip redis.Pipeliner) error {
					valuesCmd = pip.ZRangeWithScores(topicName, 0, currentTime)
					pip.ZRemRangeByScore(topicName, "0", strconv.FormatInt(currentTime, 10))
					return nil
				})
				if err != nil {
					log.Printf("zset pip error: %#v \n", err)
					continue
				}
				rev := valuesCmd.Val()
				for _, revBody := range rev {
					msg := &Message{}
					json.Unmarshal([]byte(revBody.Member.(string)), msg)
					if s.handler != nil {
						s.handler.HandleMessage(msg)
					}
				}
			}
		}
	}()
}
*/
