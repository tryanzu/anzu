package main

import (
    "encoding/json"
    "reflect"
    "strconv"
    "gopkg.in/mgo.v2/bson"
)

type UserAchievements struct {
    Saw int        `json:"saw"`
    Posts int      `json:"posts"`
    Comments int   `json:"comments"`
    GivenVotes int `json:"given_votes"`
}

type Level struct {
    Id bson.ObjectId `bson:"_id,omitempty" json:"id"`
    Step    int `bson:"step" json:"step"`
    Slug    string `bson:"slug" json:"slug"`
    Name    string `bson:"name" json:"name"`
    Description string `bson:"description" json:"description"`
    Badges  []string `bson:"badges" json:"badges"`
}

type Badge struct {
    Id bson.ObjectId `bson:"_id,omitempty" json:"id"`
    Slug    string `bson:"slug" json:"slug"`
    Icon    string `bson:"icon" json:"icon"`
    Name    string `bson:"name" json:"name"`
    Description string `bson:"description" json:"description"`
    Requirements  map[string]interface{} `bson:"req" json:"requirements"`
}

func warmUpUserAchievements (user User) UserAchievements {

    id := user.Id

    var saw int
    var posts int
    var comments int
    var givenvotes int

    saw = user.Stats.Saw

    // Get the count of owned posts
    posts, err := database.C("posts").Find(bson.M{"user_id": id}).Count()

    if err != nil {
        panic(err)
    }

    // Get the count of user comments
    var posts_in []Post
    err = database.C("posts").Find(bson.M{"users": id}).All(&posts_in)

    if err != nil {
        panic(err)
    }

    comments = 0
    for _, post := range posts_in {

        for _, comment := range post.Comments.Set {

            if comment.UserId == id {
                comments = comments + 1
            }
        }
    }

    // Get the count of user given votes
    givenvotes, err = database.C("votes").Find(bson.M{"user_id": id}).Count()

    if err != nil {
        panic(err)
    }

    var userCache string

    heat := UserAchievements{
        Saw: saw,
        Posts: posts,
        Comments: comments,
        GivenVotes: givenvotes,
    }

    encoded, err := json.Marshal(heat)

    if err != nil {
        panic(err)
    }

    userCache = string(encoded)
    store    := "achievements." + id.Hex()
    err = redis.Set(store, userCache, 43200, 0, false, false)

    if err != nil {
        panic(err)
    }

    return heat
}

func getUserAchievements (user User) UserAchievements {

    id     := user.Id.Hex()
    stored := "achievements." + id

    warmup, _ := redis.Get(stored)
    var cached UserAchievements

    if warmup == nil {

        // Warm up the user achievements in cache
        cached = warmUpUserAchievements(user)

    } else {

        // Unmarshal already warmed up user achievements
        if err := json.Unmarshal(warmup, &cached); err != nil {
            panic(err)
        }
    }

    return cached
}

func updateUserAchievement (user User, achievement string) {

    id := user.Id.Hex()
    achievements := getUserAchievements(user)
    accomplished := reflect.ValueOf(&achievements).Elem()

    for i := 0; i < accomplished.NumField(); i++ {

        t := accomplished.Type().Field(i)
        json_tag := t.Tag
        name := json_tag.Get("json")

        if name == achievement {

            count := accomplished.Field(i).Int()

            // Increase by one the counter
            accomplished.Field(i).SetInt(count + 1)
        }
    }

    encoded, err := json.Marshal(achievements)

    if err != nil {
        panic(err)
    }

    userCache := string(encoded)
    store    := "achievements." + id
    err = redis.Set(store, userCache, 43200, 0, false, false)

    if err != nil {
        panic(err)
    }

    // Detect the user possible level ups on another thread
    go detectUserRecentAchievements(user)
}

func detectUserRecentAchievements (user User) {

    
}

func getStepNeededAchievements (step int) []Badge {

    cache_path := "achievements.steps." + strconv.Itoa(step)
    cached, err := redis.Get(cache_path)

    if cached == nil {

        var level Level

        err = database.C("levels").Find(bson.M{"step": step}).One(&level)

        if err != nil {
            panic(err)
        }

        var badges []Badge

        err = database.C("badges").Find(bson.M{"slug": bson.M{"$in": level.Badges}}).All(&badges)

        if err != nil {
            panic(err)
        }

        encoded, err := json.Marshal(badges)

        if err != nil {
            panic(err)
        }

        stepCache := string(encoded)
        err = redis.Set(cache_path, stepCache, 0, 0, false, false)

        if err != nil {
            panic(err)
        }

        return badges

    } else {

        var badges []Badge

        // Unmarshal already warmed up step achievements
        if err := json.Unmarshal(cached, &badges); err != nil {
            panic(err)
        }

        return badges
    }
}