package feed

import (
    "github.com/fernandez14/spartangeek-blacker/model"
    "gopkg.in/mgo.v2/bson"
)

func (di *FeedModule) UpdateFeedRates(list []model.FeedPost) {

    // Recover from any panic even inside this isolated process
    defer di.Errors.Recover()

    // Services we will need along the runtime
    database := di.Mongo.Database
    redis := di.CacheService

    // Sorted list items (redis ZADD)
    zadd := make(map[string]map[string]float64)

    for _, post := range list {

        reached, _ := database.C("activity").Find(bson.M{"list": post.Id, "event": "feed"}).Count()         
        viewed, _ := database.C("activity").Find(bson.M{"related_id": post.Id, "event": "post"}).Count()

        // Calculate the rates
        view_rate    := 100.0 / float64(reached) * float64(viewed)
        comment_rate := 100.0 / float64(viewed) * float64(post.Comments.Count)
        final_rate   := (view_rate + comment_rate) / 2.0
        date := post.Created.Format("2006-01-02")

        if _, okay := zadd[date]; !okay {

            zadd[date] = map[string]float64{}
        }

        zadd[date][post.Id.Hex()] = final_rate
    }

    for date, items := range zadd {

        _, err := redis.ZAdd("feed:relevant:" + date, items)

        if err != nil {
            panic(err)
        }
    }
}

func (di *FeedModule) UpdatePostRate(post model.Post) {

    // Recover from any panic even inside this isolated process
    defer di.Errors.Recover()

    // Services we will need along the runtime
    database := di.Mongo.Database
    redis := di.CacheService

    // Sorted list items (redis ZADD)
    zadd := make(map[string]float64)

    reached, _ := database.C("activity").Find(bson.M{"list": post.Id, "event": "feed"}).Count()         
    viewed, _ := database.C("activity").Find(bson.M{"related_id": post.Id, "event": "post"}).Count()

    // Calculate the rates
    view_rate    := 100.0 / float64(reached) * float64(viewed)
    comment_rate := 100.0 / float64(viewed) * float64(post.Comments.Count)
    final_rate   := (view_rate + comment_rate) / 2.0
    date := post.Created.Format("2006-01-02")

    zadd[post.Id.Hex()] = final_rate
    
    _, err := redis.ZAdd("feed:relevant:" + date, zadd)

    if err != nil {
        panic(err)
    }   
}
