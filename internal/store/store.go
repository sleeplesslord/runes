package store

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/hbn/runes/internal/index"
	runes "github.com/hbn/runes/internal/rune"
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
func (s *Store) LoadAll(scopes ...Scope) ([]*runes.Rune, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(scopes) == 0 {
		scopes = []Scope{ScopeGlobal}
		if s.HasLocal() {
			scopes = append(scopes, ScopeLocal)
		}
	}

	var allRunes []*runes.Rune
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
func (s *Store) loadFromPath(path string) ([]*runes.Rune, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []*runes.Rune{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var runes []*runes.Rune
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
func (s *Store) Save(r *runes.Rune, scope ...Scope) error {
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
func (s *Store) GetByID(id string) (*runes.Rune, error) {
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

// Search finds runes matching query using persistent index
func (s *Store) Search(query string, limit int) ([]*runes.Rune, error) {
	// Load all runes first (needed for index building and result lookup)
	runes, err := s.LoadAll()
	if err != nil {
		return nil, err
	}

	// Determine index path
	indexPath := filepath.Join(filepath.Dir(s.globalPath), ".runes", "index.gob")
	
	// Try to load or build index
	idx, err := index.Load(indexPath)
	if err != nil || idx.IsStale(s.globalPath) {
		// Build new index
		idx = index.Build(runes)
		// Save for next time (best effort)
		_ = idx.Save(indexPath)
	}

	// Search using index
	return s.searchWithIndex(idx, runes, query, limit)
}

// searchWithIndex performs BM25 search using persistent index
func (s *Store) searchWithIndex(idx *index.Index, runes []*runes.Rune, query string, limit int) ([]*runes.Rune, error) {
	// Create id->rune map for fast lookup
	runeMap := make(map[string]*runes.Rune)
	for _, r := range runes {
		runeMap[r.ID] = r
	}
	
	// Tokenize query
	terms := tokenizeQuery(query)
	if len(terms) == 0 {
		return []*runes.Rune{}, nil
	}
	
	// Find candidates (union of docs containing any query term)
	candidates := make(map[string]bool)
	for _, term := range terms {
		if docs, ok := idx.TermOccurrences[term]; ok {
			for docID := range docs {
				candidates[docID] = true
			}
		}
	}
	
	// Score candidates
	type scored struct {
		rune  *runes.Rune
		score float64
	}
	var scoredRunes []scored
	
	for docID := range candidates {
		r, ok := runeMap[docID]
		if !ok {
			continue
		}
		
		score := s.bm25Score(idx, docID, terms)
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
	
	var results []*runes.Rune
	for _, s := range scoredRunes {
		results = append(results, s.rune)
	}
	
	return results, nil
}

// tokenizeQuery splits query into terms
func tokenizeQuery(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	var terms []string
	
	for _, w := range words {
		w = strings.TrimFunc(w, func(r rune) bool {
			return r < 'a' || r > 'z'
		})
		if len(w) >= 3 {
			terms = append(terms, w)
		}
	}
	
	return terms
}

// bm25Score calculates BM25 relevance score
func (s *Store) bm25Score(idx *index.Index, docID string, queryTerms []string) float64 {
	const k1 = 1.2
	const b = 0.75
	
	docLen := float64(idx.DocLengths[docID])
	if docLen == 0 {
		return 0
	}
	
	score := 0.0
	N := float64(idx.DocumentCount)
	
	for _, term := range queryTerms {
		n, ok := idx.TermDocFreq[term]
		if !ok {
			continue
		}
		
		// IDF: log((N - n + 0.5) / (n + 0.5))
		idf := math.Log((N - float64(n) + 0.5) / (float64(n) + 0.5))
		
		// Term frequency in this document
		f := float64(idx.TermOccurrences[term][docID])
		
		// BM25 formula
		denom := f + k1*(1-b+b*docLen/idx.AvgDocLength)
		if denom > 0 {
			score += idf * f * (k1 + 1) / denom
		}
	}
	
	return score
}

// GetBySaga finds runes linked to saga
func (s *Store) GetBySaga(sagaID string) ([]*runes.Rune, error) {
	runes, err := s.LoadAll()
	if err != nil {
		return nil, err
	}

	var results []*runes.Rune
	for _, r := range runes {
		if r.HasSaga(sagaID) {
			results = append(results, r)
		}
	}

	return results, nil
}

// Update replaces existing rune (searches both scopes)
func (s *Store) Update(updated *runes.Rune) error {
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
func loadFromPath(path string) ([]*runes.Rune, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []*runes.Rune{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var runes []*runes.Rune
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

// Delete removes a rune by ID
func (s *Store) Delete(id string) error {
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
		var filtered []*runes.Rune
		for _, r := range runes {
			if r.ID == id {
				found = true
			} else {
				filtered = append(filtered, r)
			}
		}

		if found {
			return saveToPath(path, filtered)
		}
	}

	return fmt.Errorf("rune not found: %s", id)
}

// saveToPath writes runes to a specific path
func saveToPath(path string, runes []*runes.Rune) error {
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
