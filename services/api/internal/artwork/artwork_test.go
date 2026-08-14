package artwork

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// Minimal hand-built FLAC fixtures.
//
// github.com/dhowden/tag's FLAC reader (flac.go) only requires the "fLaC"
// magic followed by a sequence of metadata blocks terminated by one with the
// "last block" flag set; it does not require a STREAMINFO block to be
// present or to precede other blocks, and it never reads past the metadata
// blocks into actual audio frames. That lets these fixtures be a handful of
// hand-assembled bytes instead of a real encoded audio file — no download,
// no external encoder.
//
// The PICTURE block layout mirrors FLAC's METADATA_BLOCK_PICTURE structure,
// verified against dhowden/tag's own readPictureBlock (vorbis.go): picture
// type, mime length + mime, description length + description, width,
// height, color depth, colors used, data length + data — all big-endian
// uint32 fields.
// ---------------------------------------------------------------------------

func writeBE32(buf *bytes.Buffer, v uint32) {
	buf.WriteByte(byte(v >> 24))
	buf.WriteByte(byte(v >> 16))
	buf.WriteByte(byte(v >> 8))
	buf.WriteByte(byte(v))
}

// buildFLACWithPicture returns a minimal FLAC file containing a single
// PICTURE metadata block (type 6) holding mimeType/data, marked as the last
// metadata block.
func buildFLACWithPicture(mimeType string, data []byte) []byte {
	var body bytes.Buffer
	writeBE32(&body, 3) // picture type 3 = "Cover (front)"
	writeBE32(&body, uint32(len(mimeType)))
	body.WriteString(mimeType)
	writeBE32(&body, 0) // description length
	writeBE32(&body, 0) // width
	writeBE32(&body, 0) // height
	writeBE32(&body, 0) // color depth
	writeBE32(&body, 0) // colors used
	writeBE32(&body, uint32(len(data)))
	body.Write(data)

	var file bytes.Buffer
	file.WriteString("fLaC")
	file.WriteByte(0x80 | 6) // last-block flag set, block type 6 = PICTURE
	blockLen := body.Len()
	file.WriteByte(byte(blockLen >> 16))
	file.WriteByte(byte(blockLen >> 8))
	file.WriteByte(byte(blockLen))
	file.Write(body.Bytes())
	return file.Bytes()
}

// buildFLACWithoutPicture returns a minimal FLAC file with a single, empty
// padding block (type 1) and no picture — a track with valid, parseable
// metadata but nothing embedded to extract.
func buildFLACWithoutPicture() []byte {
	var file bytes.Buffer
	file.WriteString("fLaC")
	file.WriteByte(0x80 | 1) // last-block flag set, block type 1 = padding
	file.WriteByte(0)
	file.WriteByte(0)
	file.WriteByte(4)
	file.Write([]byte{0, 0, 0, 0})
	return file.Bytes()
}

func TestExtractEmbedded_FLACWithPicture(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.flac")
	want := []byte("fake-cover-bytes-for-test-not-a-real-image")
	if err := os.WriteFile(path, buildFLACWithPicture("image/jpeg", want), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	img, ok, err := ExtractEmbedded(path)
	if err != nil {
		t.Fatalf("ExtractEmbedded error = %v, want nil", err)
	}
	if !ok {
		t.Fatal("ExtractEmbedded ok = false, want true")
	}
	if img.MIMEType != "image/jpeg" {
		t.Errorf("MIMEType = %q, want image/jpeg", img.MIMEType)
	}
	if !bytes.Equal(img.Data, want) {
		t.Errorf("Data = %q, want %q", img.Data, want)
	}
}

func TestExtractEmbedded_FLACWithoutPicture(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.flac")
	if err := os.WriteFile(path, buildFLACWithoutPicture(), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	_, ok, err := ExtractEmbedded(path)
	if err != nil {
		t.Fatalf("ExtractEmbedded error = %v, want nil", err)
	}
	if ok {
		t.Fatal("ExtractEmbedded ok = true, want false — fixture has no picture block")
	}
}

func TestExtractEmbedded_UnrecognizedFormatIsNotAnError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "not-audio.bin")
	if err := os.WriteFile(path, []byte("this is not an audio file at all, just plain bytes"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	_, ok, err := ExtractEmbedded(path)
	if err != nil {
		t.Fatalf("ExtractEmbedded error = %v, want nil (unrecognized format is a soft miss)", err)
	}
	if ok {
		t.Fatal("ExtractEmbedded ok = true, want false")
	}
}

func TestExtractEmbedded_FileNotFoundIsAnError(t *testing.T) {
	_, _, err := ExtractEmbedded(filepath.Join(t.TempDir(), "does-not-exist.flac"))
	if err == nil {
		t.Fatal("expected a non-nil error for a missing file")
	}
}

func TestFindSibling_CaseInsensitiveMatch(t *testing.T) {
	dir := t.TempDir()
	want := []byte("sibling-cover-bytes")
	if err := os.WriteFile(filepath.Join(dir, "Cover.JPG"), want, 0o600); err != nil {
		t.Fatalf("write sibling fixture: %v", err)
	}

	img, ok, err := FindSibling(dir)
	if err != nil {
		t.Fatalf("FindSibling error = %v, want nil", err)
	}
	if !ok {
		t.Fatal("FindSibling ok = false, want true")
	}
	if img.MIMEType != "image/jpeg" {
		t.Errorf("MIMEType = %q, want image/jpeg", img.MIMEType)
	}
	if !bytes.Equal(img.Data, want) {
		t.Errorf("Data = %q, want %q", img.Data, want)
	}
}

func TestFindSibling_PriorityOrder(t *testing.T) {
	// "cover" outranks "folder" when both are present.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "folder.png"), []byte("folder-bytes"), 0o600); err != nil {
		t.Fatalf("write folder fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cover.png"), []byte("cover-bytes"), 0o600); err != nil {
		t.Fatalf("write cover fixture: %v", err)
	}

	img, ok, err := FindSibling(dir)
	if err != nil || !ok {
		t.Fatalf("FindSibling ok=%v err=%v, want ok=true err=nil", ok, err)
	}
	if string(img.Data) != "cover-bytes" {
		t.Errorf("Data = %q, want cover-bytes (cover.* must win over folder.*)", img.Data)
	}
}

func TestFindSibling_NoMatch(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("nothing here"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	_, ok, err := FindSibling(dir)
	if err != nil {
		t.Fatalf("FindSibling error = %v, want nil", err)
	}
	if ok {
		t.Fatal("FindSibling ok = true, want false — no candidate file present")
	}
}

func TestFindSibling_UnsupportedExtensionIgnored(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "cover.bmp"), []byte("bmp-bytes"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	_, ok, err := FindSibling(dir)
	if err != nil {
		t.Fatalf("FindSibling error = %v, want nil", err)
	}
	if ok {
		t.Fatal("FindSibling ok = true for unsupported .bmp extension, want false")
	}
}

func TestFindSibling_MissingDirectoryIsAnError(t *testing.T) {
	_, _, err := FindSibling(filepath.Join(t.TempDir(), "does-not-exist"))
	if err == nil {
		t.Fatal("expected a non-nil error for a missing directory")
	}
}
