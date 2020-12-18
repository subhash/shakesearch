package main

import (
    "testing"
    "fmt"
    "sort"
)

func TestSearcher(t *testing.T) {
    searcher := Searcher{}
    searcher.Load("hamlet.txt")
    // query := "Hamlet"
    // idxs := searcher.SuffixArray.Lookup([]byte(query), -1)
    FD := make(map[int][]string, len(searcher.DF))
    counts := make([]int, 0)
    for k, v := range searcher.DF {
        if v > 400 {
            fmt.Println(k, v)
        }
        FD[v] = append(FD[v], k)
        counts = append(counts, v)
    } 
    sort.Sort(sort.Reverse(sort.IntSlice(counts)))
    for i := 0; i < 10; i++ {
        fmt.Println("max =", counts[i], FD[counts[i]], len(searcher.DF))
    }

    first := searcher.Documents[3]
    fmt.Println("first ", searcher.CompleteWorks[first[0]:first[1]])
}