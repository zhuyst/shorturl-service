package key_generator

import (
	"github.com/bwmarrin/snowflake"
	"github.com/go-redis/redis"
	"github.com/zhuyst/redsync"
)

const (
	nodeIdKey     = "SHORTURL_SERVICE:NODE_ID"
	nodeIdLockKey = "SHORTURL_SERVICE:NODE_ID_LOCK"
)

var nodeMax int64

func init() {
	// 2019-01-01 00:00:00
	snowflake.Epoch = 1546272000000

	// 3个机器位 = 8个节点
	snowflake.NodeBits = 3
	nodeMax = -1 ^ (-1 << snowflake.NodeBits)

	snowflake.StepBits = 1
}

type KeyGenerator struct {
	NodeId int64

	redisClient *redis.Client
	node        *snowflake.Node
	mutex       *redsync.Mutex
}

func New(redisClient *redis.Client) (*KeyGenerator, error) {
	redSync := redsync.New(redisClient)
	generator := &KeyGenerator{
		redisClient: redisClient,
		mutex:       redSync.NewMutex(nodeIdLockKey),
	}

	nodeId, err := generator.generateNodeId()
	if err != nil {
		return nil, err
	}
	generator.NodeId = nodeId

	node, err := snowflake.NewNode(nodeId)
	if err != nil {
		return nil, err
	}
	generator.node = node

	return generator, nil
}

func (generator *KeyGenerator) Generate() string {
	return generator.node.Generate().Base58()
}

func (generator *KeyGenerator) generateNodeId() (int64, error) {
	err := generator.mutex.Lock()
	if err != nil {
		return -1, err
	}
	redisClient := generator.redisClient

	nodeId, err := redisClient.Incr(nodeIdKey).Result()
	if err != nil {
		return -1, err
	}

	// 小于等于nodeMax可以直接返回
	if nodeId <= nodeMax {
		generator.mutex.Unlock()
		return nodeId, nil
	}

	// 处理超过nodeMax的情况
	nodeId = 0
	if err = redisClient.Set(nodeIdKey, nodeId, 0).Err(); err != nil {
		return -1, err
	}

	generator.mutex.Unlock()
	return nodeId, nil
}
