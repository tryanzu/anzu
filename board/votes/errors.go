package votes

// NotAllowed check.
type NotAllowed struct {
	Reason string
}

func (e *NotAllowed) Error() string {
	return "can't allow to perform operation"
}
