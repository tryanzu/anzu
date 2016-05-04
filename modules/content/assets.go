package content

var urlsRegexp, _ = regexp.Compile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)

func (self Module) AsyncAssetDownload(o Parseable) bool {

	c := o.GetContent()
	assets := urlsRegexp.FindAllString(c, -1)

	for _, asset := range assets {
		// Do download
	}

	o.OnParseFilterFinished("assets")

	return true
}
