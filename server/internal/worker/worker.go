package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

type Job interface {
	Name() string
	Run(ctx context.Context) error
}

type ScheduledJob struct {
	Job      Job
	Schedule string // cron expression
}

type Worker struct {
	cron   *cron.Cron
	jobs   []ScheduledJob
	logger *slog.Logger
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func New(logger *slog.Logger) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Worker{
		cron:   cron.New(cron.WithSeconds()),
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (w *Worker) Register(schedule string, job Job) {
	w.jobs = append(w.jobs, ScheduledJob{Job: job, Schedule: schedule})
}

func (w *Worker) Start() error {
	for _, sj := range w.jobs {
		j := sj
		_, err := w.cron.AddFunc(j.Schedule, func() {
			w.wg.Add(1)
			defer w.wg.Done()

			w.logger.Info("job started", slog.String("job", j.Job.Name()))

			if err := j.Job.Run(w.ctx); err != nil {
				w.logger.Error("job failed",
					slog.String("job", j.Job.Name()),
					slog.String("error", err.Error()),
				)
				return
			}

			w.logger.Info("job completed", slog.String("job", j.Job.Name()))
		})
		if err != nil {
			return err
		}
		w.logger.Info("job registered",
			slog.String("job", j.Job.Name()),
			slog.String("schedule", j.Schedule),
		)
	}

	w.cron.Start()
	w.logger.Info("worker started", slog.Int("jobs", len(w.jobs)))
	return nil
}

func (w *Worker) Stop() {
	w.logger.Info("worker stopping...")
	w.cancel()
	w.cron.Stop()

	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		w.logger.Info("worker stopped")
	case <-time.After(30 * time.Second):
		w.logger.Warn("worker stop timed out after 30s, some jobs may still be running")
	}
}
