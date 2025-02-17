package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/kurtloong/agscheduler"
	pb "github.com/kurtloong/agscheduler/services/proto"
	"github.com/kurtloong/agscheduler/stores"
)

var ctx = context.Background()

func dryRunRPC(ctx context.Context, j agscheduler.Job) {}

func testAGSchedulerRPC(t *testing.T, c pb.SchedulerClient) {
	_, err := c.Start(ctx, &emptypb.Empty{})
	assert.NoError(t, err)

	j := agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.TYPE_INTERVAL,
		Interval: "1s",
		FuncName: "github.com/kurtloong/agscheduler/services.dryRunRPC",
		Args:     map[string]any{"arg1": "1", "arg2": "2", "arg3": "3"},
	}
	assert.Empty(t, j.Status)

	pbJ, err := c.AddJob(ctx, agscheduler.JobToPbJobPtr(j))
	assert.NoError(t, err)
	j = agscheduler.PbJobPtrToJob(pbJ)
	assert.Equal(t, agscheduler.STATUS_RUNNING, j.Status)

	j.Type = agscheduler.TYPE_CRON
	j.CronExpr = "*/1 * * * *"
	pbJ, err = c.UpdateJob(ctx, agscheduler.JobToPbJobPtr(j))
	assert.NoError(t, err)
	j = agscheduler.PbJobPtrToJob(pbJ)
	assert.Equal(t, agscheduler.TYPE_CRON, j.Type)

	timezone, err := time.LoadLocation(j.Timezone)
	assert.NoError(t, err)
	nextRunTimeMax, err := time.ParseInLocation(time.DateTime, "9999-09-09 09:09:09", timezone)
	assert.NoError(t, err)

	pbJ, err = c.PauseJob(ctx, &pb.JobId{Id: j.Id})
	assert.NoError(t, err)
	j = agscheduler.PbJobPtrToJob(pbJ)
	assert.Equal(t, agscheduler.STATUS_PAUSED, j.Status)
	assert.Equal(t, nextRunTimeMax.Unix(), j.NextRunTime.Unix())

	pbJ, err = c.ResumeJob(ctx, &pb.JobId{Id: j.Id})
	assert.NoError(t, err)
	j = agscheduler.PbJobPtrToJob(pbJ)
	assert.NotEqual(t, nextRunTimeMax.Unix(), j.NextRunTime.Unix())

	_, err = c.RunJob(ctx, pbJ)
	assert.NoError(t, err)

	_, err = c.DeleteJob(ctx, &pb.JobId{Id: j.Id})
	assert.NoError(t, err)
	_, err = c.GetJob(ctx, &pb.JobId{Id: j.Id})
	assert.Contains(t, err.Error(), agscheduler.JobNotFoundError(j.Id).Error())

	_, err = c.DeleteAllJobs(ctx, &emptypb.Empty{})
	assert.NoError(t, err)
	pbJs, err := c.GetAllJobs(ctx, &emptypb.Empty{})
	assert.NoError(t, err)
	js := agscheduler.PbJobsPtrToJobs(pbJs)
	assert.Len(t, js, 0)

	_, err = c.Stop(ctx, &emptypb.Empty{})
	assert.NoError(t, err)
}

func TestRPCService(t *testing.T) {
	agscheduler.RegisterFuncs(dryRunRPC)

	store := &stores.MemoryStore{}

	scheduler := &agscheduler.Scheduler{}
	err := scheduler.SetStore(store)
	assert.NoError(t, err)

	srservice := SchedulerRPCService{
		Scheduler: scheduler,
		// Address:   "127.0.0.1:36360",
	}
	srservice.Start()

	conn, err := grpc.Dial(srservice.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()
	client := pb.NewSchedulerClient(conn)

	testAGSchedulerRPC(t, client)

	err = store.Clear()
	assert.NoError(t, err)
}
