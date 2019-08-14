package realtime

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

func countClientsWorker() {
	channels := map[string]map[*Client]struct{}{}
	changes := 0
	for {
		select {
		case client := <-counters:
			if client.Channels == nil {
				for name := range channels {
					delete(channels[name], client)
				}
				continue
			}
			client.Channels.Range(func(k, v interface{}) bool {
				name := k.(string)
				if _, exists := channels[name]; !exists {
					channels[name] = map[*Client]struct{}{}
				}
				channels[name][client] = struct{}{}
				return true
			})
			for name := range channels {
				if _, exists := client.Channels.Load(name); !exists {
					delete(channels[name], client)
				}
			}
			changes++
		case <-time.After(time.Millisecond * 500):
			if changes == 0 {
				continue
			}
			counters := make(map[string]interface{}, len(channels))
			unique := map[bson.ObjectId]struct{}{}
			peers := [][2]string{}
			for name, listeners := range channels {
				counters[name] = len(listeners)

				// Calculate the list of connected peers on the counters channel
				if name != "chat:counters" {
					continue
				}
				for client := range listeners {
					if client.User == nil {
						continue
					}
					if _, exists := unique[client.User.Id]; exists {
						continue
					}
					id := client.User.Id.Hex()
					name := client.User.UserName
					peers = append(peers, [2]string{id, name})
					unique[client.User.Id] = struct{}{}
				}
			}

			m := M{
				Channel: "chat:counters",
				Content: SocketEvent{
					Event: "update",
					Params: map[string]interface{}{
						"channels": counters,
						"peers":    peers,
					},
				}.encode(),
			}
			changes = 0
			ToChan <- m
		}
	}
}
