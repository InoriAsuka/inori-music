// Package audioanalysis computes ReplayGain loudness values for newly added
// tracks by shelling out to ffmpeg's loudnorm filter against the stored audio
// file fetched from a storage backend.
package audioanalysis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"inori-music/services/api/internal/catalog"
	"inori-music/services/api/internal/storage"
)

// referenceLoudnessLUFS is the ReplayGain 2.0 target loudness. The computed
// gain is the offset needed to bring a track's measured integrated loudness
// up to this reference: positive means "turn up", negative means "turn down".
const referenceLoudnessLUFS = -18.0

// MediaObjectLocator resolves a media object ID to its storage location.
// Satisfied by *storage.MediaObjectService.
type MediaObjectLocator interface {
	GetMediaObject(ctx context.Context, id string) (storage.MediaObject, error)
}

// ObjectFetcher fetches the raw bytes of a stored object. Satisfied by *storage.Service.
type ObjectFetcher interface {
	GetObject(ctx context.Context, backendID, objectKey string) (io.ReadCloser, error)
}

// TrackUpdater persists the analyzed ReplayGain value. Satisfied by *catalog.Service.
type TrackUpdater interface {
	UpdateTrack(ctx context.Context, id string, req catalog.UpdateTrackRequest) (catalog.Track, error)
}

// Analyzer computes and persists ReplayGain loudness values for tracks.
// Instances are always non-nil (constructed via New); calling AnalyzeTrack on
// a nil *Analyzer would panic.
type Analyzer struct {
	locator MediaObjectLocator
	fetcher ObjectFetcher
	updater TrackUpdater
	measure func(ctx context.Context, path string) (float64, error)
}

// New builds an Analyzer that resolves media objects via mediaObjects, fetches
// object bytes via storageSvc, and writes analysis results back via catalogSvc.
func New(mediaObjects *storage.MediaObjectService, storageSvc *storage.Service, catalogSvc *catalog.Service) *Analyzer {
	return &Analyzer{
		locator: mediaObjects,
		fetcher: storageSvc,
		updater: catalogSvc,
		measure: measureLoudnessFFmpeg,
	}
}

// AnalyzeTrack fetches the track's underlying audio, measures its integrated
// loudness via ffmpeg, and writes the resulting ReplayGain value back to the
// track record. It is designed to run as a fire-and-forget background task:
// backends that don't support direct reads and a missing ffmpeg binary are
// treated as graceful no-ops (logged, nil error); any other failure is
// returned so the caller can log it. In no case does analysis block or fail
// the triggering import/create call.
func (a *Analyzer) AnalyzeTrack(ctx context.Context, trackID, mediaObjectID string) error {
	obj, err := a.locator.GetMediaObject(ctx, mediaObjectID)
	if err != nil {
		return fmt.Errorf("get media object: %w", err)
	}
	rc, err := a.fetcher.GetObject(ctx, obj.BackendID, obj.ObjectKey)
	if err != nil {
		if errors.Is(err, storage.ErrProbeUnsupported) {
			log.Printf("audioanalysis: backend %s does not support direct reads, skipping track %s", obj.BackendID, trackID)
			return nil
		}
		return fmt.Errorf("fetch object: %w", err)
	}
	defer rc.Close()

	tmp, err := os.CreateTemp("", "audioanalysis-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	_, copyErr := io.Copy(tmp, rc)
	closeErr := tmp.Close()
	if copyErr != nil {
		return fmt.Errorf("write temp file: %w", copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close temp file: %w", closeErr)
	}

	loudness, err := a.measure(ctx, tmpPath)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			log.Printf("audioanalysis: ffmpeg not found, skipping track %s", trackID)
			return nil
		}
		return fmt.Errorf("measure loudness: %w", err)
	}

	gain := referenceLoudnessLUFS - loudness
	if _, err := a.updater.UpdateTrack(ctx, trackID, catalog.UpdateTrackRequest{ReplayGainDb: &gain}); err != nil {
		return fmt.Errorf("update track: %w", err)
	}
	return nil
}

// loudnormReport is the subset of ffmpeg's loudnorm JSON report used to derive
// a ReplayGain value.
type loudnormReport struct {
	InputIntegrated string `json:"input_i"`
}

// measureLoudnessFFmpeg runs ffmpeg's loudnorm filter in analysis-only mode and
// parses the integrated loudness (LUFS) from its stderr JSON report.
func measureLoudnessFFmpeg(ctx context.Context, path string) (float64, error) {
	cmd := exec.CommandContext(ctx, "ffmpeg", "-hide_banner", "-nostats", "-i", path, "-af", "loudnorm=print_format=json", "-f", "null", "-")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return 0, err
	}
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	var buf strings.Builder
	if _, err := io.Copy(&buf, stderr); err != nil {
		_ = cmd.Wait()
		return 0, fmt.Errorf("read ffmpeg output: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return 0, fmt.Errorf("ffmpeg: %w", err)
	}
	report, err := parseLoudnormReport(buf.String())
	if err != nil {
		return 0, err
	}
	loudness, err := strconv.ParseFloat(report.InputIntegrated, 64)
	if err != nil {
		return 0, fmt.Errorf("parse input_i %q: %w", report.InputIntegrated, err)
	}
	return loudness, nil
}

// parseLoudnormReport extracts the trailing JSON object emitted by ffmpeg's
// loudnorm filter from its combined stderr output.
func parseLoudnormReport(output string) (loudnormReport, error) {
	start := strings.LastIndex(output, "{")
	end := strings.LastIndex(output, "}")
	if start == -1 || end == -1 || end < start {
		return loudnormReport{}, fmt.Errorf("no loudnorm json report found in ffmpeg output")
	}
	var report loudnormReport
	if err := json.Unmarshal([]byte(output[start:end+1]), &report); err != nil {
		return loudnormReport{}, fmt.Errorf("decode loudnorm report: %w", err)
	}
	return report, nil
}
