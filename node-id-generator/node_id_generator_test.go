package node_id_generator

import (
	"github.com/go-redis/redis"
	"github.com/zhuyst/shorturl-service/helper"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const nodeMax int64 = 8

func TestNodeIdGenerator_GetNodeId(t *testing.T) {
	redisClient := helper.NewTestRedisClient()
	nodeId := testGenerate(t, redisClient, nodeMax)
	t.Logf("NodeIdGenerator_GetNodeId PASS, nodeId: %d", nodeId)
}

func TestNodeIdGenerator_MultiGenerate(t *testing.T) {
	redisClient := helper.NewTestRedisClient()

	var nodeIds []int64
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(int(nodeMax))

	var generatorId int64 = 0
	var i int64
	for i = 0; i < nodeMax; i++ {
		go func() {
			defer waitGroup.Done()

			j := atomic.LoadInt64(&generatorId)
			atomic.AddInt64(&generatorId, 1)

			nodeId := testGenerate(t, redisClient, nodeMax)
			nodeIds = append(nodeIds, nodeId)
			t.Logf("NodeIdGenerator_MultiGenerate %d: %d", j, nodeId)
		}()
	}

	waitGroup.Wait()
	checkMap := make(map[int64]bool)
	for _, nodeId := range nodeIds {
		if checkMap[nodeId] {
			t.Errorf("NodeIdGenerator_MultiGenerate ERROR, expected unique nodeId, got false")
			return
		}
		checkMap[nodeId] = true
	}

	t.Logf("NodeIdGenerator_MultiGenerate PASS")
}

func TestNodeIdGenerator_NodeHolder(t *testing.T) {
	redisClient := helper.NewTestRedisClient()

	generator := New(redisClient, nodeMax)
	nodeId, err := generator.GetNodeId()
	if err != nil {
		t.Errorf("NodeIdGenerator_GetNodeId ERROR: %s", err.Error())
		return
	}

	time.Sleep(holdKeyTime)

	if nodeId != generator.nodeId {
		t.Errorf("NodeIdGenerator_NodeHolder ERROR, "+
			"expected generator.nodeId == %d, got %d", nodeId, generator.nodeId)
		return
	}
	t.Logf("NodeIdGenerator_NodeHolder PASS")
}

func testGenerate(t *testing.T, redisClient *redis.Client, nodeMax int64) int64 {
	generator := New(redisClient, nodeMax)
	nodeId, err := generator.GetNodeId()
	if err != nil {
		t.Fatalf("NodeIdGenerator_GetNodeId ERROR: %s", err.Error())
		return -1
	}

	if nodeId < 0 || nodeId > nodeMax {
		t.Fatalf("NodeIdGenerator_GetNodeId ERROR, "+
			"expected 0 <= nodeId <= %d, got %d", nodeMax, nodeId)
		return -1
	}

	return nodeId
}
