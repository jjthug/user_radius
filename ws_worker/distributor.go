package ws_worker

import (
	"github.com/hibiken/asynq"
)

type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(residOpt asynq.RedisClientOpt) *RedisTaskDistributor {
	client := asynq.NewClient(residOpt)
	return &RedisTaskDistributor{client}
}
