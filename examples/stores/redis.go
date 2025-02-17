// go run examples/stores/base.go examples/stores/redis.go

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"

	"github.com/kurtloong/agscheduler"
	"github.com/kurtloong/agscheduler/stores"
)

func main() {
	url := "redis://127.0.0.1:6379/0"
	opt, err := redis.ParseURL(url)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to parse url: %s", err))
		os.Exit(1)
	}
	rdb := redis.NewClient(opt)
	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to database: %s", err))
		os.Exit(1)
	}
	defer rdb.Close()
	store := &stores.RedisStore{
		RDB:         rdb,
		JobsKey:     "agscheduler.example_jobs",
		RunTimesKey: "agscheduler.example_run_times",
	}

	scheduler := &agscheduler.Scheduler{}
	err = scheduler.SetStore(store)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set store: %s", err))
		os.Exit(1)
	}

	runExample(scheduler)
}
