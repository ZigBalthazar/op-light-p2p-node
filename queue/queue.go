package queue

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Queue represents a queue backed by a Redis stream.
type Queue struct {
	client *redis.Client
	ctx    context.Context
}

// Init initializes a new Queue with a Redis connection.
func Init(connStr string, ctx context.Context) *Queue {
	// Parse the Redis connection URL
	opt, err := redis.ParseURL(connStr)
	if err != nil {
		panic(fmt.Sprintf("failed to parse Redis URL: %v", err))
	}

	// Create a new Redis client
	client := redis.NewClient(opt)

	// Ping the Redis server to check the connection
	err = client.Ping(ctx).Err()
	if err != nil {
		panic(fmt.Sprintf("failed to connect to Redis: %v", err))
	}

	fmt.Println("Connected to Redis")

	return &Queue{
		client: client,
		ctx:    ctx,
	}
}

// Add adds a new entry to the specified Redis stream.
func (q *Queue) Add(streamName string, data map[string]interface{}) error {

	// Convert data to map[string]string
	stringData := make(map[string]string)
	for key, value := range data{
		stringData[key] = fmt.Sprintf("%v", value)
	}

	// Add the data to the Redis stream
	_, err := q.client.XAdd(q.ctx, &redis.XAddArgs{
		Stream: streamName,
		Values: stringData,
	}).Result()
	if err != nil {
		fmt.Println("Error adding entry to stream:", err)
		return err
	}

	fmt.Println("New block added to stream")

	return nil
}
