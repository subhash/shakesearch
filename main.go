package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"regexp"
	"math"
)

func main() {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt", 5)
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/search", handleSearch(searcher))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("Listening on port %s...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

type TFKey struct {
    term string
    doc  int
}


type Searcher struct {
	CompleteWorks string
	SuffixArray   *suffixarray.Index
	Documents	  [][2]int
	DocumentTerms map[int]int
	DF			  map[string]map[int]bool
	TF			  map[TFKey]int
	nGrams		  int
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}
		results := searcher.SearchTFIDF(query[0])
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		err := enc.Encode(results)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}
}

func (s *Searcher) Load(filename string, nGrams int) error {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
	s.CompleteWorks = string(dat)
	s.SuffixArray = suffixarray.New(dat)
	s.nGrams = nGrams

	// Index documents
	lineBrkRegex := regexp.MustCompile("(\r\n){2,}")
	lineBreaks := s.SuffixArray.FindAllIndex(lineBrkRegex, -1)
	s.Documents = make([][2]int, 0) 
	docEndIdx := make([]int, 0)
	docStart := 0
	for _, lineBrk := range lineBreaks {
		s.Documents = append(s.Documents, [2]int{docStart, lineBrk[0]})
		docEndIdx = append(docEndIdx, lineBrk[0])
		docStart = lineBrk[1]
	}
	s.Documents = append(s.Documents, [2]int{docStart, len(s.CompleteWorks)-1})

	// Index terms
	sepIdx := s.SuffixArray.FindAllIndex(regexp.MustCompile(`[\r\n\s]+`), -1)
	terms := make([]string, 0)
	termIdx := make([][2]int, 0)
	termStart := 0
	for _, idx := range sepIdx {
		term := s.CompleteWorks[termStart:idx[0]]
		term = string(regexp.MustCompile("[^a-zA-Z0-9]+").ReplaceAll([]byte(term), []byte("")))
		term = strings.ToLower(term)
		if term != "" {
			terms = append(terms, term)
			termIdx = append(termIdx, [2]int{termStart, idx[1]})
		}
		termStart = idx[1]
	} 
	
	// Index DF and TF
	s.DF = map[string]map[int]bool{}
	s.TF = make(map[TFKey]int)
	s.DocumentTerms = make(map[int]int)
	for i, tIdx := range termIdx {
		doc := sort.SearchInts(docEndIdx, tIdx[0])
		term := terms[i]
		if s.DF[term] == nil {
			s.DF[term] = map[int]bool{}
		}
		s.DF[term][doc] = true
		s.TF[TFKey{term: term, doc: doc}] += 1
		// NGRAM
		nGram := term
		for ngi:=1; ngi < s.nGrams && i + ngi < len(termIdx); ngi++ {
			nGram += " " + terms[i+ngi]
			s.TF[TFKey{term: nGram, doc: doc}] += 1		
		}
		s.DocumentTerms[doc] += 1
	}

	return nil
}

func (s *Searcher) SearchTFIDF(query string) []map[string]interface{} {
	tokens := strings.Split(query, " ")
	results := []struct {
		Doc 	[2]int
		Score	float64
		TermScores []float64
		DocSize	int
	}{}
	for i:=0; i<len(tokens); i++ {
		tokens[i] = strings.ToLower(tokens[i])
		tokens[i] = string(regexp.MustCompile("[^a-zA-Z0-9]+").ReplaceAll(
			[]byte(tokens[i]), []byte("")))
	}
	for di, d := range s.Documents {
		totalScore := 0.0
		termScores := []float64{}
		Nt := s.DocumentTerms[di]
		for ti, t := range tokens {
			Nd := float64(len(s.Documents))
			df := float64(len(s.DF[t]))
			idf := float64(Nd)/float64(df+1)
			tf := s.TF[TFKey{ term: t, doc: di }]
			// NGRAM
			nGram := t
			for ngi:=1; ngi < s.nGrams && ti + ngi < len(tokens); ngi++ {
				nGram += " " + tokens[ti + ngi]
				tf += (ngi+1)*s.TF[TFKey{ term: nGram, doc: di }]
			}

			// Regularize short documents
			if Nt < 100 {
				Nt = Nt + 100
			}
			termScore := (float64(tf)/float64(Nt)) * 
				math.Log(idf)
			totalScore += termScore
			termScores = append(termScores, termScore)
		}
		score := totalScore
		if score > 0.0 {
			results = append(results, struct {
				Doc 	[2]int
				Score	float64
				TermScores []float64
				DocSize int
			} { Doc: d, Score: score, TermScores: termScores,
				DocSize: s.DocumentTerms[di] })
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	ranked := []map[string]interface{}{} 
	for _, res := range results {
		ranked = append(ranked, map[string]interface{}{
			"score": res.Score,
			"doc-size": res.DocSize,
			"snippet": s.CompleteWorks[res.Doc[0]:res.Doc[1]],
			"term-scores": res.TermScores,
		})
	}
	return ranked
}

func (s *Searcher) Search(query string) []string {
	idxs := s.SuffixArray.Lookup([]byte(query), -1)
	results := []string{}
	for _, idx := range idxs {
		results = append(results, s.CompleteWorks[idx-250:idx+250])
	}
	fmt.Println(results)
	return results
}


