package key_generator

import (
	"github.com/bwmarrin/snowflake"
	"github.com/go-redis/redis"
	"shorturl_service/node-id-generator"
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

	redisClient     *redis.Client
	node            *snowflake.Node
	nodeIdGenerator *node_id_generator.NodeIdGenerator
}

func New(redisClient *redis.Client) (*KeyGenerator, error) {
	nodeIdGenerator := node_id_generator.New(redisClient, nodeMax)
	nodeId, err := nodeIdGenerator.GetNodeId()
	if err != nil {
		return nil, err
	}

	node, err := snowflake.NewNode(nodeId)
	if err != nil {
		return nil, err
	}

	return &KeyGenerator{
		NodeId:          nodeId,
		redisClient:     redisClient,
		node:            node,
		nodeIdGenerator: nodeIdGenerator,
	}, nil
}

func (generator *KeyGenerator) Generate() string {
	return generator.node.Generate().Base58()
}
