package content

type Parseable interface {
	GetContent() string
	UpdateContent(string) Parseable
	GetParseableMeta() map[string]interface{}
}
