// go run examples/http/http_server.go

package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/kurtloong/agscheduler"
	"github.com/kurtloong/agscheduler/examples"
	"github.com/kurtloong/agscheduler/services"
	"github.com/kurtloong/agscheduler/stores"
)

func main() {
	agscheduler.RegisterFuncs(examples.PrintMsg)

	store := &stores.MemoryStore{}

	scheduler := &agscheduler.Scheduler{}
	err := scheduler.SetStore(store)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set store: %s", err))
		os.Exit(1)
	}

	shservice := services.SchedulerHTTPService{
		Scheduler: scheduler,
		Address:   "127.0.0.1:36370",
	}
	err = shservice.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to start service: %s", err))
		os.Exit(1)
	}

	select {}
}
