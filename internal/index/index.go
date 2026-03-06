// Package index provides persistent inverted index for fast BM25 search
package index

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	runes "github.com/sleeplesslord/runes/internal/rune"
)

// Index represents persistent inverted index for BM25
type Index struct {
	Version       int
	BuiltAt       time.Time
	DocumentCount int
	AvgDocLength  float64
	
	// Inverted index: term -> docID -> frequency
	TermDocFreq     map[string]int            // term -> number of docs containing it
	TermOccurrences map[string]map[string]int // term -> docID -> frequency
	DocLengths      map[string]int            // docID -> total term count
}

// New creates empty index
func New() *Index {
	return &Index{
		Version:         1,
		BuiltAt:         time.Now(),
		TermDocFreq:     make(map[string]int),
		TermOccurrences: make(map[string]map[string]int),
		DocLengths:      make(map[string]int),
	}
}

// Build builds index from all runes
func Build(runeList []*runes.Rune) *Index {
	idx := New()
	idx.DocumentCount = len(runeList)
	
	totalLength := 0
	
	for _, r := range runeList {
		terms := extractTerms(r)
		idx.DocLengths[r.ID] = len(terms)
		totalLength += len(terms)
		
		// Count term frequencies
		seen := make(map[string]bool)
		for _, term := range terms {
			if idx.TermOccurrences[term] == nil {
				idx.TermOccurrences[term] = make(map[string]int)
			}
			idx.TermOccurrences[term][r.ID]++
			
			if !seen[term] {
				idx.TermDocFreq[term]++
				seen[term] = true
			}
		}
	}
	
	if idx.DocumentCount > 0 {
		idx.AvgDocLength = float64(totalLength) / float64(idx.DocumentCount)
	}
	
	return idx
}

// IsStale checks if index needs rebuild
func (idx *Index) IsStale(sourcePath string) bool {
	info, err := os.Stat(sourcePath)
	if err != nil {
		return true
	}
	return info.ModTime().After(idx.BuiltAt)
}

// Save persists index to disk
func (idx *Index) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating index directory: %w", err)
	}
	
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating index file: %w", err)
	}
	defer file.Close()
	
	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(idx); err != nil {
		return fmt.Errorf("encoding index: %w", err)
	}
	
	return nil
}

// Load reads index from disk
func Load(path string) (*Index, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	var idx Index
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&idx); err != nil {
		return nil, fmt.Errorf("decoding index: %w", err)
	}
	
	return &idx, nil
}

// LoadOrBuild loads index if fresh, otherwise builds from runes
func LoadOrBuild(indexPath string, runeList []*runes.Rune, sourcePath string) (*Index, error) {
	// Try to load existing
	idx, err := Load(indexPath)
	if err == nil && !idx.IsStale(sourcePath) {
		return idx, nil
	}
	
	// Build new index
	idx = Build(runeList)
	
	// Save for next time
	_ = idx.Save(indexPath) // Best effort
	
	return idx, nil
}

// extractTerms extracts searchable terms from rune
func extractTerms(r *runes.Rune) []string {
	var text string
	text += r.Title + " "
	text += r.Problem + " "
	text += r.Solution + " "
	text += r.Pattern + " "
	text += r.Learned + " "
	
	for _, tag := range r.Tags {
		text += tag + " " + tag + " " // Double weight
	}
	
	return tokenize(text)
}

// tokenize splits text into terms
func tokenize(text string) []string {
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
