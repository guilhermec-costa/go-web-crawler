package infra

import (
	"guilhermec-costa/go-web-crawler/crawler/validation"
	"guilhermec-costa/go-web-crawler/server/types"
	"log/slog"
)

type JobMonitorQueue struct {
	Q   chan types.Job
	len int
}

type JobMonitor struct {
	queue     JobMonitorQueue
	processor types.JobProcessor
}

func (m *JobMonitor) TriggerJob(userId string, params validation.CrawlerParams) error {
	job := types.Job{
		Params: params,
		UserId: userId,
	}

	select {
	case m.queue.Q <- job:
		slog.Info("extraction job queued")
		return nil

	default:
		slog.Error("extraction queue is full")
	}

	return nil
}

func NewJobMonitor(p types.JobProcessor, queueLen int) *JobMonitor {
	m := &JobMonitor{
		queue: JobMonitorQueue{
			Q:   make(chan types.Job, queueLen),
			len: queueLen,
		},
		processor: p,
	}
	m.Start()
	return m
}

func (j *JobMonitor) Start() {
	slog.Info("starting job queue monitor", "buffer_size", j.queue.len)
	go func() {
		for job := range j.queue.Q {
			if err := j.processor(job); err != nil {
				slog.Error("failed to complete job processing", "err", err, "params", job.Params)
			} else {
				slog.Info("finished processing job successfully", "params", job.Params)
			}
		}
	}()
}
