package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fjrevoredo/openclaudio/internal/markdown"
)

func TestSaveDetectsConflict(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "note.md")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	svc, err := New(dir, markdown.New())
	if err != nil {
		t.Fatal(err)
	}
	doc, err := svc.Read("note.md", "split")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err = svc.Save(SaveRequest{
		RelativePath:   "note.md",
		Text:           "new text",
		LastModifiedNS: doc.LastModifiedNS,
		ContentHash:    doc.ContentHash,
	})
	if err == nil {
		t.Fatal("Save() error = nil, want conflict")
	}
	if _, ok := err.(*ConflictError); !ok {
		t.Fatalf("Save() error = %T, want *ConflictError", err)
	}
}

func TestCopyPathKinds(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "note.md")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	svc, err := New(dir, markdown.New())
	if err != nil {
		t.Fatal(err)
	}

	got, err := svc.CopyPath("note.md", "relative_backticks")
	if err != nil {
		t.Fatal(err)
	}
	if got != "`note.md`" {
		t.Fatalf("CopyPath() = %q, want `note.md`", got)
	}
}

func TestInvalidUTF8FilesAreReadOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.md")
	if err := os.WriteFile(path, []byte{0xff, 0xfe, 0xfd}, 0o644); err != nil {
		t.Fatal(err)
	}

	svc, err := New(dir, markdown.New())
	if err != nil {
		t.Fatal(err)
	}

	doc, err := svc.Read("bad.md", "split")
	if err != nil {
		t.Fatal(err)
	}
	if !doc.InvalidUTF8 || !doc.IsReadOnly {
		t.Fatalf("doc = %+v, want invalid UTF-8 read-only", doc)
	}
	if doc.View != "raw" {
		t.Fatalf("doc.View = %q, want raw", doc.View)
	}
	if !strings.Contains(doc.Text, "ff fe fd") {
		t.Fatalf("doc.Text = %q, want hex dump", doc.Text)
	}

	_, err = svc.Save(SaveRequest{
		RelativePath:   "bad.md",
		Text:           "replacement",
		LastModifiedNS: doc.LastModifiedNS,
		ContentHash:    doc.ContentHash,
	})
	if err == nil || err.Error() != "invalid UTF-8 files are read-only in v1" {
		t.Fatalf("Save() error = %v, want invalid UTF-8 rejection", err)
	}
}

func TestAbsolutePathsAreRejected(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "note.md")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	svc, err := New(dir, markdown.New())
	if err != nil {
		t.Fatal(err)
	}

	if _, err := svc.Read("/note.md", "raw"); err == nil || err.Error() != "absolute paths are not allowed" {
		t.Fatalf("Read() error = %v, want absolute path rejection", err)
	}

	if _, err := svc.CopyPath("/note.md", "relative"); err == nil || err.Error() != "absolute paths are not allowed" {
		t.Fatalf("CopyPath() error = %v, want absolute path rejection", err)
	}

	_, err = svc.Save(SaveRequest{
		RelativePath:   "/note.md",
		Text:           "replacement",
		LastModifiedNS: 0,
		ContentHash:    "",
	})
	if err == nil || err.Error() != "absolute paths are not allowed" {
		t.Fatalf("Save() error = %v, want absolute path rejection", err)
	}
}
