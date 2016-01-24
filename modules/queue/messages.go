package queue

import (
	"github.com/iron-io/iron_go3/mq"
	
	"encoding/json"
)

func Push(name, event string, params map[string]interface{}) error {
    
    message := map[string]interface{}{
        "fire": event,
    }
    
    for key, value := range params {
        message[key] = value
    }
    
    q := mq.New(name)
    
    data, err := json.Marshal(message)
    
    if err != nil {
        return err
    }
    
	_, err = q.PushMessage(mq.Message{Body: string(data)})
	
	return err
}

func PushWDelay(name, event string, params map[string]interface{}, delay int) error {
    
    message := map[string]interface{}{
        "fire": event,
    }
    
    for key, value := range params {
        message[key] = value
    }
    
    q := mq.New(name)
    
    data, err := json.Marshal(message)
    
    if err != nil {
        return err
    }
    
	_, err = q.PushMessage(mq.Message{Body: string(data), Delay: int64(delay)})
	
	return err
}