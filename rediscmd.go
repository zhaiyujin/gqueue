/**
 * package: cache
 * name:  cache
 * author: zhaiyujin
 * date: 2021/5/18 10:08
 * description:
 */
package gqueue

import "time"

type RedisCmd interface {
	RPush(string, string) error
	LPop(string) (string, error)
	BLPop(time time.Duration, keys ...string) ([]string, error)
	ZAdd(key string, z ...ZAdd) error
	ZRemRangeByScore(key, min, max string) (int, error)
	ZRangeScores(key string, start, stop int64) ([]string, error)
	Send(func(r *Redis)) ([]string, error)
}
