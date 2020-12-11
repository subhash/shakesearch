# ShakeSearch

**Demo**: https://still-oasis-25000.herokuapp.com/

## Notes
1. Case insensitive exact search. When that fails, bag of words to support punctuation problems. Stemming etc does not seem useful for this usecase
2. Extracts stanza snippet around query
3. Sonnet, Play, Act, Scene are identified as appropriate
4. First time in Golang. Please excuse non-idiomatic code, if any.
