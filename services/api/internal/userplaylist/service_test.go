package userplaylist_test

import (
	"context"
	"testing"
	"time"

	"inori-music/services/api/internal/userplaylist"
)

func fixedClock(t time.Time) func() time.Time { return func() time.Time { return t } }

func TestCreatePlaylist(t *testing.T) {
	svc := userplaylist.NewService(userplaylist.NewMemoryRepository())
	p, err := svc.CreatePlaylist(context.Background(), "user1", userplaylist.CreateRequest{Name: "My Mix"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "My Mix" || p.UserID != "user1" || p.ID == "" {
		t.Fatalf("unexpected playlist: %+v", p)
	}
	if len(p.TrackIDs) != 0 {
		t.Fatalf("expected empty track list, got %v", p.TrackIDs)
	}
}

func TestCreatePlaylist_EmptyName(t *testing.T) {
	svc := userplaylist.NewService(userplaylist.NewMemoryRepository())
	_, err := svc.CreatePlaylist(context.Background(), "user1", userplaylist.CreateRequest{Name: "  "})
	if err != userplaylist.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestListPlaylists(t *testing.T) {
	svc := userplaylist.NewService(userplaylist.NewMemoryRepository())
	now := time.Now()
	svc.WithClock(fixedClock(now))
	svc.CreatePlaylist(context.Background(), "user1", userplaylist.CreateRequest{Name: "A"}) //nolint:errcheck
	svc.WithClock(fixedClock(now.Add(time.Second)))
	svc.CreatePlaylist(context.Background(), "user1", userplaylist.CreateRequest{Name: "B"}) //nolint:errcheck
	svc.CreatePlaylist(context.Background(), "user2", userplaylist.CreateRequest{Name: "C"}) //nolint:errcheck

	list, err := svc.ListPlaylists(context.Background(), "user1")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 playlists, got %d", len(list))
	}
	// Newest first
	if list[0].Name != "B" {
		t.Fatalf("expected B first, got %s", list[0].Name)
	}
}

func TestGetPlaylist_ForbiddenWhenWrongUser(t *testing.T) {
	svc := userplaylist.NewService(userplaylist.NewMemoryRepository())
	p, _ := svc.CreatePlaylist(context.Background(), "user1", userplaylist.CreateRequest{Name: "Private"})
	_, err := svc.GetPlaylist(context.Background(), "user2", p.ID)
	if err != userplaylist.ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestUpdatePlaylist(t *testing.T) {
	svc := userplaylist.NewService(userplaylist.NewMemoryRepository())
	p, _ := svc.CreatePlaylist(context.Background(), "user1", userplaylist.CreateRequest{Name: "Old"})
	newName := "New"
	updated, err := svc.UpdatePlaylist(context.Background(), "user1", p.ID, userplaylist.UpdateRequest{Name: &newName})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "New" {
		t.Fatalf("expected New, got %s", updated.Name)
	}
}

func TestDeletePlaylist(t *testing.T) {
	svc := userplaylist.NewService(userplaylist.NewMemoryRepository())
	p, _ := svc.CreatePlaylist(context.Background(), "user1", userplaylist.CreateRequest{Name: "Bye"})
	if err := svc.DeletePlaylist(context.Background(), "user1", p.ID); err != nil {
		t.Fatal(err)
	}
	list, _ := svc.ListPlaylists(context.Background(), "user1")
	if len(list) != 0 {
		t.Fatalf("expected empty list after delete, got %d", len(list))
	}
}

func TestAddTrack(t *testing.T) {
	svc := userplaylist.NewService(userplaylist.NewMemoryRepository())
	p, _ := svc.CreatePlaylist(context.Background(), "user1", userplaylist.CreateRequest{Name: "Tracks"})
	if err := svc.AddTrack(context.Background(), "user1", p.ID, "track-1"); err != nil {
		t.Fatal(err)
	}
	// Idempotent
	if err := svc.AddTrack(context.Background(), "user1", p.ID, "track-1"); err != nil {
		t.Fatal(err)
	}
	got, _ := svc.GetPlaylist(context.Background(), "user1", p.ID)
	if len(got.TrackIDs) != 1 || got.TrackIDs[0] != "track-1" {
		t.Fatalf("unexpected tracks: %v", got.TrackIDs)
	}
}

func TestRemoveTrack(t *testing.T) {
	svc := userplaylist.NewService(userplaylist.NewMemoryRepository())
	p, _ := svc.CreatePlaylist(context.Background(), "user1", userplaylist.CreateRequest{Name: "Tracks"})
	svc.AddTrack(context.Background(), "user1", p.ID, "track-1") //nolint:errcheck
	svc.AddTrack(context.Background(), "user1", p.ID, "track-2") //nolint:errcheck
	if err := svc.RemoveTrack(context.Background(), "user1", p.ID, "track-1"); err != nil {
		t.Fatal(err)
	}
	got, _ := svc.GetPlaylist(context.Background(), "user1", p.ID)
	if len(got.TrackIDs) != 1 || got.TrackIDs[0] != "track-2" {
		t.Fatalf("unexpected tracks after remove: %v", got.TrackIDs)
	}
}

func TestSetTracks(t *testing.T) {
	svc := userplaylist.NewService(userplaylist.NewMemoryRepository())
	p, _ := svc.CreatePlaylist(context.Background(), "user1", userplaylist.CreateRequest{Name: "Set"})
	svc.AddTrack(context.Background(), "user1", p.ID, "old-track") //nolint:errcheck
	if err := svc.SetTracks(context.Background(), "user1", p.ID, []string{"t1", "t2", "t3"}); err != nil {
		t.Fatal(err)
	}
	got, _ := svc.GetPlaylist(context.Background(), "user1", p.ID)
	if len(got.TrackIDs) != 3 || got.TrackIDs[0] != "t1" || got.TrackIDs[2] != "t3" {
		t.Fatalf("unexpected tracks after set: %v", got.TrackIDs)
	}
}
