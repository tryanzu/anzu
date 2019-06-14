package config

import (
	"github.com/dop251/goja"
)

// BanReason config def.
type BanReason struct {
	Code string `hcl:"effects"`
}

// BanEffects def.
type BanEffects struct {
	Duration  int64
	IPAddress bool
}

// Effects from config (duration, ipaddress).
func (re BanReason) Effects(times int) (BanEffects, error) {
	if len(re.Code) == 0 {
		return BanEffects{60, false}, nil
	}
	vm := goja.New()
	vm.RunString(`
		var exports = {};
	`)
	vm.Set("banN", times)
	if _, err := vm.RunString(re.Code); err != nil {
		return BanEffects{}, err
	}
	obj := vm.Get("exports").ToObject(vm)
	duration := obj.Get("duration").ToInteger()
	ip := obj.Get("ip").ToBoolean()

	return BanEffects{duration, ip}, nil
}
