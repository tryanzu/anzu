package notifications

import (
	"fmt"
	"github.com/fernandez14/spartangeek-blacker/board/legacy/model"
)

type FakeBroadcaster struct {
}

func (broadcaster FakeBroadcaster) Send(message *model.UserFirebaseNotification) {

	fmt.Printf("\n\n%v\n\n", message)

	// Used on tests (dont do anything just yet)
	return
}
