package workers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type SnapshotPort interface {
	ExecuteCheck(ctx context.Context, checkID uuid.UUID, url string, schemaName string) error
}

type SnapshotJob struct {
	CheckID    uuid.UUID
	URL        string
	SchemaName string
}

type WorkerPool struct {
	jobQueue     chan SnapshotJob
	snapshotPort SnapshotPort
	wg           sync.WaitGroup
	quit         chan struct{}
}

func NewWorkerPool(snapshotPort SnapshotPort, bufferSize int) *WorkerPool {
	return &WorkerPool{
		jobQueue:     make(chan SnapshotJob, bufferSize),
		snapshotPort: snapshotPort,
		quit:         make(chan struct{}),
	}
}

func (p *WorkerPool) Start(concurrency int) {
	for i := 0; i < concurrency; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
	logger.Info("WorkerPool started", zap.Int("concurrency", concurrency))
}

func (p *WorkerPool) Stop() {
	close(p.quit)
	p.wg.Wait()
	logger.Info("WorkerPool stopped")
}

// QueueLength returns the current number of jobs waiting in the queue.
func (p *WorkerPool) QueueLength() int {
	return len(p.jobQueue)
}

func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()
	for {
		select {
		case job := <-p.jobQueue:
			logger.Debug("Worker received job", zap.Int("worker_id", id), zap.String("check_id", job.CheckID.String()))
			if err := p.snapshotPort.ExecuteCheck(context.Background(), job.CheckID, job.URL, job.SchemaName); err != nil {
				logger.Error("Worker failed to execute check",
					zap.Int("worker_id", id),
					zap.String("check_id", job.CheckID.String()),
					zap.Error(err))
			}
		case <-p.quit:
			return
		}
	}
}

// Dispatch enqueues a job with retry and exponential backoff.
// It attempts up to 3 times with delays of 500ms, 1s, and 2s before giving up.
func (p *WorkerPool) Dispatch(ctx context.Context, checkID uuid.UUID, url string, schemaName string) error {
	job := SnapshotJob{
		CheckID:    checkID,
		URL:        url,
		SchemaName: schemaName,
	}

	backoffs := []time.Duration{500 * time.Millisecond, 1 * time.Second, 2 * time.Second}

	for attempt := 0; attempt <= len(backoffs); attempt++ {
		select {
		case p.jobQueue <- job:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Queue full, will retry
		}

		if attempt == len(backoffs) {
			break
		}

		// Wait with backoff before retrying
		timer := time.NewTimer(backoffs[attempt])
		select {
		case p.jobQueue <- job:
			timer.Stop()
			return nil
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
			// Continue to next attempt
		}
	}

	logger.Error("Job queue full after retries, dropping job",
		zap.String("check_id", checkID.String()),
		zap.String("url", url),
		zap.String("schema", schemaName))
	return fmt.Errorf("worker pool backpressure: job queue full after %d retries, dropped job for check %s", len(backoffs), checkID)
}
