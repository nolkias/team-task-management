package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"teamtask/internal/domain"
)

type TaskListResult struct {
	Tasks []domain.Task
	Total int
}

type TaskListCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewTaskListCache(client *redis.Client, ttl time.Duration) *TaskListCache {
	return &TaskListCache{client: client, ttl: ttl}
}

func BuildTaskListKey(f domain.TaskFilter, p domain.PageRequest) string {
	status := "-"
	if f.Status != "" {
		status = string(f.Status)
	}
	assignee := "-"
	if f.AssigneeID != nil {
		assignee = fmt.Sprintf("%d", *f.AssigneeID)
	}
	return fmt.Sprintf("task_list:%d:%s:%s:%d:%d", f.TeamID, status, assignee, p.Page, p.PageSize)
}

func teamKeysSetKey(teamID int64) string {
	return fmt.Sprintf("task_list_keys:%d", teamID)
}

func (c *TaskListCache) Get(ctx context.Context, key string) (*TaskListResult, error) {
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var result TaskListResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *TaskListCache) Set(ctx context.Context, teamID int64, key string, result *TaskListResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	if err := c.client.Set(ctx, key, data, c.ttl).Err(); err != nil {
		return err
	}
	return c.client.SAdd(ctx, teamKeysSetKey(teamID), key).Err()
}

func (c *TaskListCache) InvalidateTeam(ctx context.Context, teamID int64) error {
	setKey := teamKeysSetKey(teamID)
	keys, err := c.client.SMembers(ctx, setKey).Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return err
		}
	}
	return c.client.Del(ctx, setKey).Err()
}
