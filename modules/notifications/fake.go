package notifications

import (
	"fmt"
	"github.com/tryanzu/core/board/legacy/model"
)

type FakeBroadcaster struct {
}

func (broadcaster FakeBroadcaster) Send(message *model.UserFirebaseNotification) {

	fmt.Printf("\n\n%v\n\n", message)

	// Used on tests (dont do anything just yet)
	return
}
