package config

import (
	"github.com/dop251/goja"
)

type ReactionEffect struct {
	Code string `hcl:"exec"`
}

type Rewards struct {
	Provider int64
	Receiver int64
}

// Rewards is a way to transform the reaction code from system config (javascript definition) into a reward struct.
func (re ReactionEffect) Rewards() (Rewards, error) {
	vm := goja.New()
	vm.RunString(`
		var exports = {};
	`)
	if _, err := vm.RunString(re.Code); err != nil {
		return Rewards{}, err
	}
	obj := vm.Get("exports").ToObject(vm)
	provider := obj.Get("provider").ToInteger()
	receiver := obj.Get("receiver").ToInteger()
	return Rewards{provider, receiver}, nil
}
