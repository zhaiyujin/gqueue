/**
 * package: service
 * author: zhaiyujin
 * description:
 */
package gqueue

import (
	"encoding/json"
	"github.com/gogf/gf/container/gvar"
	"github.com/gogf/gf/database/gredis"
	"github.com/gogf/gf/frame/g"
	"github.com/gomodule/redigo/redis"
	"time"
)

//Redis redis cache
type Redis struct {
	Conn *gredis.Conn
}

type ZAdd struct {
	Score  float64
	Member interface{}
}

//redis向列表尾部插入指定值
func (r *Redis) RPush(key string, value string) error {
	_, err := g.Redis().Do("RPUSH", key, value)
	if err != nil {
		return err
	}
	return nil
}

//先进先出返回存储列表中第一个元素
func (r *Redis) LPop(key string) (string, error) {
	res, err := g.Redis().DoVar("lpop", key)
	return res.String(), err
}

//blpop功能和lpop类似但是blpop阻塞如果列表没有元素会阻塞列表直到等待超时或发现可弹出元素为止
func (r *Redis) BLPop(timeout time.Duration, keys ...string) ([]string, error) {
	var (
		res *gvar.Var
		err error
	)

	args := make([]interface{}, len(keys)+1)

	for i, key := range keys {
		args[i] = key
	}

	//处理时间
	if timeout > 0 {
		args[len(keys)] = int64(timeout / time.Second)
	} else {
		args[len(keys)] = 0
	}

	res, err = g.Redis().DoVar("BLPOP", args...)

	return res.Strings(), err
}

func (r *Redis) ZAdd(key string, z ...ZAdd) error {
	a := make([]interface{}, 2*len(z)+1)
	a[0] = key
	for i, m := range z {
		a[i*2+1] = m.Score
		a[i*2+2] = m.Member
	}
	_, err := g.Redis().DoVar("zadd", a...)
	return err
}

func (r *Redis) ZRemRangeByScore(key, start, stop string) (int, error) {
	res, err := g.Redis().DoVar("ZREMRANGEBYSCORE", key, start, stop)
	if err != nil {
		return 0, err
	}
	return res.Int(), err
}

func (r *Redis) ZRangeScores(key string, start, stop int64) ([]string, error) {
	res, err := g.Redis().DoVar("ZRANGEBYSCORE", key, start, stop)
	if err != nil {
		return nil, err
	}
	return res.Strings(), err
}

func (r *Redis) PipelineZRemRangeByScore(key, start, stop string) error {
	return r.Conn.Send("ZREMRANGEBYSCORE", key, start, stop)

}

func (r *Redis) PipelineZRangeScores(key string, start, stop int64) error {
	return r.Conn.Send("ZRANGEBYSCORE", key, start, stop)
}

func (r *Redis) TxPipelined() {
	panic("implement me")
}

//NewRedis 实例化
func NewRedis() *Redis {
	return &Redis{}
}

func (r *Redis) Send(fn func(r *Redis)) ([]string, error) {
	r.Conn = g.Redis().Conn()
	defer r.Conn.Close()
	fn(r)
	//刷新将输出缓冲区刷新到Redis服务器。
	r.Conn.Flush()
	res, err := r.Conn.ReceiveVar()
	r.Conn.Receive() //nolint:errcheck
	return res.Strings(), err

}

//Get 获取一个值
func (r *Redis) Get(key string) interface{} {

	var data []byte
	var err error
	if data, err = redis.Bytes(g.Redis().Do("GET", key)); err != nil {
		return nil
	}
	var reply interface{}

	if err = json.Unmarshal(data, &reply); err != nil {
		return nil
	}

	return reply
}

//Set 设置一个值
func (r *Redis) Set(key string, val interface{}, timeout time.Duration) (err error) {

	var data []byte
	if data, err = json.Marshal(val); err != nil {
		return err
	}
	if _, err = g.Redis().DoWithTimeout(timeout, "SET", key, data); err != nil {
		return err
	}

	return
}

//IsExist 判断key是否存在
func (r *Redis) IsExist(key string) bool {

	a, _ := g.Redis().Do("EXISTS", key)
	i := a.(int64)
	return i > 0
}

//Delete 删除
func (r *Redis) Delete(key string) error {

	if _, err := g.Redis().Do("DEL", key); err != nil {
		return err
	}

	return nil
}
