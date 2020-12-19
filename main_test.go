package main

import (
    "testing"
    "fmt"
    "sort"
)

func TestSearcher(t *testing.T) {
    searcher := Searcher{}
    searcher.Load("sonnets.txt")
    // query := "Hamlet"
    // idxs := searcher.SuffixArray.Lookup([]byte(query), -1)
    FD := make(map[int][]string, len(searcher.DF))
    counts := make([]int, 0)
    for k, v := range searcher.DF {
        FD[v] = append(FD[v], k)
        counts = append(counts, v)
    } 
    sort.Sort(sort.Reverse(sort.IntSlice(counts)))
    for i := 0; i < 100; i++ {
        idx := counts[i]
        fmt.Println("max =", idx, FD[idx])
    }

    for i:=0; i<10; i++ {
        doc := searcher.Documents[i]
        fmt.Println("doc ", searcher.CompleteWorks[doc[0]:doc[1]])
    }
}