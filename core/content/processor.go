package content

import (
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("content")

// Content processor definition.
type Processor func(deps, Parseable, tags) (Parseable, error)
type Preprocessor func(deps, Parseable) (Parseable, error)

// Postprocess a parseable type.
func Postprocess(d deps, c Parseable) (processed Parseable, err error) {
	starts := time.Now()
	list := parseTags(c)
	pipeline := []Processor{
		postReplaceMentionTags,
		postReplaceAssetTags,
	}

	// Run pipeline over parseable.
	processed = c
	for _, fn := range pipeline {
		processed, err = fn(d, processed, list)

		if err != nil {
			return
		}
	}

	elapsed := time.Since(starts)
	log.Debugf("postprocess content	took=%v", elapsed)
	return
}

// Preprocess a parseable type.
func Preprocess(d deps, c Parseable) (processed Parseable, err error) {
	starts := time.Now()
	pipeline := []Preprocessor{
		preReplaceMentionTags,
		preReplaceAssetTags,
	}

	// Run pipeline over parseable.
	processed = c
	for _, fn := range pipeline {
		processed, err = fn(d, processed)
		if err != nil {
			return
		}
	}

	elapsed := time.Since(starts)
	log.Debugf("preprocess took = %v", elapsed)
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
