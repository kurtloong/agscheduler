// go run examples/stores/base.go examples/stores/memory.go

package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/kurtloong/agscheduler"
	"github.com/kurtloong/agscheduler/stores"
)

func main() {
	store := &stores.MemoryStore{}

	scheduler := &agscheduler.Scheduler{}
	err := scheduler.SetStore(store)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set store: %s", err))
		os.Exit(1)
	}

	runExample(scheduler)
}
