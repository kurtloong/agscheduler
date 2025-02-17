// go run examples/cluster/cluster_main.go

package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/kurtloong/agscheduler"
	"github.com/kurtloong/agscheduler/examples"
	"github.com/kurtloong/agscheduler/services"
	"github.com/kurtloong/agscheduler/stores"
)

var endpoint = flag.String("e", "127.0.0.1:36380", "Cluster Main Node endpoint")
var endpointHTTP = flag.String("eh", "127.0.0.1:36390", "Cluster Main Node endpoint HTTP")
var schedulerEndpoint = flag.String("se", "127.0.0.1:36360", "Cluster Main Node Scheduler endpoint")
var queue = flag.String("q", "default", "Cluster Main Node queue")

func main() {
	agscheduler.RegisterFuncs(examples.PrintMsg)

	flag.Parse()

	store := &stores.MemoryStore{}

	cn := &agscheduler.ClusterNode{
		MainEndpoint:      *endpoint,
		Endpoint:          *endpoint,
		EndpointHTTP:      *endpointHTTP,
		SchedulerEndpoint: *schedulerEndpoint,
		Queue:             *queue,
	}

	scheduler := &agscheduler.Scheduler{}
	err := scheduler.SetStore(store)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set store: %s", err))
		os.Exit(1)
	}
	err = scheduler.SetClusterNode(context.TODO(), cn)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set cluster node: %s", err))
		os.Exit(1)
	}

	cservice := &services.ClusterService{Cn: cn}
	err = cservice.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to start cluster service: %s", err))
		os.Exit(1)
	}

	select {}
}
