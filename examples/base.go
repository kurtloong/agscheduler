package examples

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kurtloong/agscheduler"
)

func PrintMsg(ctx context.Context, j agscheduler.Job) {
	slog.Info(fmt.Sprintf("Run job `%s` %s\n\n", j.FullName(), j.Args))
}
