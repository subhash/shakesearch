package main

import (
    "testing"
    "strings"
)

func TestTF(t *testing.T) {
    s := Searcher{}
    s.Load("sonnets.txt", 3)
    res1 := s.TF[TFKey{term: "fairest", doc: 3}]
    res2 := s.TF[TFKey{term: "thy", doc: 5}]
    if res1 != 1 || res2 != 7 {
        t.Errorf("Unexpected results %v %v", res1, res2)
    }
}

func TestDF(t *testing.T) {
    s := Searcher{}
    s.Load("sonnets.txt", 3)
    res1 := len(s.DF["fair"])
    res2 := len(s.Documents)
    if res1 != 32 || res2 != 311 {
        t.Errorf("Unexpected results %v %v", res1, res2)
    }
}


func TestSearch(t *testing.T) {
    searcher := Searcher{}
    searcher.Load("completeworks.txt", 5)
    results := searcher.SearchTFIDF("love is not love")
    prefix := "Let me not to the marriage of true minds" 
    actual := results[0]["snippet"].(string)
    if !strings.HasPrefix(actual, prefix) {
        t.Errorf("Expected %v, Obtained %v", prefix, actual)
    }

    results = searcher.SearchTFIDF("to be or not to be")
    prefix = "HAMLET.\r\nTo be, or not to be, that is the question:" 
    actual = results[0]["snippet"].(string)
    if !strings.HasPrefix(actual, prefix) {
        t.Errorf("Expected %v, Obtained %v", prefix, actual)
    }
    
    results = searcher.SearchTFIDF("we call a rose by other name")
    prefix = "JULIET.\r\n’Tis but thy name that is my enemy;" 
    actual = results[0]["snippet"].(string)
    if !strings.HasPrefix(actual, prefix) {
        t.Errorf("Expected %v, Obtained %v", prefix, actual)
    }
    

}

func Hamlet(t *testing.T) {
    searcher := Searcher{}
    searcher.Load("hamlet.txt", 3)
    results := searcher.SearchTFIDF("to be or not to be")
    prefix := "HAMLET.\r\nTo be, or not to be, that is the question:" 
    actual := results[0]["snippet"].(string)
    if !strings.HasPrefix(actual, prefix) {
        t.Errorf("Expected %v, Obtained %v", prefix, actual)
    }
}
