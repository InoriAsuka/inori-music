package storage

import (
	"context"
	"errors"
)

// CompositeProber tries multiple probe implementations until one supports the backend.
type CompositeProber struct {
	probers []Prober
}

func NewCompositeProber(probers ...Prober) *CompositeProber {
	return &CompositeProber{probers: probers}
}

func (prober *CompositeProber) Probe(ctx context.Context, backend StorageBackend) error {
	var unsupported error
	for _, candidate := range prober.probers {
		err := candidate.Probe(ctx, backend)
		if errors.Is(err, ErrProbeUnsupported) {
			unsupported = err
			continue
		}
		return err
	}
	if unsupported != nil {
		return unsupported
	}
	return ErrProbeUnsupported
}
