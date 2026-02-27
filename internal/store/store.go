package store

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/hbn/runes/internal/rune"
)

// Store handles persistence of runes
type Store struct {
	globalPath string
	localPath  string
	mu         sync.RWMutex
}

// New creates a new Store with global path, auto-detects local
func New(globalPath string) (*Store, error) {
	dir := filepath.Dir(globalPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating global store directory: %w", err)
	}

	s := &Store{globalPath: globalPath}

	// Check for local .runes directory
	if localPath := findLocalRunesDir(); localPath != "" && localPath != globalPath {
		s.localPath = localPath
	}

	return s, nil
}

// findLocalRunesDir searches for .runes/ directory in current or parent directories
// Stops at home directory to avoid finding ~/.runes as "local"
func findLocalRunesDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	home, _ := os.UserHomeDir()

	for {
		// Stop at home directory
		if dir == home {
			break
		}

		runesDir := filepath.Join(dir, ".runes")
		if info, err := os.Stat(runesDir); err == nil && info.IsDir() {
			return filepath.Join(runesDir, "runes.jsonl")
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return ""
}

// DefaultPath returns default storage path
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".runes/runes.jsonl"
	}
	return filepath.Join(home, ".runes", "runes.jsonl")
}

// HasLocal returns true if local store exists
func (s *Store) HasLocal() bool {
	return s.localPath != ""
}

// LocalPath returns the local store path
func (s *Store) LocalPath() string {
	return s.localPath
}

// Scope defines where runes are stored
type Scope int

const (
	ScopeGlobal Scope = iota
	ScopeLocal
)

// LoadAll reads runes from specified scopes
func (s *Store) LoadAll(scopes ...Scope) ([]*rune.Rune, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(scopes) == 0 {
		scopes = []Scope{ScopeGlobal}
		if s.HasLocal() {
			scopes = append(scopes, ScopeLocal)
		}
	}

	var allRunes []*rune.Rune
	for _, scope := range scopes {
		var path string
		switch scope {
		case ScopeGlobal:
			path = s.globalPath
		case ScopeLocal:
			path = s.localPath
		}
		if path == "" {
			continue
		}

		runes, err := s.loadFromPath(path)
		if err != nil {
			return nil, fmt.Errorf("loading from %v: %w", scope, err)
		}
		allRunes = append(allRunes, runes...)
	}

	return allRunes, nil
}

// loadFromPath loads runes from a specific path
func (s *Store) loadFromPath(path string) ([]*rune.Rune, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []*rune.Rune{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var runes []*rune.Rune
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var r rune.Rune
		if err := json.Unmarshal(scanner.Bytes(), &r); err != nil {
			continue
		}
		runes = append(runes, &r)
	}

	return runes, scanner.Err()
}

// Save appends a rune (default: local if in project, else global)
func (s *Store) Save(r *rune.Rune, scope ...Scope) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Determine scope
	targetScope := ScopeGlobal
	if s.HasLocal() && (len(scope) == 0 || scope[0] == ScopeLocal) {
		targetScope = ScopeLocal
	}
	if len(scope) > 0 {
		targetScope = scope[0]
	}

	path := s.globalPath
	if targetScope == ScopeLocal && s.localPath != "" {
		path = s.localPath
	}

	// Ensure directory exists for local
	if targetScope == ScopeLocal {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating local store directory: %w", err)
		}
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening store: %w", err)
	}
	defer file.Close()

	data, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("encoding rune: %w", err)
	}

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("writing rune: %w", err)
	}
	if _, err := file.WriteString("\n"); err != nil {
		return fmt.Errorf("writing newline: %w", err)
	}

	return nil
}

// InitLocal creates a local .runes directory in current working directory
func (s *Store) InitLocal() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	runesDir := filepath.Join(cwd, ".runes")
	if err := os.MkdirAll(runesDir, 0755); err != nil {
		return fmt.Errorf("creating .runes directory: %w", err)
	}

	s.localPath = filepath.Join(runesDir, "runes.jsonl")
	return nil
}

// GetByID finds rune by ID
func (s *Store) GetByID(id string) (*rune.Rune, error) {
	runes, err := s.LoadAll()
	if err != nil {
		return nil, err
	}

	for _, r := range runes {
		if r.ID == id {
			return r, nil
		}
	}

	return nil, fmt.Errorf("rune not found: %s", id)
}

// Search finds runes matching query
func (s *Store) Search(query string, limit int) ([]*rune.Rune, error) {
	runes, err := s.LoadAll()
	if err != nil {
		return nil, err
	}

	// Score and filter
	type scored struct {
		rune  *rune.Rune
		score float64
	}
	var scoredRunes []scored

	for _, r := range runes {
		score := r.SearchScore(query)
		if score > 0 {
			scoredRunes = append(scoredRunes, scored{r, score})
		}
	}

	// Sort by score desc
	sort.Slice(scoredRunes, func(i, j int) bool {
		return scoredRunes[i].score > scoredRunes[j].score
	})

	// Apply limit
	if limit > 0 && limit < len(scoredRunes) {
		scoredRunes = scoredRunes[:limit]
	}

	var results []*rune.Rune
	for _, s := range scoredRunes {
		results = append(results, s.rune)
	}

	return results, nil
}

// GetBySaga finds runes linked to saga
func (s *Store) GetBySaga(sagaID string) ([]*rune.Rune, error) {
	runes, err := s.LoadAll()
	if err != nil {
		return nil, err
	}

	var results []*rune.Rune
	for _, r := range runes {
		if r.HasSaga(sagaID) {
			results = append(results, r)
		}
	}

	return results, nil
}

// Update replaces existing rune (searches both scopes)
func (s *Store) Update(updated *rune.Rune) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Try local first, then global
	scopes := []string{s.localPath, s.globalPath}
	for _, path := range scopes {
		if path == "" {
			continue
		}

		runes, err := loadFromPath(path)
		if err != nil {
			return err
		}

		found := false
		for i, r := range runes {
			if r.ID == updated.ID {
				runes[i] = updated
				found = true
				break
			}
		}

		if found {
			return saveToPath(path, runes)
		}
	}

	return fmt.Errorf("rune not found: %s", updated.ID)
}

// loadFromPath loads runes from a specific path
func loadFromPath(path string) ([]*rune.Rune, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []*rune.Rune{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var runes []*rune.Rune
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var r rune.Rune
		if err := json.Unmarshal(scanner.Bytes(), &r); err != nil {
			continue
		}
		runes = append(runes, &r)
	}

	return runes, scanner.Err()
}

// saveToPath writes runes to a specific path
func saveToPath(path string, runes []*rune.Rune) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating store: %w", err)
	}
	defer file.Close()

	for _, r := range runes {
		data, err := json.Marshal(r)
		if err != nil {
			return fmt.Errorf("encoding rune: %w", err)
		}
		if _, err := file.Write(data); err != nil {
			return fmt.Errorf("writing rune: %w", err)
		}
		if _, err := file.WriteString("\n"); err != nil {
			return fmt.Errorf("writing newline: %w", err)
		}
	}

	return nil
}
