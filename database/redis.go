package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/powershitxyz/SolanaProbe/config"
	redis "github.com/redis/go-redis/v9"
)

var rdb *redis.Client
var ctx = context.Background()

func init() {
	var conf = config.GetConfig()
	if conf.AllStart > 0 {
		rdb = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf(conf.Redis.Host+":%d", conf.Redis.Port),
			Password: conf.Redis.Password,
			DB:       conf.Redis.Db,
		})
	}
}

func GetRedis() *redis.Client {
	return rdb
}

func PublishTokenSearch(tokenSearch []byte) {
	go func() {
		cmd := rdb.Publish(ctx, "tokensearch", tokenSearch)
		val, err := cmd.Result()
		if err != nil {
			log.Println("publish token search error", val, err)
			return
		}

	}()

}
func PublishS(channel string, msg []byte) {
	go func() {
		cmd := rdb.Publish(ctx, channel, msg)
		val, err := cmd.Result()
		if err != nil {
			log.Println("publish token search error", val, err)
			return
		}
		//log.Printf("message published to %d subscribers\n", val)
	}()
}
func GetCacheListByIndex(key string, index int64) (string, error) {
	return rdb.LIndex(context.Background(), key, int64(index)).Result()
}

// 将数据左插入到列表，并设置过期时间
func SetCacheObjectListByLeft(key string, dataList interface{}, size int, expireByDays int) (int64, error) {
	// 先修剪列表
	SetTrim(key, size)

	// 左侧添加数据
	count, err := rdb.LPush(ctx, key, dataList).Result()
	if err != nil {
		return 0, err
	}

	// 设置过期时间，如果 expireByDays 大于 0
	if expireByDays > 0 {
		_, err = rdb.Expire(ctx, key, time.Duration(expireByDays)*24*time.Hour).Result()
		if err != nil {
			return count, err
		}
	}

	return count, nil
}

// 修剪列表到指定大小
func SetTrim(key string, size int) error {
	listSize, err := rdb.LLen(ctx, key).Result()
	if err != nil {
		return err
	}

	if listSize > int64(size) {
		_, err := rdb.LTrim(ctx, key, 0, int64(size)-1).Result()
		if err != nil {
			return err
		}
	}
	return nil
}
func SetCacheObjectListByIndex(key string, index int64, value interface{}) {
	err := rdb.LSet(ctx, key, index, value).Err()
	if err != nil {
		fmt.Println("Error setting value:", err)
		return
	}
}

// 从 Redis 获取哈希表字段并反序列化为结构体
func HGet(hashKey string, field string, result interface{}) error {
	// 获取哈希表字段值
	data, err := rdb.HGet(ctx, hashKey, field).Result()
	if err != nil {
		if err == redis.Nil {
			// 如果字段不存在，返回 nil 而不是错误
			return nil
		}
		return fmt.Errorf("error getting hash field: %v", err)
	}

	// 将 JSON 数据反序列化为结构体实例
	err = json.Unmarshal([]byte(data), result)
	if err != nil {
		return fmt.Errorf("error unmarshaling JSON to struct: %v", err)
	}

	return nil
}

// 将结构体序列化为 JSON 并写入 Redis 哈希表字段
func HSet(hashKey string, field string, data interface{}) error {
	// 将结构体序列化为 JSON 字符串
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling struct to JSON: %v", err)
	}

	// 将 JSON 数据写入 Redis 哈希表字段
	err = rdb.HSet(ctx, hashKey, field, jsonData).Err()
	if err != nil {
		return fmt.Errorf("error setting hash field: %v", err)
	}

	return nil
}
func RedisExpire(key string, expireTime time.Duration) {
	err := rdb.Expire(ctx, key, expireTime).Err()
	if err != nil {
		log.Printf("RedisExpire,key:%s, err:%v ", key, err)
		return
	}
}
