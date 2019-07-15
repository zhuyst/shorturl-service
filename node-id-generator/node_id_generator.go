package node_id_generator

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/satori/go.uuid"
	"github.com/zhuyst/redsync"
	"github.com/zhuyst/shorturl-service/logger"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	// nodeIdKeyPrefix 节点ID占用位
	nodeIdKeyPrefix = "SHORTURL_SERVICE:NODE_ID"

	// nodeIdLockKeyPrefix 节点ID占用分布式锁
	nodeIdLockKeyPrefix = "SHORTURL_SERVICE:NODE_ID_LOCK"

	// getNodeIdLockKey 获取节点ID统一分布式锁
	getNodeIdLockKey = "SHORTURL_SERVICE:GET_NODE_ID_LOCK"

	// holdKeyTime 一个节点每次能保持/续约的时间
	holdKeyTime = time.Second * 60
)

// NodeIdGenerator 节点ID生成器
type NodeIdGenerator struct {
	nodeId  int64 // 当前节点ID
	nodeMax int64 // 最大节点数量

	redisClient    *redis.Client
	redSync        *redsync.RedSync
	getNodeIdMutex *redsync.Mutex // 函数GetNodeId的分布式锁

	nodeIdKey     string         // 当前nodeId在Redis中的Key
	nodeIdLockKey string         // 对进行nodeIdKey进行读写的分布式锁Key
	nodeIdMutex   *redsync.Mutex // 使用nodeIdLockKey实例化的分布式锁

	nodeUUID   string       // 存储nodeId对应的value值
	nodeHolder *time.Ticker // 维持nodeId的定时器
}

// New 实例化一个NodeIdGenerator
func New(redisClient *redis.Client, nodeMax int64) *NodeIdGenerator {
	redSync := redsync.New(redisClient)

	return &NodeIdGenerator{
		nodeId:         -1,
		nodeMax:        nodeMax,
		redisClient:    redisClient,
		redSync:        redSync,
		getNodeIdMutex: redSync.NewMutex(getNodeIdLockKey),
	}
}

// GetNodeId 获取当前NodeId
func (generator *NodeIdGenerator) GetNodeId() (int64, error) {
	if err := generator.getNodeIdMutex.Lock(); err != nil {
		return -1, err
	}
	defer generator.getNodeIdMutex.Unlock()

	if generator.nodeId != -1 {
		return generator.nodeId, nil
	}

	return generator.generateNodeId()
}

// generateNodeId 生成NodeId，并初始化NodeHolder
func (generator *NodeIdGenerator) generateNodeId() (int64, error) {
	// 询问从0到nodeMax是否有坑位
	var i int64
	for i = 0; i < generator.nodeMax; i++ {
		key := fmt.Sprintf("%s:%d", nodeIdKeyPrefix, i)

		// 有Node占了坑位，跳过
		err := generator.redisClient.Get(key).Err()
		if err == nil {
			logger.Info("generateNodeId: NodeId %d exists, skip", i)
			continue
		}

		// 数据库错误直接返回
		if err != redis.Nil {
			return -1, err
		}

		// 找到坑位，将索引设为nodeId
		generator.nodeId = i
		generator.nodeIdKey = key

		// 维持该NodeId
		if err := generator.startNodeHolder(); err != nil {
			logger.Error("startNodeHolder FAIL, NodeId: %d, Error: %s", i, err.Error())
			return -1, err
		}

		// 监听应用退出信号，及时清除nodeId占用
		if err := generator.startListenSignal(); err != nil {
			logger.Error("startListenSignal FAIL, NodeId: %d, Error: %s", i, err.Error())
			return -1, err
		}

		logger.Info("generateNodeId: Get NodeId: %d", i)
		return generator.nodeId, nil
	}

	return -1, fmt.Errorf("nodeNumber reached the maximum: %d", generator.nodeMax)
}

// startNodeHolder 启动NodeHolder，维持当前持有的NodeId
func (generator *NodeIdGenerator) startNodeHolder() error {
	nodeUUID := uuid.NewV4().String()
	generator.nodeUUID = nodeUUID

	renewTime := holdKeyTime - 20*time.Second
	nodeHolder := time.NewTicker(renewTime)
	generator.nodeHolder = nodeHolder

	// 往redis设置nodeId
	if err := generator.setNodeId(); err != nil {
		return err
	}

	generator.nodeIdLockKey = fmt.Sprintf("%s:%d", nodeIdLockKeyPrefix, generator.nodeId)
	generator.nodeIdMutex = generator.redSync.NewMutex(generator.nodeIdLockKey)

	logger.Info("startNodeHolder, NodeId: %d, NodeUUID: %s", generator.nodeId, nodeUUID)

	// 启动新协程，维持nodeId
	go func() {
		for range nodeHolder.C {
			if err := generator.resetNodeId(); err != nil {
				logger.Error("NodeHolder ERROR, NodeId: %d, Error: %s", generator.nodeId, err.Error())
				panic(err)
			}
			logger.Info("NodeHolder resetNodeId, NodeId: %d", generator.nodeId)
		}
	}()

	return nil
}

// setNodeId 往Redis设置NodeId
func (generator *NodeIdGenerator) setNodeId() error {
	return generator.redisClient.Set(generator.nodeIdKey, generator.nodeUUID, holdKeyTime).Err()
}

// resetNodeId 使用当前NodeId刷新Redis中的NodeId，增加持有时长
func (generator *NodeIdGenerator) resetNodeId() error {
	if err := generator.nodeIdMutex.Lock(); err != nil {
		return err
	}
	defer generator.nodeIdMutex.Unlock()

	nodeUUIDFromRedis, err := generator.redisClient.Get(generator.nodeIdKey).Result()
	if err != nil {
		return err
	}

	// 通过value值检查防止错误持有NodeId（一种非常特殊的情况）
	if nodeUUIDFromRedis != generator.nodeUUID {
		return fmt.Errorf("nodeUUIDFromRedis: %s != generator.nodeUUID: %s",
			nodeUUIDFromRedis, generator.nodeUUID)
	}

	if err := generator.setNodeId(); err != nil {
		return err
	}

	return nil
}

// startListenSignal 监听应用退出信号，放弃持有当前NodeId并清除NodeId在Redis中的占用
func (generator *NodeIdGenerator) startListenSignal() error {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL,
		syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	// 清除存储在redis中的nodeId
	go func() {
		for range c {
			if err := generator.redisClient.Del(generator.nodeIdKey).Err(); err != nil {
				logger.Error("ClearNodeId FAIL, Error: %s", err.Error())
				panic(err)
				return
			}

			logger.Info("ClearNodeId SUCCESS, NodeId: %d", generator.nodeId)
			os.Exit(0)
		}
	}()

	return nil
}
