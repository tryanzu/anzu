package feed

import (
    "github.com/fernandez14/spartangeek-blacker/model"
    "gopkg.in/mgo.v2/bson"
    "strconv"
)

func (di *FeedModule) UpdateFeedRates(list []model.FeedPost) {

    // Recover from any panic even inside this isolated process
    defer di.Errors.Recover()

    // Services we will need along the runtime
    redis := di.CacheService

    // Sorted list items (redis ZADD)
    zadd := make(map[string]map[string]float64)

    for _, post := range list {

        var reached, viewed int

        // Get reach and views 
        reached, viewed = di.getPostReachViews(post.Id)

        total := reached + viewed

        if total > 101 {
            
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
    redis := di.CacheService

    // Sorted list items (redis ZADD)
    zadd := make(map[string]float64)

    // Get reach and views 
    reached, viewed := di.getPostReachViews(post.Id)

    total := reached + viewed

    if total > 101 {
        
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
}

func (di *FeedModule) getPostReachViews(id bson.ObjectId) (int, int) {

    var reached, viewed int

    // Services we will need along the runtime
    database := di.Mongo.Database
    redis := di.CacheService

    list_count, _ := redis.Get("feed:count:list:" + id.Hex())

    if list_count == nil {
        
        reached, _ = database.C("activity").Find(bson.M{"list": id, "event": "feed"}).Count()    
        err := redis.Set("feed:count:list:" + id.Hex(), strconv.Itoa(reached), 1800, 0, false, false)

        if err != nil {
            panic(err)
        }  
    } else {

        reached, _ = strconv.Atoi(string(list_count))
    }

    viewed_count, _ := redis.Get("feed:count:post:" + id.Hex())

    if viewed_count == nil {

        viewed, _ = database.C("activity").Find(bson.M{"related_id": id, "event": "post"}).Count()
        err := redis.Set("feed:count:post:" + id.Hex(), strconv.Itoa(viewed), 1800, 0, false, false)

        if err != nil {
            panic(err)
        }
    } else {

        viewed, _ = strconv.Atoi(string(viewed_count))
    }

    return reached, viewed
}
