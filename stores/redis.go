package stores

import (
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/kwkwc/agscheduler"
)

const (
	jobs_key      = "agscheduler.jobs"
	run_times_key = "agscheduler.run_times"
)

type RedisStore struct {
	RDB *redis.Client
}

func (s *RedisStore) Init() error {
	return nil
}

func (s *RedisStore) AddJob(j agscheduler.Job) error {
	state, err := agscheduler.StateDump(j)
	if err != nil {
		return err
	}

	_, err = s.RDB.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(ctx, jobs_key, j.Id, state)
		pipe.ZAdd(ctx, run_times_key, redis.Z{Score: float64(j.NextRunTime.UTC().Unix()), Member: j.Id})
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *RedisStore) GetJob(id string) (agscheduler.Job, error) {
	state, err := s.RDB.HGet(ctx, jobs_key, id).Bytes()
	if err == redis.Nil {
		return agscheduler.Job{}, agscheduler.JobNotFoundError(id)
	}
	if err != nil {
		return agscheduler.Job{}, err
	}

	return agscheduler.StateLoad(state)
}

func (s *RedisStore) GetAllJobs() ([]agscheduler.Job, error) {
	mapStates, err := s.RDB.HGetAll(ctx, jobs_key).Result()
	if err != nil {
		return nil, err
	}

	var jobList []agscheduler.Job
	for _, v := range mapStates {
		j, err := agscheduler.StateLoad([]byte(v))
		if err != nil {
			return nil, err
		}
		jobList = append(jobList, j)
	}

	return jobList, nil
}

func (s *RedisStore) UpdateJob(j agscheduler.Job) error {
	state, err := agscheduler.StateDump(j)
	if err != nil {
		return err
	}

	_, err = s.RDB.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(ctx, jobs_key, j.Id, state)
		pipe.ZAdd(ctx, run_times_key, redis.Z{Score: float64(j.NextRunTime.UTC().Unix()), Member: j.Id})
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *RedisStore) DeleteJob(id string) error {
	_, err := s.RDB.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HDel(ctx, jobs_key, id)
		pipe.ZRem(ctx, run_times_key, id)
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *RedisStore) DeleteAllJobs() error {
	_, err := s.RDB.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Del(ctx, jobs_key)
		pipe.Del(ctx, run_times_key)
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *RedisStore) GetNextRunTime() (time.Time, error) {
	sliceRunTimes, err := s.RDB.ZRangeWithScores(ctx, run_times_key, 0, 0).Result()
	if err != nil || len(sliceRunTimes) == 0 {
		return time.Time{}, nil
	}

	nextRunTimeMin := time.Unix(int64(sliceRunTimes[0].Score), 0).UTC()
	return nextRunTimeMin, nil
}

func (s *RedisStore) Clear() error {
	return s.DeleteAllJobs()
}
