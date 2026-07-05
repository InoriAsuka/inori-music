package audioanalysis

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"inori-music/services/api/internal/catalog"
	"inori-music/services/api/internal/storage"
)

func TestAnalyzerSkipsOnUnsupportedBackend(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	locator := &stubLocator{
		objects: map[string]storage.MediaObject{
			"A": {ID: "A", BackendID: "b1", ObjectKey: "song.mp3"},
		},
	}
	fetcher := &stubFetcher{err: storage.ErrProbeUnsupported}
	updater := &stubUpdater{}

	a := &Analyzer{locator: locator, fetcher: fetcher, updater: updater, measure: failIfCalled(t)}
	if err := a.AnalyzeTrack(ctx, "track-1", "A"); err != nil {
		t.Fatalf("expected nil error on unsupported backend, got %v", err)
	}
	if len(updater.writes) != 0 {
		t.Fatal("expected no writeback on unsupported backend")
	}
	if !fetcher.called {
		t.Fatal("fetcher was not called — the analyzer should attempt the backend before detecting it's unsupported")
	}

}

func TestAnalyzerMeasuresAndWritesbackSuccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	locator := &stubLocator{
		objects: map[string]storage.MediaObject{
			"A": {ID: "A", BackendID: "b1", ObjectKey: "song.mp3"},
		},
	}
	fetcher := &stubFetcher{body: []byte("fake-audio-bytes")}
	updater := &stubUpdater{
		existing: map[string]catalog.Track{
			"track-1": {ID: "track-1"},
		},
	}

	const integratedLUFS = -14.0
	expectedGain := referenceLoudnessLUFS - integratedLUFS

	a := &Analyzer{locator: locator, fetcher: fetcher, updater: updater, measure: stubMeasure(integratedLUFS, nil)}
	if err := a.AnalyzeTrack(ctx, "track-1", "A"); err != nil {
		t.Fatalf("expected nil error on happy path, got %v", err)
	}
	if len(updater.writes) != 1 {
		t.Fatalf("expected exactly 1 writeback, got %d", len(updater.writes))
	}
	got := updater.writes[0].ReplayGainDb
	if got == nil {
		t.Fatal("expected non-nil ReplayGainDb")
	}
	if *got != expectedGain {
		t.Fatalf("ReplayGainDb = %v, want %v", *got, expectedGain)
	}
}

func TestAnalyzerNoWritebackOnMeasureFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	locator := &stubLocator{
		objects: map[string]storage.MediaObject{
			"A": {ID: "A", BackendID: "b1", ObjectKey: "song.mp3"},
		},
	}
	fetcher := &stubFetcher{body: []byte("audio")}
	updater := &stubUpdater{
		existing: map[string]catalog.Track{
			"track-1": {ID: "track-1"},
		},
	}

	measureErr := errors.New("ffmpeg: exit status 1")
	a := &Analyzer{locator: locator, fetcher: fetcher, updater: updater, measure: stubMeasure(0, measureErr)}
	if err := a.AnalyzeTrack(ctx, "track-1", "A"); err == nil {
		t.Fatal("expected error on measure failure, got nil")
	}
	if len(updater.writes) != 0 {
		t.Fatal("expected no writeback when measure fails")
	}
}

func TestParseLoudnormReportExtractsIntegrated(t *testing.T) {
	input := `[Parsed_loudnorm_0 @ 0x55a0c78] {
 "input_i" : "-14.50",
 "input_tp" : "-1.50",
 "input_lra" : "8.00",
 "input_thresh" : "-24.50",
 "output_i" : "-18.00",
 "target_offset" : "0.00"
}
trailing garbage`
	r, err := parseLoudnormReport(input)
	if err != nil {
		t.Fatalf("parseLoudnormReport: %v", err)
	}
	if r.InputIntegrated != "-14.50" {
		t.Fatalf("InputIntegrated = %q, want -14.50", r.InputIntegrated)
	}
}

func TestParseLoudnormReportNoJson(t *testing.T) {
	_, err := parseLoudnormReport("no json here at all")
	if err == nil {
		t.Fatal("expected error when no JSON present")
	}
}

// --- test helpers ---

var (
	_ MediaObjectLocator = (*stubLocator)(nil)
	_ ObjectFetcher       = (*stubFetcher)(nil)
	_ TrackUpdater        = (*stubUpdater)(nil)
)

type stubLocator struct {
	objects map[string]storage.MediaObject
	err     error
}

func (s *stubLocator) GetMediaObject(_ context.Context, id string) (storage.MediaObject, error) {
	if s.err != nil {
		return storage.MediaObject{}, s.err
	}
	o, ok := s.objects[id]
	if !ok {
		return storage.MediaObject{}, errors.New("not found")
	}
	return o, nil
}

type stubFetcher struct {
	body   []byte
	err    error
	called bool
}

func (s *stubFetcher) GetObject(_ context.Context, _, _ string) (io.ReadCloser, error) {
	s.called = true
	if s.err != nil {
		return nil, s.err
	}
	return io.NopCloser(strings.NewReader(string(s.body))), nil
}

type stubUpdater struct {
	existing map[string]catalog.Track
	writes   []catalog.UpdateTrackRequest
}

func (s *stubUpdater) UpdateTrack(_ context.Context, id string, req catalog.UpdateTrackRequest) (catalog.Track, error) {
	s.writes = append(s.writes, req)
	if t, ok := s.existing[id]; ok {
		return t, nil
	}
	return catalog.Track{}, errors.New("not found")
}

func stubMeasure(loudness float64, err error) func(context.Context, string) (float64, error) {
	return func(_ context.Context, _ string) (float64, error) {
		return loudness, err
	}
}

func failIfCalled(t *testing.T) func(context.Context, string) (float64, error) {
	t.Helper()
	return func(_ context.Context, _ string) (float64, error) {
		t.Fatal("measure stub should not be called")
		return 0, nil
	}
}
