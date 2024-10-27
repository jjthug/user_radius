package ws_worker

import (
	"github.com/hibiken/asynq"
)

type RedisTaskProcessor struct {
	server *asynq.Server
}

func NewRedisTaskProcessor(redisOpts asynq.RedisClientOpt) *RedisTaskProcessor {
	server := asynq.NewServer(redisOpts, asynq.Config{})

	return &RedisTaskProcessor{
		server: server,
	}
}

func (m *Manager) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskLocationUpdate, m.ProcessTaskLocationUpdate)

	return m.taskProcessor.server.Start(mux)
}
