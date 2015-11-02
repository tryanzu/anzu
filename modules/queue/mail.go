package queue 

import (
	"log"
)

type MailJob struct {
	
}

func (self MailJob) StoreDelayedResponse(params map[string]interface{}) {
		
	log.Println("Done from StoreDelayedResponse")
}