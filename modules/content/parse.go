package content

func (self Module) Parse(o Parseable) error {

	chain := []func(Parseable) bool{
		self.AsyncAssetDownload,
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
