package storage

import (
	"context"
	"errors"
	"sync"
	"time"
)

// RefreshResult captures an isolated backend refresh outcome.
type RefreshResult struct {
	BackendID string          `json:"backendId"`
	Skipped   bool            `json:"skipped"`
	Probe     *ProbeResult    `json:"probe,omitempty"`
	Capacity  *CapacityReport `json:"capacity,omitempty"`
	Errors    []string        `json:"errors,omitempty"`
}

// RefreshReport captures one batch refresh run.
type RefreshReport struct {
	StartedAt   time.Time       `json:"startedAt"`
	CompletedAt time.Time       `json:"completedAt"`
	Results     []RefreshResult `json:"results"`
}

// RefreshEnabledBackends probes and reads capacity for all configured backends without short-circuiting failures.
func (service *Service) RefreshEnabledBackends(ctx context.Context) (RefreshReport, error) {
	backends, err := service.repository.List(ctx)
	if err != nil {
		return RefreshReport{}, err
	}
	report := RefreshReport{StartedAt: service.now().UTC(), Results: make([]RefreshResult, 0, len(backends))}
	for _, backend := range backends {
		result := RefreshResult{BackendID: backend.ID}
		if !backend.Enabled {
			result.Skipped = true
			report.Results = append(report.Results, result)
			continue
		}

		probe, probeErr := service.ProbeBackend(ctx, backend.ID)
		result.Probe = &probe
		if probeErr != nil {
			result.Errors = append(result.Errors, probeErr.Error())
		}
		capacity, capacityErr := service.GetBackendCapacity(ctx, backend.ID)
		if capacityErr == nil {
			result.Capacity = &capacity
		} else if !errors.Is(capacityErr, ErrCapacityUnsupported) {
			result.Errors = append(result.Errors, capacityErr.Error())
		}
		report.Results = append(report.Results, result)
	}
	report.CompletedAt = service.now().UTC()
	return report, nil
}

// RefreshScheduler runs periodic backend refresh until its context is canceled.
type RefreshScheduler struct {
	service  *Service
	interval time.Duration
	onReport func(RefreshReport, error)
	once     sync.Once
}

func NewRefreshScheduler(service *Service, interval time.Duration, onReport func(RefreshReport, error)) *RefreshScheduler {
	return &RefreshScheduler{service: service, interval: interval, onReport: onReport}
}

func (scheduler *RefreshScheduler) Run(ctx context.Context) {
	scheduler.once.Do(func() {
		if scheduler.interval <= 0 {
			return
		}
		ticker := time.NewTicker(scheduler.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				report, err := scheduler.service.RefreshEnabledBackends(ctx)
				if scheduler.onReport != nil {
					scheduler.onReport(report, err)
				}
			}
		}
	})
}
