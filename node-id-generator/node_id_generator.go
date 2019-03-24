package node_id_generator

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/zhuyst/redsync"
	"time"
)

const (
	nodeIdKeyPrefix = "SHORTURL_SERVICE:NODE_ID"
	nodeIdLockKey   = "SHORTURL_SERVICE:NODE_ID_LOCK"

	holdKeyTime = time.Minute * 5
)

type NodeIdGenerator struct {
	NodeId int64

	nodeMax     int64
	redisClient *redis.Client
	mutex       *redsync.Mutex

	nodeIdKey  string
	nodeHolder *time.Ticker
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
	if err := generator.mutex.Lock(); err != nil {
		return -1, err
	}
	defer generator.mutex.Unlock()

	// 询问从0到nodeMax是否有坑位
	var i int64
	for i = 0; i < generator.nodeMax; i++ {
		key := fmt.Sprintf("%s:%d", nodeIdKeyPrefix, i)

		// 有Node占了坑位，跳过
		err := generator.redisClient.Get(key).Err()
		if err == nil {
			continue
		}

		// 数据库错误直接返回
		if err != redis.Nil {
			return -1, err
		}

		// 开始生成NodeId
		generator.NodeId = i
		generator.nodeIdKey = key
		if err := generator.startNodeHolder(); err != nil {
			return -1, err
		}
		return generator.NodeId, nil
	}

	return -1, fmt.Errorf("nodeNumber reached the maximum: %d", generator.nodeMax)
}

func (generator *NodeIdGenerator) startNodeHolder() error {
	if generator.nodeIdKey == "" {
		return errors.New("need nodeIdKey to startNodeHolder")
	}

	setFunc := func() error {
		return generator.redisClient.Set(generator.nodeIdKey, true, holdKeyTime).Err()
	}

	renewTime := holdKeyTime - 30*time.Second
	nodeHolder := time.NewTicker(renewTime)
	generator.nodeHolder = nodeHolder

	if err := setFunc(); err != nil {
		return err
	}

	go func() {
		for range nodeHolder.C {
			if err := setFunc(); err != nil {
				panic(err)
			}
		}
	}()

	return nil
}
