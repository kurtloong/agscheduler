// go run examples/stores/base.go examples/stores/redis.go

package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"

	"github.com/kwkwc/agscheduler"
	"github.com/kwkwc/agscheduler/stores"
)

func main() {
	url := "redis://127.0.0.1:6379/0"
	opt, err := redis.ParseURL(url)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to parse url: %s", err))
		os.Exit(1)
	}
	rdb := redis.NewClient(opt)
	store := &stores.RedisStore{
		RDB:         rdb,
		JobsKey:     "agscheduler.example_jobs",
		RunTimesKey: "agscheduler.example_run_times",
	}

	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)

	runExample(scheduler)
}
