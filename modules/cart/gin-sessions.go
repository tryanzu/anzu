package cart

import (
	"encoding/json"
	"github.com/gin-gonic/contrib/sessions"
	"fmt"
)

type GinGonicSession struct {
	Session sessions.Session
}

func (gcs GinGonicSession) Restore(where interface{}) error {

	session := gcs.Session
	data := session.Get("cart")

	if data == nil {

		return nil
	} else {

		encoded := data.(string)

		if err := json.Unmarshal([]byte(encoded), &where); err != nil {

			return err
		}

		return nil
	}
}

func (gcs GinGonicSession) Save(data interface{}) error {

	encoded, err := json.Marshal(data)

	if err != nil {
		return err
	}

	session := gcs.Session

	session.Set("cart", string(encoded))
	session.Save()

	fmt.Printf("%v\n", session.Get("cart").(string))

	return nil
}
