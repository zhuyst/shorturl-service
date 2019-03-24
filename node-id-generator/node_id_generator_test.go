package node_id_generator

import (
	"fmt"
	"github.com/go-redis/redis"
	"shorturl_service/helper"
	"sync"
	"sync/atomic"
	"testing"
)

func TestNodeIdGenerator_Generate(t *testing.T) {
	var nodeMax int64 = 8
	redisClient := helper.NewTestRedisClient()

	nodeId, err := testGenerate(redisClient, nodeMax)
	if err != nil {
		t.Error(err.Error())
		return
	}

	t.Logf("NodeIdGenerator_Generate PASS, nodeId: %d", nodeId)
}

func TestNodeIdGenerator_MultiGenerate(t *testing.T) {
	var nodeMax int64 = 8
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

			nodeId, err := testGenerate(redisClient, nodeMax)
			if err != nil {
				t.Error(err.Error())
				return
			}

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

func testGenerate(redisClient *redis.Client, nodeMax int64) (int64, error) {
	generator := New(redisClient, nodeMax)
	nodeId, err := generator.Generate()
	if err != nil {
		return -1, fmt.Errorf("NodeIdGenerator_Generate ERROR: %s", err.Error())
	}

	if nodeId < 0 || nodeId > nodeMax {
		return -1, fmt.Errorf("NodeIdGenerator_Generate ERROR, "+
			"expected 0 <= nodeId <= %d, got %d", nodeMax, nodeId)
	}

	return nodeId, nil
}
