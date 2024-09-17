package queue

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

type Queue struct {
	client *redis.Client
	ctx    context.Context
}

// Init initializes the Redis client and connects to Redis.
func Init(connStr string, ctx context.Context) *Queue {
	opt, err := redis.ParseURL(connStr)
	if err != nil {
		log.Fatalf("failed to parse Redis URL: %v", err)
	}

	client := redis.NewClient(opt)

	err = client.Ping(ctx).Err()
	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}

	fmt.Println("Connected to Redis")

	return &Queue{
		client: client,
		ctx:    ctx,
	}
}

// Add pushes a string to the head of a Redis list.
func (q *Queue) Add(listName string, data string) error {
	_, err := q.client.LPush(q.ctx, listName, data).Result()
	if err != nil {
		fmt.Println("Error adding entry to list:", err)
		return err
	}

	fmt.Println("String added to list")

	return nil
}
