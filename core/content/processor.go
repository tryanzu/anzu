package content

import (
	"log"
	"time"
)

// Content processor definition.
type Processor func(Deps, Parseable, tags) (Parseable, error)
type Preprocessor func(Deps, Parseable) (Parseable, error)

// Postprocess a parseable type.
func Postprocess(deps Deps, c Parseable) (processed Parseable, err error) {
	starts := time.Now()
	list := parseTags(c)
	pipeline := []Processor{
		postReplaceMentionTags,
	}

	// Run pipeline over parseable.
	processed = c
	for _, fn := range pipeline {
		processed, err = fn(deps, processed, list)

		if err != nil {
			return
		}
	}

	elapsed := time.Since(starts)
	log.Printf("Parsable postprocess took: %v\n", elapsed)
	return
}

// Preprocess a parseable type.
func Preprocess(deps Deps, c Parseable) (processed Parseable, err error) {
	starts := time.Now()
	pipeline := []Preprocessor{
		preReplaceMentionTags,
	}

	// Run pipeline over parseable.
	processed = c
	for _, fn := range pipeline {
		processed, err = fn(deps, processed)

		if err != nil {
			return
		}
	}

	elapsed := time.Since(starts)
	log.Printf("Parsable preprocess took: %v\n", elapsed)
	return
}

func parseTags(c Parseable) (list []tag) {

	// Use regex to find all tags inside the parseable content.
	found := tagRegex.FindAllString(c.GetContent(), -1)
	for _, match := range found {

		// Having parsed all tags now destructure the tag params.
		params := tagParamsRegex.FindAllString(match, -1)
		count := len(params) - 1

		for n, param := range params {
			if n != count {
				params[n] = param[:len(param)-1]
			}
		}

		if len(params) > 0 {
			list = append(list, tag{match, params[0], params[1:]})
		}
	}
	return
}
