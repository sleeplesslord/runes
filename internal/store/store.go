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
	path string
	mu   sync.RWMutex
}

// New creates a new Store
func New(path string) (*Store, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating store directory: %w", err)
	}
	return &Store{path: path}, nil
}

// DefaultPath returns default storage path
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".runes/runes.jsonl"
	}
	return filepath.Join(home, ".runes", "runes.jsonl")
}

// LoadAll reads all runes
func (s *Store) LoadAll() ([]*rune.Rune, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	file, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []*rune.Rune{}, nil
		}
		return nil, fmt.Errorf("opening store: %w", err)
	}
	defer file.Close()

	var runes []*rune.Rune
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var r rune.Rune
		if err := json.Unmarshal(scanner.Bytes(), &r); err != nil {
			continue // Skip malformed
		}
		runes = append(runes, &r)
	}

	return runes, scanner.Err()
}

// Save appends a rune
func (s *Store) Save(r *rune.Rune) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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

// Update replaces existing rune
func (s *Store) Update(updated *rune.Rune) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	runes, err := s.loadAllUnlocked()
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

	if !found {
		return fmt.Errorf("rune not found: %s", updated.ID)
	}

	return s.saveAllUnlocked(runes)
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

// loadAllUnlocked reads without locking
func (s *Store) loadAllUnlocked() ([]*rune.Rune, error) {
	file, err := os.Open(s.path)
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

// saveAllUnlocked writes without locking
func (s *Store) saveAllUnlocked(runes []*rune.Rune) error {
	file, err := os.Create(s.path)
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
