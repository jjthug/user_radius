package ws_worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
)

const TaskLocationUpdate = "task:location_update"

func (manager *Manager) DistributeTaskLocationUpdate(ctx context.Context,
	payload *LocationUpdateEvent,
	opts ...asynq.Option,
) error {
	distributor := manager.taskDistributor
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}
	task := asynq.NewTask(TaskLocationUpdate, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %v", err)
	}

	log.Println("info=>", info)
	log.Println("payload=>", task.Payload())
	return nil
}

func (manager *Manager) ProcessTaskLocationUpdate(ctx context.Context, task *asynq.Task) error {
	var payload LocationUpdateEvent
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload %v", asynq.SkipRetry)
	}

	clientId := payload.From
	// send to all clients in the radius
	for c := range manager.clients {
		if c.nearbyClients[clientId] {
			c.egress <- task.Payload()
		}
	}

	return nil
}
