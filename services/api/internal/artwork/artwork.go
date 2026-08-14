// Package artwork resolves album cover art directly from audio files on a
// filesystem-backed storage backend (local / NFS / SMB), without requiring
// anything to have been extracted or registered at import time.
//
// This exists because the catalog import workflow (catalog.Service.ImportTrack)
// is metadata-only — it never opens the audio file — so Album.ArtworkMediaObjectID
// is never populated for the vast majority of libraries. Re-importing or
// re-scanning the library to backfill it is off the table (see requirement.md
// v5.39.0), so cover art is instead resolved on demand, the same way track
// streaming already serves bytes for backends that cannot issue presigned
// URLs. Callers are expected to cache the result (see httpapi's artworkCache)
// since resolution means opening a file and parsing its tags.
package artwork

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
)

// Image is a resolved cover image: raw bytes plus the MIME type they should
// be served with.
type Image struct {
	MIMEType string
	Data     []byte
}

// siblingBaseNames lists the sibling-file basenames to look for, in priority
// order. "cover" and "folder" are the two most common conventions in ripped
// libraries (Windows Explorer/Media Player popularized "folder.jpg"; most
// other tooling uses "cover.jpg"); "front" and "album" are less common but
// cheap to also support.
var siblingBaseNames = []string{"cover", "folder", "front", "album"}

// siblingExtensions lists the image extensions to look for, in priority
// order, alongside the MIME type each maps to.
var siblingExtensions = []struct {
	ext      string
	mimeType string
}{
	{".jpg", "image/jpeg"},
	{".jpeg", "image/jpeg"},
	{".png", "image/png"},
	{".webp", "image/webp"},
}

// ExtractEmbedded opens the audio file at path and returns its embedded
// cover picture, if any. Supports every format github.com/dhowden/tag
// supports: FLAC, OGG/Vorbis, MP4, ID3v2 (MP3), ID3v1, DSF.
//
// A file that cannot be opened returns a non-nil error (a real I/O problem
// worth surfacing). A file that opens fine but isn't a recognizable/tagged
// audio format, or has no embedded picture, returns (Image{}, false, nil) —
// that is the expected, common case for most tracks and callers should fall
// back to FindSibling rather than treat it as a failure.
func ExtractEmbedded(path string) (Image, bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return Image{}, false, err
	}
	defer f.Close()

	meta, err := tag.ReadFrom(f)
	if err != nil {
		// Unrecognized format or unparsable tags — not an error the caller
		// needs to act on, just "nothing embedded here".
		return Image{}, false, nil
	}
	pic := meta.Picture()
	if pic == nil || len(pic.Data) == 0 {
		return Image{}, false, nil
	}
	mimeType := strings.TrimSpace(pic.MIMEType)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return Image{MIMEType: mimeType, Data: pic.Data}, true, nil
}

// FindSibling looks in dir for a cover/folder/front/album image file
// (case-insensitive basename, .jpg/.jpeg/.png/.webp extension, checked in
// that priority order) and returns its contents. Sibling image files are
// extremely common in ripped libraries, independent of whatever embedded
// artwork (if any) the audio files themselves carry.
//
// Returns (Image{}, false, nil) when dir has no matching file. Returns a
// non-nil error only for a real I/O problem (e.g. dir does not exist).
func FindSibling(dir string) (Image, bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return Image{}, false, err
	}
	// Single directory read, case-insensitive lookup — cheaper than statting
	// every (basename, extension) combination individually.
	byLower := make(map[string]string, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		byLower[strings.ToLower(e.Name())] = e.Name()
	}
	for _, base := range siblingBaseNames {
		for _, candidate := range siblingExtensions {
			real, ok := byLower[base+candidate.ext]
			if !ok {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, real))
			if err != nil {
				return Image{}, false, err
			}
			return Image{MIMEType: candidate.mimeType, Data: data}, true, nil
		}
	}
	return Image{}, false, nil
}
