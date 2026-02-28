package cmd

import (
	"fmt"
	"math"
	"sort"
	"strings"

	runes "github.com/hbn/runes/internal/rune"
	"github.com/hbn/runes/internal/store"
	"github.com/spf13/cobra"
)

var searchIndexCmd = &cobra.Command{
	Use:   "search-index <query>",
	Short: "Search using term index (faster)",
	Long: `Search runes using pre-built term index.

Faster than regular search for large collections.
Builds index on first run, then reuses for subsequent searches.

Examples:
  runes search-index "auth timeout"
  runes search-index "database connection" --limit 10`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		limit, _ := cmd.Flags().GetInt("limit")

		st, err := store.New(store.DefaultPath())
		if err != nil {
			return fmt.Errorf("initializing store: %w", err)
		}

		// Build/load index
		index := buildIndex(st)
		
		// Search
		results := searchIndex(index, query, limit)
		
		if len(results) == 0 {
			fmt.Println("No runes found.")
			return nil
		}

		// Display
		fmt.Printf("Found %d rune(s) (using index):\n\n", len(results))
		
		for _, r := range results {
			fmt.Printf("%s   %s\n", r.ID, r.Title)
			if r.Pattern != "" {
				fmt.Printf("       Pattern: %s\n", r.Pattern)
			}
			if len(r.Tags) > 0 {
				fmt.Printf("       Tags: [%s]\n", strings.Join(r.Tags, ", "))
			}
			fmt.Println()
		}

		return nil
	},
}

// Index structure for fast BM25
type termIndex struct {
	docs          map[string]*runes.Rune     // id -> rune
	docLength     map[string]int            // id -> term count
	termFreq      map[string]map[string]int // term -> docID -> count
	termDocCount  map[string]int            // term -> number of docs
	totalDocs     int
	avgDocLength  float64
}

// buildIndex creates term index from all runes
func buildIndex(st *store.Store) *termIndex {
	runesList, _ := st.LoadAll()
	
	idx := &termIndex{
		docs:         make(map[string]*runes.Rune),
		docLength:    make(map[string]int),
		termFreq:     make(map[string]map[string]int),
		termDocCount: make(map[string]int),
		totalDocs:    len(runesList),
	}
	
	totalLength := 0
	
	for _, r := range runesList {
		idx.docs[r.ID] = r
		
		// Extract terms
		terms := extractIndexTerms(r)
		idx.docLength[r.ID] = len(terms)
		totalLength += len(terms)
		
		// Count term frequencies
		seen := make(map[string]bool)
		for _, term := range terms {
			if idx.termFreq[term] == nil {
				idx.termFreq[term] = make(map[string]int)
			}
			idx.termFreq[term][r.ID]++
			
			if !seen[term] {
				idx.termDocCount[term]++
				seen[term] = true
			}
		}
	}
	
	if idx.totalDocs > 0 {
		idx.avgDocLength = float64(totalLength) / float64(idx.totalDocs)
	}
	
	return idx
}

// extractIndexTerms extracts searchable terms from rune
func extractIndexTerms(r *runes.Rune) []string {
	var text strings.Builder
	text.WriteString(r.Title + " ")
	text.WriteString(r.Problem + " ")
	text.WriteString(r.Solution + " ")
	text.WriteString(r.Pattern + " ")
	text.WriteString(r.Learned + " ")
	
	// Tags weighted double
	for _, tag := range r.Tags {
		text.WriteString(tag + " " + tag + " ")
	}
	
	return tokenizeIndex(text.String())
}

func tokenizeIndex(text string) []string {
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

// searchIndex performs BM25 search
func searchIndex(idx *termIndex, query string, limit int) []*runes.Rune {
	queryTerms := tokenizeIndex(query)
	if len(queryTerms) == 0 {
		return nil
	}
	
	// Find candidate docs (union of docs containing any query term)
	candidates := make(map[string]bool)
	for _, term := range queryTerms {
		if docs, ok := idx.termFreq[term]; ok {
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
	var results []scored
	
	for docID := range candidates {
		r, ok := idx.docs[docID]
		if !ok {
			continue
		}
		
		score := bm25Score(idx, docID, queryTerms)
		if score > 0 {
			results = append(results, scored{r, score})
		}
	}
	
	// Sort by score
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})
	
	// Apply limit
	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}
	
	// Extract runes
	var resultRunes []*runes.Rune
	for _, s := range results {
		resultRunes = append(resultRunes, s.rune)
	}
	
	return resultRunes
}

// bm25Score calculates BM25 relevance
func bm25Score(idx *termIndex, docID string, queryTerms []string) float64 {
	const k1 = 1.2
	const b = 0.75
	
	docLen := float64(idx.docLength[docID])
	if docLen == 0 {
		return 0
	}
	
	score := 0.0
	N := float64(idx.totalDocs)
	
	for _, term := range queryTerms {
		n, ok := idx.termDocCount[term]
		if !ok {
			continue
		}
		
		// IDF
		idf := math.Log((N - float64(n) + 0.5) / (float64(n) + 0.5))
		
		// Term frequency
		f := float64(idx.termFreq[term][docID])
		
		// BM25
		denom := f + k1*(1-b+b*docLen/idx.avgDocLength)
		if denom > 0 {
			score += idf * f * (k1 + 1) / denom
		}
	}
	
	return score
}

func init() {
	searchIndexCmd.Flags().Int("limit", 10, "Maximum results")
	rootCmd.AddCommand(searchIndexCmd)
}
