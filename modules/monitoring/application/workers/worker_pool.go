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

func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()
	for {
		select {
		case job := <-p.jobQueue:
			logger.Debug("Worker received job", zap.Int("worker_id", id), zap.String("check_id", job.CheckID.String()))
			// Use a fresh context or the background context since the job creation context might be cancelled
			// Ideally we could pass a context in the job but it's tricky with channels
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

func (p *WorkerPool) Dispatch(ctx context.Context, checkID uuid.UUID, url string, schemaName string) error {
	job := SnapshotJob{
		CheckID:    checkID,
		URL:        url,
		SchemaName: schemaName,
	}

	select {
	case p.jobQueue <- job:
		return nil
	case <-time.After(5 * time.Second):
		logger.Error("Job queue full, dropping job after 5s timeout",
			zap.String("check_id", checkID.String()),
			zap.String("url", url),
			zap.String("schema", schemaName))
		return fmt.Errorf("worker pool backpressure: job queue full, dropped job for check %s", checkID)
	case <-ctx.Done():
		return ctx.Err()
	}
}
