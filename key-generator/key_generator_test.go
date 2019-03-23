package key_generator

import (
	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis"
	"sync"
	"testing"
)

func TestNewKeyGenerator(t *testing.T) {
	generator, err := newKeyGenerator()
	if err != nil {
		t.Errorf("NewKeyGenerator ERROR: %s", err.Error())
		return
	}

	t.Logf("NewKeyGenerator PASS, NodeId: %d", generator.NodeId)
}

func TestKeyGenerator_Generate(t *testing.T) {
	generator, err := newKeyGenerator()
	if err != nil {
		t.Errorf("NewKeyGenerator ERROR: %s", err.Error())
		return
	}

	t.Logf("KeyGenerator_Generate PASS, key: %s", generator.Generate())
}

func TestMultiGenerate(t *testing.T) {
	generator, err := newKeyGenerator()
	if err != nil {
		t.Errorf("NewKeyGenerator ERROR: %s", err.Error())
		return
	}

	for i := 1; i <= 1000; i++ {
		t.Logf("MultiGenerate %d, ID: %s", i, generator.Generate())
	}
}

func TestMultiKeyGenerator(t *testing.T) {
	generatorNumber := int(nodeMax)
	var generators []*KeyGenerator
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(generatorNumber)

	redisClient := newTestRedisClient()
	redisClient.Set(nodeIdKey, generatorNumber-2, 0)
	generatorId := 0
	for i := 0; i < generatorNumber; i++ {
		go func() {
			defer waitGroup.Done()
			generatorId++
			j := generatorId

			generator, err := New(redisClient)
			if err != nil {
				t.Fatalf("NewKeyGenerator %d ERROR: %s", j, err.Error())
				return
			}
			t.Logf("MultiKeyGenerator %d, NodeId: %d", j, generator.NodeId)
			generators = append(generators, generator)
		}()
	}

	waitGroup.Wait()

	checkMap := make(map[int64]bool)
	for _, generator := range generators {
		if checkMap[generator.NodeId] {
			t.Error("MultiKeyGenerator ERROR, expected unique NodeId, got false")
			return
		}
		checkMap[generator.NodeId] = true
	}

	t.Log("MultiKeyGenerator PASS")
}

func TestMultiKeyGenerator_Generate(t *testing.T) {
	generatorNumber := int(nodeMax)
	generateNumber := 100
	var keys []string

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(generatorNumber * generateNumber)

	redisClient := newTestRedisClient()
	redisClient.Set(nodeIdKey, generatorNumber-2, 0)
	for i := 0; i < generatorNumber; i++ {
		go func() {
			generator, err := New(redisClient)
			if err != nil {
				t.Fatalf("NewKeyGenerator ERROR: %s", err.Error())
				return
			}

			for z := 0; z < generateNumber; z++ {
				go func() {
					defer waitGroup.Done()

					key := generator.Generate()
					t.Logf("KeyGenerator_Generate nodeId: %d, key: %s",
						generator.NodeId, key)

					keys = append(keys, key)
				}()
			}
		}()
	}

	waitGroup.Wait()

	checkMap := make(map[string]bool)
	for _, key := range keys {
		if checkMap[key] {
			t.Errorf("MultiKeyGenerator_Generate ERROR, expected unique id, got false")
			return
		}
		checkMap[key] = true
	}

	t.Log("MultiKeyGenerator_Generate PASS")
}

func newKeyGenerator() (*KeyGenerator, error) {
	redisClient := newTestRedisClient()
	return New(redisClient)
}

func newTestRedisClient() *redis.Client {
	ms, err := miniredis.Run()
	if err != nil {
		panic(err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: ms.Addr(),
	})
	if err = redisClient.Ping().Err(); err != nil {
		panic(err)
	}

	return redisClient
}