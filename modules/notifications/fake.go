package notifications

import (
    "github.com/fernandez14/spartangeek-blacker/model"
    "fmt"
)

type FakeBroadcaster struct {
    
}

func (broadcaster FakeBroadcaster) Send(message *model.UserFirebaseNotification) {
    
    fmt.Printf("\n\n%v\n\n", message)
    
    // Used on tests (dont do anything just yet)
    return
}