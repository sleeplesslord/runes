package rune

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Rune represents a captured solution or knowledge entry
type Rune struct {
	ID        string   `yaml:"id" json:"id"`
	Title     string   `yaml:"title" json:"title"`
	Problem   string   `yaml:"problem" json:"problem"`
	Solution  string   `yaml:"solution" json:"solution"`
	Pattern   string   `yaml:"pattern,omitempty" json:"pattern,omitempty"`
	Tags      []string `yaml:"tags" json:"tags"`
	Sagas     []string `yaml:"sagas,omitempty" json:"sagas,omitempty"`
	Learned   string   `yaml:"learned,omitempty" json:"learned,omitempty"`
	CreatedAt time.Time `yaml:"created_at" json:"created_at"`
	UpdatedAt time.Time `yaml:"updated_at" json:"updated_at"`
}

// ToYAML serializes rune to YAML with frontmatter
func (r *Rune) ToYAML() ([]byte, error) {
	return yaml.Marshal(r)
}

// FromYAML parses rune from YAML
func FromYAML(data []byte) (*Rune, error) {
	var r Rune
	if err := yaml.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}
	return &r, nil
}

// New creates a new rune with generated ID
func New(title string) *Rune {
	now := time.Now()
	return &Rune{
		ID:        generateID(),
		Title:     title,
		Tags:      []string{},
		Sagas:     []string{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// HasTag returns true if rune has the given tag
func (r *Rune) HasTag(tag string) bool {
	for _, t := range r.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// AddTag adds a tag if not present
func (r *Rune) AddTag(tag string) {
	if r.HasTag(tag) {
		return
	}
	r.Tags = append(r.Tags, tag)
	r.UpdatedAt = time.Now()
}

// HasSaga returns true if rune links to saga
func (r *Rune) HasSaga(sagaID string) bool {
	for _, s := range r.Sagas {
		if s == sagaID {
			return true
		}
	}
	return false
}

// LinkSaga adds saga reference
func (r *Rune) LinkSaga(sagaID string) {
	if r.HasSaga(sagaID) {
		return
	}
	r.Sagas = append(r.Sagas, sagaID)
	r.UpdatedAt = time.Now()
}

// SearchScore returns relevance score for query using BM25-like ranking
func (r *Rune) SearchScore(query string) float64 {
	query = strings.ToLower(query)
	queryTerms := strings.Fields(query)
	
	if len(queryTerms) == 0 {
		return 0.0
	}
	
	// BM25 parameters
	const k1 = 1.2  // term frequency saturation
	const b = 0.75  // length normalization
	
	// Field weights (title most important)
	weights := map[string]float64{
		"title":    3.0,
		"pattern":  2.5,
		"tags":     2.0,
		"problem":  1.5,
		"solution": 1.3,
		"learned":  1.0,
	}
	
	// Calculate field lengths (for normalization)
	fieldLengths := map[string]int{
		"title":    len(strings.Fields(r.Title)),
		"pattern":  len(strings.Fields(r.Pattern)),
		"tags":     len(r.Tags),
		"problem":  len(strings.Fields(r.Problem)),
		"solution": len(strings.Fields(r.Solution)),
		"learned":  len(strings.Fields(r.Learned)),
	}
	
	// Average field lengths (approximate)
	avgLengths := map[string]float64{
		"title":    10.0,
		"pattern":  3.0,
		"tags":     3.0,
		"problem":  50.0,
		"solution": 50.0,
		"learned":  30.0,
	}
	
	score := 0.0
	
	for _, term := range queryTerms {
		// Score each field for this term
		for field, weight := range weights {
			freq := r.termFrequency(term, field)
			if freq == 0 {
				continue
			}
			
			fieldLen := float64(fieldLengths[field])
			avgLen := avgLengths[field]
			
			// BM25 formula: IDF * (freq * (k1 + 1)) / (freq + k1 * (1 - b + b * fieldLen/avgLen))
			// Simplified: just use term frequency with length normalization
			norm := 1.0 - b + b*(fieldLen/avgLen)
			fieldScore := weight * (freq * (k1 + 1)) / (freq + k1*norm)
			
			score += fieldScore
		}
	}
	
	// Normalize by number of query terms
	return score / float64(len(queryTerms))
}

// termFrequency counts occurrences of term in field
func (r *Rune) termFrequency(term, field string) float64 {
	text := ""
	switch field {
	case "title":
		text = r.Title
	case "pattern":
		text = r.Pattern
	case "problem":
		text = r.Problem
	case "solution":
		text = r.Solution
	case "learned":
		text = r.Learned
	case "tags":
		count := 0.0
		for _, tag := range r.Tags {
			if strings.Contains(strings.ToLower(tag), term) {
				count += 1.0
			}
		}
		return count
	}
	
	// Count occurrences in text
	text = strings.ToLower(text)
	term = strings.ToLower(term)
	count := 0.0
	for {
		idx := strings.Index(text, term)
		if idx == -1 {
			break
		}
		count++
		if idx+len(term) >= len(text) {
			break
		}
		text = text[idx+len(term):]
	}
	return count
}

// generateID creates unique identifier
func generateID() string {
	const alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
	now := time.Now().UnixNano()
	result := make([]byte, 4)
	for i := 0; i < 4; i++ {
		result[i] = alphabet[now%int64(len(alphabet))]
		now /= int64(len(alphabet))
	}
	return string(result)
}
