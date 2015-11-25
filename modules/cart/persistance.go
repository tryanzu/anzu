package cart

type Persistance interface {
	Restore() error
	Save() error
}

