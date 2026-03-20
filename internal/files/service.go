package files

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fjrevoredo/openclaudio/internal/markdown"
)

type Service struct {
	root     string
	rootReal string
	md       *markdown.Renderer
}

type TreeNode struct {
	Name         string
	RelativePath string
	DOMID        string
	IsDir        bool
	Size         int64
	ModTime      time.Time
	IsMarkdown   bool
}

type Document struct {
	Name           string
	RelativePath   string
	AbsolutePath   string
	Text           string
	RenderedHTML   string
	View           string
	IsMarkdown     bool
	IsReadOnly     bool
	InvalidUTF8    bool
	LastModifiedNS int64
	ContentHash    string
	Size           int64
}

type SaveRequest struct {
	RelativePath   string
	Text           string
	LastModifiedNS int64
	ContentHash    string
}

type SaveResult struct {
	LastModifiedNS int64  `json:"lastModifiedNs"`
	ContentHash    string `json:"contentHash"`
	RenderedHTML   string `json:"renderedHtml"`
	Message        string `json:"message"`
}

type ConflictError struct {
	Message        string `json:"message"`
	LastModifiedNS int64  `json:"lastModifiedNs"`
	ContentHash    string `json:"contentHash"`
}

func (e *ConflictError) Error() string { return e.Message }

func New(root string, md *markdown.Renderer) (*Service, error) {
	rootReal, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace root: %w", err)
	}
	return &Service{root: root, rootReal: rootReal, md: md}, nil
}

func (s *Service) List(relPath, query string) ([]TreeNode, error) {
	cleanRel, err := cleanRelative(relPath)
	if err != nil {
		return nil, err
	}
	dirPath, err := s.resolveExisting(cleanRel)
	if err != nil {
		return nil, err
	}
	stat, err := os.Stat(dirPath)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, errors.New("path is not a directory")
	}

	if query != "" {
		return s.search(cleanRel, dirPath, query)
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	nodes := make([]TreeNode, 0, len(entries))
	for _, entry := range entries {
		if entry.Name() == ".git" {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		childRel := joinRelative(cleanRel, entry.Name())
		nodes = append(nodes, TreeNode{
			Name:         entry.Name(),
			RelativePath: childRel,
			DOMID:        domID(childRel),
			IsDir:        entry.IsDir(),
			Size:         info.Size(),
			ModTime:      info.ModTime(),
			IsMarkdown:   isMarkdown(entry.Name()),
		})
	}

	sortTree(nodes)
	return nodes, nil
}

func (s *Service) search(baseRel, baseDir, query string) ([]TreeNode, error) {
	var nodes []TreeNode
	needle := strings.ToLower(query)

	err := filepath.WalkDir(baseDir, func(fullPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() == ".git" {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if fullPath == baseDir {
			return nil
		}
		rel, err := filepath.Rel(s.rootReal, fullPath)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if !strings.Contains(strings.ToLower(rel), needle) {
			return nil
		}
		info, statErr := d.Info()
		if statErr != nil {
			return statErr
		}
		nodes = append(nodes, TreeNode{
			Name:         d.Name(),
			RelativePath: rel,
			DOMID:        domID(rel),
			IsDir:        d.IsDir(),
			Size:         info.Size(),
			ModTime:      info.ModTime(),
			IsMarkdown:   isMarkdown(d.Name()),
		})
		if d.IsDir() {
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sortTree(nodes)
	return nodes, nil
}

func (s *Service) Read(relPath, view string) (Document, error) {
	cleanRel, err := cleanRelative(relPath)
	if err != nil {
		return Document{}, err
	}
	fullPath, err := s.resolveExisting(cleanRel)
	if err != nil {
		return Document{}, err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return Document{}, err
	}
	if info.IsDir() {
		return Document{}, errors.New("path is a directory")
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return Document{}, err
	}

	doc := Document{
		Name:           filepath.Base(fullPath),
		RelativePath:   cleanRel,
		AbsolutePath:   fullPath,
		View:           normalizeView(view),
		IsMarkdown:     isMarkdown(fullPath),
		LastModifiedNS: info.ModTime().UnixNano(),
		ContentHash:    hashBytes(data),
		Size:           info.Size(),
	}

	if !utf8.Valid(data) {
		doc.Text = hex.Dump(data)
		doc.InvalidUTF8 = true
		doc.IsReadOnly = true
		doc.View = "raw"
		return doc, nil
	}

	doc.Text = string(data)
	if doc.IsMarkdown {
		rendered, err := s.md.Render(doc.Text)
		if err != nil {
			return Document{}, err
		}
		doc.RenderedHTML = rendered
	}

	return doc, nil
}

func (s *Service) Save(req SaveRequest) (SaveResult, error) {
	cleanRel, err := cleanRelative(req.RelativePath)
	if err != nil {
		return SaveResult{}, err
	}
	fullPath, err := s.resolveExisting(cleanRel)
	if err != nil {
		return SaveResult{}, err
	}

	currentBytes, err := os.ReadFile(fullPath)
	if err != nil {
		return SaveResult{}, err
	}
	info, err := os.Stat(fullPath)
	if err != nil {
		return SaveResult{}, err
	}

	currentHash := hashBytes(currentBytes)
	currentMod := info.ModTime().UnixNano()
	if currentMod != req.LastModifiedNS || currentHash != req.ContentHash {
		return SaveResult{}, &ConflictError{
			Message:        "file changed on disk; reload before saving",
			LastModifiedNS: currentMod,
			ContentHash:    currentHash,
		}
	}
	if !utf8.Valid(currentBytes) {
		return SaveResult{}, errors.New("invalid UTF-8 files are read-only in v1")
	}

	if err := os.WriteFile(fullPath, []byte(req.Text), info.Mode().Perm()); err != nil {
		return SaveResult{}, err
	}

	newInfo, err := os.Stat(fullPath)
	if err != nil {
		return SaveResult{}, err
	}

	res := SaveResult{
		LastModifiedNS: newInfo.ModTime().UnixNano(),
		ContentHash:    hashBytes([]byte(req.Text)),
		Message:        "saved",
	}
	if isMarkdown(fullPath) {
		rendered, err := s.md.Render(req.Text)
		if err != nil {
			return SaveResult{}, err
		}
		res.RenderedHTML = rendered
	}
	return res, nil
}

func (s *Service) CopyPath(relPath, kind string) (string, error) {
	cleanRel, err := cleanRelative(relPath)
	if err != nil {
		return "", err
	}
	fullPath, err := s.resolveExisting(cleanRel)
	if err != nil {
		return "", err
	}

	switch kind {
	case "relative":
		return cleanRel, nil
	case "relative_backticks":
		return "`" + cleanRel + "`", nil
	case "absolute":
		return fullPath, nil
	default:
		return "", errors.New("invalid path copy kind")
	}
}

func (s *Service) resolveExisting(rel string) (string, error) {
	target := filepath.Join(s.rootReal, filepath.FromSlash(rel))
	real, err := filepath.EvalSymlinks(target)
	if err != nil {
		return "", err
	}
	if !withinRoot(s.rootReal, real) {
		return "", errors.New("path escapes workspace root")
	}
	return real, nil
}

func cleanRelative(rel string) (string, error) {
	if rel == "" || rel == "." || rel == "/" {
		return "", nil
	}
	clean := path.Clean("/" + rel)
	clean = strings.TrimPrefix(clean, "/")
	if clean == "." || clean == "" {
		return "", nil
	}
	if strings.HasPrefix(clean, "../") || clean == ".." {
		return "", errors.New("invalid relative path")
	}
	return clean, nil
}

func joinRelative(parent, child string) string {
	if parent == "" {
		return child
	}
	return parent + "/" + child
}

func withinRoot(root, target string) bool {
	return target == root || strings.HasPrefix(target, root+string(os.PathSeparator))
}

func isMarkdown(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".md", ".markdown", ".mdown":
		return true
	default:
		return false
	}
}

func sortTree(nodes []TreeNode) {
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].IsDir != nodes[j].IsDir {
			return nodes[i].IsDir
		}
		return strings.ToLower(nodes[i].Name) < strings.ToLower(nodes[j].Name)
	})
}

func domID(rel string) string {
	sum := sha256.Sum256([]byte(rel))
	return "node-" + hex.EncodeToString(sum[:6])
}

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func normalizeView(view string) string {
	switch view {
	case "raw", "rendered", "split":
		return view
	default:
		return "split"
	}
}
