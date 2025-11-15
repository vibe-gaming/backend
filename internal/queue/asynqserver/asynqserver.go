package asynqserver

import (
	"github.com/hibiken/asynq"
	"github.com/vibe-gaming/backend/internal/cache"
	"github.com/vibe-gaming/backend/internal/config"
	"github.com/vibe-gaming/backend/internal/queue/processor"
	"github.com/vibe-gaming/backend/internal/queue/task"
	"github.com/vibe-gaming/backend/internal/worker"
)

func New(cfg config.Cache, workers *worker.Workers) (*asynq.Server, *asynq.ServeMux) {
	mux, queues := getQueues(workers)
	srv := asynq.NewServer(
		RedisOptions(cfg),
		asynq.Config{
			Concurrency: 10,
			LogLevel:    asynq.ErrorLevel,
			Queues:      queues,
		},
	)

	return srv, mux
}

func RedisOptions(cfg config.Cache) asynq.RedisConnOpt {
	var opts asynq.RedisConnOpt
	if cfg.Type == cache.RedisTypeCluster {
		opts = asynq.RedisClusterClientOpt{Addrs: cfg.RedisCluster.Addresses}
	} else {
		opts = asynq.RedisClientOpt{Addr: cfg.Redis.Address}
	}
	return opts
}

func getQueues(workers *worker.Workers) (*asynq.ServeMux, map[string]int) {
	mux := asynq.NewServeMux()
	mux.Handle(task.SendEmailTaskName, processor.NewSendEmailProcessor(workers))
	queues := map[string]int{
		task.SendEmailQueueName: 1,
	}
	return mux, queues
}
