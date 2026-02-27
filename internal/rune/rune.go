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

// SearchScore returns relevance score for query (0-1)
func (r *Rune) SearchScore(query string) float64 {
	query = strings.ToLower(query)
	
	// Check title (highest weight)
	if strings.Contains(strings.ToLower(r.Title), query) {
		return 1.0
	}
	
	// Check problem/solution
	if strings.Contains(strings.ToLower(r.Problem), query) {
		return 0.8
	}
	if strings.Contains(strings.ToLower(r.Solution), query) {
		return 0.7
	}
	
	// Check tags
	for _, tag := range r.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return 0.6
		}
	}
	
	// Check pattern/learned
	if strings.Contains(strings.ToLower(r.Pattern), query) {
		return 0.5
	}
	if strings.Contains(strings.ToLower(r.Learned), query) {
		return 0.4
	}
	
	// Check linked sagas
	for _, sagaID := range r.Sagas {
		if strings.Contains(strings.ToLower(sagaID), query) {
			return 0.9 // High relevance for saga links
		}
	}
	
	return 0.0
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
