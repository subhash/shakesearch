package main

import (
	"bytes"
	"regexp"
	"encoding/json"
	"fmt"
	"strings"
	"sort"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
	ActIndex	  [][]int
	SceneIndex	  [][]int
	PlayIndex	  [][]int
	SonnetIndex	  [][]int
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
	s.LoadIndices()
	return nil
}

func (s *Searcher) LoadIndices() {
	romanExp := "M{0,4}(?:CM|CD|D?C{0,3})(?:XC|XL|L?X{0,3})(?:IX|IV|V?I{0,3})"
	actExp := fmt.Sprintf(`ACT (%[1]s).*`, romanExp)
	s.ActIndex = s.SuffixArray.FindAllIndex(regexp.MustCompile(actExp), -1)
	sceneExp := fmt.Sprintf(`SCENE (%[1]s).*`, romanExp)
	s.SceneIndex = s.SuffixArray.FindAllIndex(regexp.MustCompile(sceneExp), -1)
	playExp := `.+(\r\n)+Contents`
	s.PlayIndex = s.SuffixArray.FindAllIndex(regexp.MustCompile(playExp), -1)
	sonnetExp := `\r\n\s+\d+\s+\r\n`
	s.SonnetIndex = s.SuffixArray.FindAllIndex(regexp.MustCompile(sonnetExp), -1)
}

func (s *Searcher) LookupPlayIndices(idx int) (string, string, string) {
	act, scene, play := "", "", ""
	actIdx := sort.Search(len(s.ActIndex), func(i int) bool { return s.ActIndex[i][0] > idx})
	if actIdx > 0 {
		prev := s.ActIndex[actIdx-1]
		act = s.CompleteWorks[prev[0]:prev[1]]
		act = strings.ReplaceAll(act, "\r", "")
	}
	sceneIdx := sort.Search(len(s.SceneIndex), func(i int) bool { return s.SceneIndex[i][0] > idx})
	if sceneIdx > 0 {
		prev := s.SceneIndex[sceneIdx-1]
		scene = s.CompleteWorks[prev[0]:prev[1]]
		scene = strings.ReplaceAll(scene, "\r", "")
	}
	playIdx := sort.Search(len(s.PlayIndex), func(i int) bool { return s.PlayIndex[i][0] > idx})
	if playIdx > 0 {
		prev := s.PlayIndex[playIdx-1]
		play = s.CompleteWorks[prev[0]:prev[1]]
		play = strings.ReplaceAll(play, "\r\n", "")
		play = strings.ReplaceAll(play, "Contents", "")
	}
	return act, scene, play
}

func (s *Searcher) LookupSonnetIndex(idx int) string {
	sonnet := ""
	sonnetIdx := sort.Search(len(s.SonnetIndex), func(i int) bool { return s.SonnetIndex[i][0] > idx})
	if sonnetIdx > 0 {
		prev := s.SonnetIndex[sonnetIdx-1]
		sonnet = s.CompleteWorks[prev[0]:prev[1]]
		sonnet = strings.ReplaceAll(sonnet, "\r\n", "")
		sonnet = strings.ReplaceAll(sonnet, " ", "")
	}
	return sonnet
}

func (s *Searcher) Search(query string) []map[string]interface{} {
	linesExp := `(?:.+\r\n)*`
	regex := fmt.Sprintf(`(?i)(%[1]s.*%[2]s.*\r\n%[1]s)`, linesExp, query)
	results := s.SearchRegex(regex)
	if len(results) > 0 {
		return results
	} else {
		bagWordsQuery := strings.Join(strings.Fields(query), ".*")
		bagWordsRegex := fmt.Sprintf(`(?i)(%[1]s.*%[2]s.*\r\n%[1]s)`, linesExp, bagWordsQuery)
		bagWordsResults := s.SearchRegex(bagWordsRegex)
		return bagWordsResults
	}
}

func (s *Searcher) SearchRegex(regex string) []map[string]interface{} {
	re := regexp.MustCompile(regex)
	idxs := s.SuffixArray.FindAllIndex(re, -1)
	results := []map[string]interface{}{}
	for _, idx := range idxs {
		itemStart, itemEnd := idx[0], idx[1]
		match := s.CompleteWorks[itemStart:itemEnd]
		snippet := strings.Split(match, "\r\n")
		var item = map[string]interface{}{}
		if itemStart < s.PlayIndex[0][0] {
			if itemStart > s.SonnetIndex[0][0] {
				sonnet := s.LookupSonnetIndex(itemStart)
				item = map[string]interface{}{"snippet": snippet, 
					"sonnet": sonnet}
			} else {
				item = map[string]interface{}{"snippet": snippet}
			}
		} else {
			act, scene, play := s.LookupPlayIndices(itemStart)
			item = map[string]interface{}{"snippet": snippet,
				"play": play,
				"act": act,
				"scene": scene}
		}
		results = append(results, item)
	}
	return results
}
