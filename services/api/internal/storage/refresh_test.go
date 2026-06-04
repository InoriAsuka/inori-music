package storage

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type fakeProber struct {
	calls atomic.Int32
	err   error
}

func (prober *fakeProber) Probe(_ context.Context, _ StorageBackend) error {
	prober.calls.Add(1)
	return prober.err
}

type fakeCapacityProvider struct {
	calls atomic.Int32
	err   error
}

func (provider *fakeCapacityProvider) Capacity(_ context.Context, backend StorageBackend) (CapacityReport, error) {
	provider.calls.Add(1)
	return CapacityReport{BackendID: backend.ID, TotalBytes: 100, AvailableBytes: 40, UsedBytes: 60}, provider.err
}

func TestRefreshEnabledBackendsSkipsDisabledAndIsolatesErrors(t *testing.T) {
	ctx := context.Background()
	repository := NewMemoryRepository()
	service := NewService(repository)
	service.prober = &fakeProber{err: errors.New("probe boom")}
	service.capacityProvider = &fakeCapacityProvider{}
	registerRefreshBackend(t, service, "enabled", true, t.TempDir())
	registerRefreshBackend(t, service, "disabled", false, t.TempDir())

	report, err := service.RefreshEnabledBackends(ctx)
	if err != nil {
		t.Fatalf("RefreshEnabledBackends() error = %v", err)
	}
	if len(report.Results) != 2 {
		t.Fatalf("result count = %d, want 2", len(report.Results))
	}
	if report.Results[0].BackendID != "disabled" || !report.Results[0].Skipped {
		t.Fatalf("disabled result = %+v, want skipped disabled backend", report.Results[0])
	}
	if report.Results[1].BackendID != "enabled" || len(report.Results[1].Errors) != 1 || report.Results[1].Capacity == nil {
		t.Fatalf("enabled result = %+v, want isolated probe error and capacity", report.Results[1])
	}
}

func TestRefreshSchedulerStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	repository := NewMemoryRepository()
	service := NewService(repository)
	prober := &fakeProber{}
	service.prober = prober
	service.capacityProvider = &fakeCapacityProvider{}
	registerRefreshBackend(t, service, "enabled", true, t.TempDir())
	reports := make(chan RefreshReport, 4)
	scheduler := NewRefreshScheduler(service, 5*time.Millisecond, func(report RefreshReport, err error) {
		if err != nil {
			t.Errorf("scheduler refresh error = %v", err)
		}
		reports <- report
	})
	done := make(chan struct{})
	go func() {
		scheduler.Run(ctx)
		close(done)
	}()

	select {
	case <-reports:
	case <-time.After(time.Second):
		t.Fatal("scheduler did not refresh before timeout")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("scheduler did not stop after cancellation")
	}
	calls := prober.calls.Load()
	time.Sleep(20 * time.Millisecond)
	if prober.calls.Load() != calls {
		t.Fatalf("probe calls after cancellation = %d, want %d", prober.calls.Load(), calls)
	}
}

func registerRefreshBackend(t *testing.T, service *Service, id string, enabled bool, root string) {
	t.Helper()
	_, err := service.RegisterBackend(context.Background(), StorageBackend{
		ID: id, Type: BackendTypeLocal, DisplayName: id, Enabled: enabled,
		Config: BackendConfig{Local: &LocalConfig{RootPath: root}},
	})
	if err != nil {
		t.Fatalf("RegisterBackend(%s) error = %v", id, err)
	}
}
