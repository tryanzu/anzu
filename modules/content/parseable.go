package content

type Parseable interface {
	GetContent() string
	UpdateContent(string) bool
	OnParseFilterFinished(string) bool
	OnParseFinished() bool
	GetParseableMeta() map[string]interface{}
}
