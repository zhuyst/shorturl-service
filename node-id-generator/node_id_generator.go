package node_id_generator

import (
	"github.com/go-redis/redis"
	"github.com/zhuyst/redsync"
)

const (
	NodeIdKey     = "SHORTURL_SERVICE:NODE_ID"
	nodeIdLockKey = "SHORTURL_SERVICE:NODE_ID_LOCK"
)

type NodeIdGenerator struct {
	nodeMax     int64
	redisClient *redis.Client
	mutex       *redsync.Mutex
}

func New(redisClient *redis.Client, nodeMax int64) *NodeIdGenerator {
	redSync := redsync.New(redisClient)

	return &NodeIdGenerator{
		nodeMax:     nodeMax,
		redisClient: redisClient,
		mutex:       redSync.NewMutex(nodeIdLockKey),
	}
}

func (generator *NodeIdGenerator) Generate() (int64, error) {
	err := generator.mutex.Lock()
	if err != nil {
		return -1, err
	}
	redisClient := generator.redisClient

	nodeId, err := redisClient.Incr(NodeIdKey).Result()
	if err != nil {
		return -1, err
	}

	// 小于等于nodeMax可以直接返回
	if nodeId <= generator.nodeMax {
		generator.mutex.Unlock()
		return nodeId, nil
	}

	// 处理超过nodeMax的情况
	nodeId = 0
	if err = redisClient.Set(NodeIdKey, nodeId, 0).Err(); err != nil {
		return -1, err
	}

	generator.mutex.Unlock()
	return nodeId, nil
}
