/**
 * package: gqueue
 * name:  redis_test
 * author: zhaiyujin
 * date: 2021/5/19 15:42
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

func TestZAdd(t *testing.T) {

	tm := time.Now().Add(time.Second * 30)

	err := NewRedis().ZAdd("zdd", ZAdd{
		Score:  float64(tm.Unix()),
		Member: string("fuckyoubaby"),
	})
	if err != nil {
		fmt.Println(err.Error())
	}
}

//按照指定区间有序排序
func TestZRangeWithScores(t *testing.T) {

	res, err := NewRedis().ZRangeScores("zdd:zset", 0, time.Now().Unix())
	if err != nil {
		fmt.Println(err.Error())
	}
	g.Dump(res)
	/*
		c.redis.ZRemRangeByScore(topicName, "0", strconv.FormatInt(currentTime, 10))*/
}

//删除指定区间有序
func TestZRemRangeByScore(t *testing.T) {
	NewRedis().ZRemRangeByScore("zdd:zset", "0", gconv.String(time.Now().Unix()))
}

func TestBlpop(t *testing.T) {
	var revBody string
	res, err := NewRedis().BLPop(time.Second, "queue:list")
	if err != nil {
		fmt.Println(err.Error())
	}
	if len(res) >= 2 {
		revBody = res[1]
	}
	g.Dump(revBody)

}
