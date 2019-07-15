package key_generator

import (
	"github.com/bwmarrin/snowflake"
	"github.com/go-redis/redis"
	"github.com/zhuyst/shorturl-service/logger"
	"github.com/zhuyst/shorturl-service/node-id-generator"
)

// nodeMax 最大节点数
var nodeMax int64

func init() {
	// 2019-01-01 00:00:00
	snowflake.Epoch = 1546272000000

	// 3个机器位 = 8个节点
	snowflake.NodeBits = 3
	nodeMax = -1 ^ (-1 << snowflake.NodeBits)

	snowflake.StepBits = 1
}

// KeyGenerator 分布式ID生成器
type KeyGenerator struct {
	NodeId int64 // 当前节点ID

	redisClient     *redis.Client
	node            *snowflake.Node
	nodeIdGenerator *node_id_generator.NodeIdGenerator
}

// New 实例化一个KeyGenerator
func New(redisClient *redis.Client) (*KeyGenerator, error) {
	nodeIdGenerator := node_id_generator.New(redisClient, nodeMax)
	nodeId, err := nodeIdGenerator.GetNodeId()
	if err != nil {
		logger.Error("GetNodeId FAIL, Error: %s", err.Error())
		return nil, err
	}

	node, err := snowflake.NewNode(nodeId)
	if err != nil {
		logger.Error("Snowflake NewNode FAIL, Error: %s", err.Error())
		return nil, err
	}

	return &KeyGenerator{
		NodeId:          nodeId,
		redisClient:     redisClient,
		node:            node,
		nodeIdGenerator: nodeIdGenerator,
	}, nil
}

// Generate 生成一串分布式ID
func (generator *KeyGenerator) Generate() string {
	return generator.node.Generate().Base58()
}
