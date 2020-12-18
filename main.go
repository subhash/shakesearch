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
)

func main() {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
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

type Searcher struct {
	CompleteWorks string
	SuffixArray   *suffixarray.Index
	LineBreaks	  [][]int
	Documents	  [][2]int
	DF			  map[string]int
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}
		results := searcher.Search(query[0])
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

func (s *Searcher) Load(filename string) error {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
	s.CompleteWorks = string(dat)
	s.SuffixArray = suffixarray.New(dat)
	
	lineBrkRegex := regexp.MustCompile("(\r\n){2,}")
	s.LineBreaks = s.SuffixArray.FindAllIndex(lineBrkRegex, -1)
	s.Documents = make([][2]int, 0) 
	fmt.Println("Documents - ", len(s.LineBreaks))
	docStart := 0
	for _, lineBrk := range s.LineBreaks {
		s.Documents = append(s.Documents, [2]int{docStart, lineBrk[0]})
		docStart = lineBrk[1]
	}
	s.Documents = append(s.Documents, [2]int{docStart, len(s.CompleteWorks)-1})

	terms := strings.Split(s.CompleteWorks, " ")
	termMap := make(map[string]int)
	for _, t := range terms {
		t = strings.ReplaceAll(strings.ReplaceAll(t, "\r", " "), "\n", " ")
		t = strings.ToLower(t)
		termMap[t] += 1
	}
	fmt.Println("Terms - ", len(terms), len(termMap))
	
	s.DF = make(map[string]int)
	for t, _ := range termMap {
		termOccurences := s.SuffixArray.FindAllIndex(regexp.MustCompile(fmt.PrintS(" %s ", t)), -1)
		if t == string('o') {
			fmt.Println("occurrenced of o ", len(termOccurences))
		}
		tOccur := make([]int)
		
		sort.Ints(termOccurences)
		ti := 0
		for _, doc := range s.Documents {
			docEnd := doc[1]
			if ti >= len(termOccurences) {
				break
			}
			if termOccurences[ti] <= docEnd {
				s.DF[t] += 1
			}
			for ti < len(termOccurences) && termOccurences[ti] <= docEnd {
				ti += 1
			}
		}


	}

	return nil
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


