package terminal

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Bookmark represents a saved directory location
type Bookmark struct {
	Name        string
	Path        string
	Description string
}

// BookmarkManager handles directory bookmarks
type BookmarkManager struct {
	bookmarks   map[string]Bookmark
	configPath  string
	initialized bool
}

// NewBookmarkManager creates a new bookmark manager
func NewBookmarkManager(configDir string) *BookmarkManager {
	configPath := filepath.Join(configDir, "bookmarks.json")
	return &BookmarkManager{
		bookmarks:  make(map[string]Bookmark),
		configPath: configPath,
	}
}

// Initialize loads bookmarks from the config file
func (bm *BookmarkManager) Initialize() error {
	if bm.initialized {
		return nil
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(bm.configPath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return err
		}
	}

	// Load bookmarks if config file exists
	if _, err := os.Stat(bm.configPath); !os.IsNotExist(err) {
		data, err := os.ReadFile(bm.configPath)
		if err != nil {
			return err
		}

		var bookmarks []Bookmark
		if err := json.Unmarshal(data, &bookmarks); err != nil {
			return err
		}

		for _, bookmark := range bookmarks {
			bm.bookmarks[bookmark.Name] = bookmark
		}
	}

	bm.initialized = true
	return nil
}

// AddBookmark adds a new directory bookmark
func (bm *BookmarkManager) AddBookmark(name, path, description string) error {
	if err := bm.Initialize(); err != nil {
		return err
	}

	if name == "" || path == "" {
		return errors.New("bookmark name and path cannot be empty")
	}

	// Verify the path exists and is a directory
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		return errors.New("path is not a directory")
	}

	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	bm.bookmarks[name] = Bookmark{
		Name:        name,
		Path:        absPath,
		Description: description,
	}

	return bm.saveBookmarks()
}

// RemoveBookmark deletes an existing bookmark
func (bm *BookmarkManager) RemoveBookmark(name string) error {
	if err := bm.Initialize(); err != nil {
		return err
	}

	if _, exists := bm.bookmarks[name]; !exists {
		return errors.New("bookmark does not exist")
	}

	delete(bm.bookmarks, name)
	return bm.saveBookmarks()
}

// GetBookmark retrieves a bookmark by name
func (bm *BookmarkManager) GetBookmark(name string) (Bookmark, error) {
	if err := bm.Initialize(); err != nil {
		return Bookmark{}, err
	}

	bookmark, exists := bm.bookmarks[name]
	if !exists {
		return Bookmark{}, errors.New("bookmark not found")
	}

	return bookmark, nil
}

// ListBookmarks returns all defined bookmarks
func (bm *BookmarkManager) ListBookmarks() []Bookmark {
	if err := bm.Initialize(); err != nil {
		return nil
	}

	bookmarks := make([]Bookmark, 0, len(bm.bookmarks))
	for _, bookmark := range bm.bookmarks {
		bookmarks = append(bookmarks, bookmark)
	}

	return bookmarks
}

// saveBookmarks saves all bookmarks to the config file
func (bm *BookmarkManager) saveBookmarks() error {
	bookmarks := make([]Bookmark, 0, len(bm.bookmarks))
	for _, bookmark := range bm.bookmarks {
		bookmarks = append(bookmarks, bookmark)
	}

	data, err := json.MarshalIndent(bookmarks, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(bm.configPath, data, 0644)
}
