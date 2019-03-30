package node_id_generator

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/satori/go.uuid"
	"github.com/zhuyst/redsync"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	nodeIdKeyPrefix = "SHORTURL_SERVICE:NODE_ID"
	nodeIdLockKey   = "SHORTURL_SERVICE:NODE_ID_LOCK"

	holdKeyTime = time.Second * 60
)

type NodeIdGenerator struct {
	nodeId int64

	nodeMax     int64
	redisClient *redis.Client
	mutex       *redsync.Mutex

	nodeIdKey  string
	nodeUUID   string
	nodeHolder *time.Ticker
}

func New(redisClient *redis.Client, nodeMax int64) *NodeIdGenerator {
	redSync := redsync.New(redisClient)

	return &NodeIdGenerator{
		nodeId:      -1,
		nodeMax:     nodeMax,
		redisClient: redisClient,
		mutex:       redSync.NewMutex(nodeIdLockKey),
	}
}

func (generator *NodeIdGenerator) GetNodeId() (int64, error) {
	if err := generator.mutex.Lock(); err != nil {
		return -1, err
	}
	defer generator.mutex.Unlock()

	if generator.nodeId != -1 {
		return generator.nodeId, nil
	}

	return generator.generateNodeId()
}

func (generator *NodeIdGenerator) generateNodeId() (int64, error) {
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
		generator.nodeId = i
		generator.nodeIdKey = key
		if err := generator.startNodeHolder(); err != nil {
			return -1, err
		}

		if err := generator.startListenSignal(); err != nil {
			return -1, err
		}

		return generator.nodeId, nil
	}

	return -1, fmt.Errorf("nodeNumber reached the maximum: %d", generator.nodeMax)
}

func (generator *NodeIdGenerator) startNodeHolder() error {
	nodeIdKey := generator.nodeIdKey
	if nodeIdKey == "" {
		return errors.New("need nodeIdKey to startNodeHolder")
	}

	nodeUUID := uuid.NewV4().String()
	generator.nodeUUID = nodeUUID

	redisClient := generator.redisClient
	setFunc := func() error {
		return redisClient.Set(nodeIdKey, nodeUUID, holdKeyTime).Err()
	}

	renewTime := holdKeyTime - 20*time.Second
	nodeHolder := time.NewTicker(renewTime)
	generator.nodeHolder = nodeHolder

	if err := setFunc(); err != nil {
		return err
	}

	go func() {
		for range nodeHolder.C {
			nodeUUIDFromRedis, err := redisClient.Get(nodeIdKey).Result()
			if err != nil {
				panic(err)
				return
			}

			if nodeUUIDFromRedis != generator.nodeUUID {
				err := fmt.Errorf("nodeUUIDFromRedis: %s != generator.nodeUUID: %s",
					nodeUUIDFromRedis, generator.nodeUUID)
				panic(err)
				return
			}

			if err := setFunc(); err != nil {
				panic(err)
			}
		}
	}()

	return nil
}

func (generator *NodeIdGenerator) startListenSignal() error {
	nodeIdKey := generator.nodeIdKey
	if nodeIdKey == "" {
		return errors.New("need nodeIdKey to startListenSignal")
	}

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL,
		syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		for range c {
			if err := generator.redisClient.Del(generator.nodeIdKey).Err(); err != nil {
				panic(err)
				return
			}
			os.Exit(0)
		}
	}()

	return nil
}
