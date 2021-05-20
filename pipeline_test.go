/**
 * package: gqueue
 * name:  pipeline_test
 * author: zhaiyujin
 * date: 2021/5/19 17:17
 * description:
 */
package gqueue

import (
	"fmt"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/util/gconv"
	"testing"
	"time"
)

/**
Send将命令写入连接的输出缓冲区。Flush将连接的输出缓冲区刷新到服务器。接收从服务器读取单个答复。
*/

func TestPipeLine(t *testing.T) {

	topicName := "zdd:zset"
	currentTime := time.Now().Unix()

	re, err := NewRedis().Send(
		func(r *Redis) {
			r.PipelineZRangeScores(topicName, 0, currentTime)                     //nolint:errcheck
			r.PipelineZRemRangeByScore(topicName, "0", gconv.String(currentTime)) //nolint:errcheck
		},
	)

	if err != nil {
		fmt.Println(err.Error())
	}
	g.Dump(gconv.Strings(re))
	/*rev, err = redis.NewRedis().ZRangeScores(topicName, 0, currentTime)

	_, err = redis.NewRedis().ZRemRangeByScore(topicName, "0", strconv.FormatInt(currentTime, 10)) */ //nolint:errcheck

}

//压测一下普通方式BenchmarkPipeLine-4   	   16549	     68624 ns/op
func BenchmarkNormal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v1, _ := g.Redis().DoVar("SET", "foo", "bar")
		v2, _ := g.Redis().DoVar("GET", "foo")
		fmt.Println(v1)
		fmt.Println(v2)
	}
}

//压测一下通道方式BenchmarkPipeLine-4   	   24122	     55599 ns/op
func BenchmarkPipeLine(b *testing.B) {
	conn := g.Redis().Conn()
	defer conn.Close()
	for i := 0; i < b.N; i++ {
		conn.Send("SET", "foo", "bar")
		conn.Send("GET", "foo")
		conn.Flush()
		// reply from SET
		v1, _ := conn.Receive()
		fmt.Println(v1)
		// reply from GET
		v2, _ := conn.Receive()
		fmt.Println(gconv.String(v2))
	}
}

func BenchmarkPipeLine2(b *testing.B) {
	c := g.Redis().Conn()
	defer c.Close()
	for i := 0; i < b.N; i++ {
		c.Send("MULTI")
		c.Send("SET", "foo", "bar")
		c.Send("GET", "foo")
		r, err := c.DoVar("EXEC")
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(gconv.Strings(r)) // prints [1, 1]
	}
}

/**
Do方法结合了Send、Flush和Receive方法的功能。
Do方法首先写入命令并刷新输出缓冲区。接下来，Do方法接收所有挂起的回复，包括Do刚刚发送的命令的回复。如果收到的任何回复都是错误，则Do返回错误。如果没有错误，则Do返回最后一个答复。如果Do方法的命令参数是“”，则Do方法将刷新输出缓冲区并接收挂起的应答，而不发送命令。使用Send和Do方法实现流水线事务。
*/

func TestDaNiPP(t *testing.T) {
	c := g.Redis().Conn()
	defer c.Close()
	c.Send("MULTI")
	c.Send("SET", "foo", "tonicgb")
	c.Send("GET", "foo")
	c.Send("ZRANGEBYSCORE", "zdd:zset", "0", 1621486509)
	r, err := c.DoVar("EXEC")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(gconv.Strings(r)) // prints [1, 1]
}
