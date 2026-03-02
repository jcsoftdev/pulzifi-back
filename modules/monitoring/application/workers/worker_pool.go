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

// FailCheckFunc is called to mark a check as "error" when the worker cannot
// recover from a failure (panic or returned error).
type FailCheckFunc func(ctx context.Context, checkID uuid.UUID, schemaName string, errMsg string)

type SnapshotJob struct {
	CheckID    uuid.UUID
	URL        string
	SchemaName string
}

type WorkerPool struct {
	jobQueue     chan SnapshotJob
	snapshotPort SnapshotPort
	failCheck    FailCheckFunc
	wg           sync.WaitGroup
	quit         chan struct{}
}

func NewWorkerPool(snapshotPort SnapshotPort, bufferSize int, failCheck FailCheckFunc) *WorkerPool {
	return &WorkerPool{
		jobQueue:     make(chan SnapshotJob, bufferSize),
		snapshotPort: snapshotPort,
		failCheck:    failCheck,
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
			p.executeJob(id, job)
		case <-p.quit:
			return
		}
	}
}

func (p *WorkerPool) executeJob(workerID int, job SnapshotJob) {
	logger.Debug("Worker received job", zap.Int("worker_id", workerID), zap.String("check_id", job.CheckID.String()))

	// Recover from panics so the worker goroutine stays alive and the
	// check is marked as failed instead of staying "pending" forever.
	defer func() {
		if r := recover(); r != nil {
			errMsg := fmt.Sprintf("worker panic: %v", r)
			logger.Error("Worker panicked during check execution",
				zap.Int("worker_id", workerID),
				zap.String("check_id", job.CheckID.String()),
				zap.String("error", errMsg))
			if p.failCheck != nil {
				p.failCheck(context.Background(), job.CheckID, job.SchemaName, errMsg)
			}
		}
	}()

	if err := p.snapshotPort.ExecuteCheck(context.Background(), job.CheckID, job.URL, job.SchemaName); err != nil {
		logger.Error("Worker failed to execute check",
			zap.Int("worker_id", workerID),
			zap.String("check_id", job.CheckID.String()),
			zap.Error(err))
		// ExecuteCheck already marks the check as "error" internally for most
		// paths. This catch handles early failures (e.g. check not found in DB).
		if p.failCheck != nil {
			p.failCheck(context.Background(), job.CheckID, job.SchemaName, err.Error())
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
