package content

import (
	"regexp"
)

var mention_regex, _ = regexp.Compile(`(?i)\B\@([\w\-]+)(#[0-9]+)*`)

func (self Module) ParseMentionTags(o Parseable) error {

}

func (self Module) ParseTags(o Parseable) error {

	chain := []func(Parseable) bool{
		self.ParseMentionTags,
		self.ParseContentMentions,
	}

	for _, fn := range chain {
		next := fn(o)

		if !next {
			break
		}
	}

	o.OnParseFinished()

	return nil
}
