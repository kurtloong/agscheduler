package stores

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kurtloong/agscheduler"
)

func dryRunStores(ctx context.Context, j agscheduler.Job) {}

func testAGScheduler(t *testing.T, s *agscheduler.Scheduler) {
	agscheduler.RegisterFuncs(dryRunStores)

	s.Start()

	j := agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.TYPE_INTERVAL,
		Interval: "1s",
		Func:     dryRunStores,
		Args:     map[string]any{"arg1": "1", "arg2": "2", "arg3": "3"},
	}
	assert.Empty(t, j.FuncName)
	assert.Empty(t, j.Status)

	j, err := s.AddJob(j)
	assert.NoError(t, err)
	assert.Equal(t, agscheduler.STATUS_RUNNING, j.Status)
	assert.NotEmpty(t, j.FuncName)

	j.Type = agscheduler.TYPE_CRON
	j.CronExpr = "*/1 * * * *"
	j, err = s.UpdateJob(j)
	assert.NoError(t, err)
	assert.Equal(t, agscheduler.TYPE_CRON, j.Type)

	timezone, err := time.LoadLocation(j.Timezone)
	assert.NoError(t, err)
	nextRunTimeMax, err := time.ParseInLocation(time.DateTime, "9999-09-09 09:09:09", timezone)
	assert.NoError(t, err)

	j, err = s.PauseJob(j.Id)
	assert.NoError(t, err)
	assert.Equal(t, agscheduler.STATUS_PAUSED, j.Status)
	assert.Equal(t, nextRunTimeMax.Unix(), j.NextRunTime.Unix())

	j, err = s.ResumeJob(j.Id)
	assert.NoError(t, err)
	assert.NotEqual(t, nextRunTimeMax.Unix(), j.NextRunTime.Unix())

	err = s.RunJob(j)
	assert.NoError(t, err)

	err = s.DeleteJob(j.Id)
	assert.NoError(t, err)
	_, err = s.GetJob(j.Id)
	assert.ErrorIs(t, err, agscheduler.JobNotFoundError(j.Id))

	err = s.DeleteAllJobs()
	assert.NoError(t, err)
	js, err := s.GetAllJobs()
	assert.NoError(t, err)
	assert.Len(t, js, 0)

	s.Stop()
}
