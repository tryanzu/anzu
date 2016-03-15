package posts

var legalSlug = regexp.MustCompile(`^([a-zA-Z0-9\-\.|/]+)$`)

func (this API) Get(c *gin.Context) {

    var kind string
    var post *feed.Post
    var err error

	id := c.Params.ByName("id")

	if bson.IsObjectIdHex(id) {
		kind = "id"
	}

	if legalSlug.MatchString(id) && kind == "" {
		kind = "slug"
	}

	if kind == "" {
		c.JSON(400, gin.H{"message": "Invalid request, id not valid.", "status": "error"})
		return
	}

    if kind == "id" {
        post, err = this.Feed.Post(bson.ObjectIdHex(id))
    } else {
        post, err = this.Feed.Post(bson.M{"slug": id})
    }

    if err != nil {
        c.JSON(404, gin.H{"message": "Couldnt found post with that slug.", "status": "error"})
		return
    }

    // Needed data loading to show post
    post.LoadComments(-10, 0)

    if post.Data().
    post.LoadUsers()
    post.LoadVotes()


	// Get the collection
	collection := database.C("posts")
	post := model.Post{}

	// Try to fetch the needed post by id
	if post_type == "id" {
		err = collection.FindId(bson.ObjectIdHex(id)).Select(bson.M{"comments.set": bson.M{"$slice": -10}}).One(&post)
	}

	if post_type == "slug" {
		err = collection.Find(bson.M{"slug": id}).Select(bson.M{"comments.set": bson.M{"$slice": -10}}).One(&post)
	}

	if err != nil {
		c.JSON(404, gin.H{"message": "Couldnt found post with that slug.", "status": "error"})
		return
	}

	// Get the users and stuff
	if post.Users != nil && len(post.Users) > 0 {

		var users []model.User

		// Get the users
		collection := database.C("users")

		err := collection.Find(bson.M{"_id": bson.M{"$in": post.Users}}).All(&users)

		if err != nil {
			panic(err)
		}

		usersMap := make(map[bson.ObjectId]interface{})

		var description string

		for _, user := range users {
			description = "Solo otro Spartan Geek más"

			if len(user.Description) > 0 {
				description = user.Description
			}

			usersMap[user.Id] = map[string]interface{}{
				"id":          user.Id.Hex(),
				"username":    user.UserName,
				"description": description,
				"image":       user.Image,
				"level":       user.Gaming.Level,
				"roles":       user.Roles,
			}

			if user.Id == post.UserId {
				// Set the author
				post.Author = user
			}
		}

		// Name of the set to get
		_, signed_in := c.Get("token")

		// Look for votes that has been already given
		var votes []model.Vote
		var likes []model.Vote
		var liked model.Vote

		if signed_in {

			user_id := c.MustGet("user_id")
			user_bson_id := bson.ObjectIdHex(user_id.(string))

			err = database.C("votes").Find(bson.M{"type": "component", "related_id": post.Id, "user_id": user_bson_id}).All(&votes)

			// Get the likes given by the current user
			_ = database.C("votes").Find(bson.M{"type": "comment", "related_id": post.Id, "user_id": user_bson_id}).All(&likes)

			err = database.C("votes").Find(bson.M{"type": "post", "related_id": post.Id, "user_id": user_bson_id}).One(&liked)

			if err == nil {

				post.Liked = liked.Value
			}

			// Increase user saw posts and its gamification in another thread
			go func(user_id bson.ObjectId, users []model.User) {

				var target model.User

				// Update the user saw posts
				_ = database.C("users").Update(bson.M{"_id": user_id}, bson.M{"$inc": bson.M{"stats.saw": 1}})
				player := false

				for _, user := range users {

					if user.Id == user_id {

						// The user is a player of the post so we dont have to get it from the database again
						player = true
						target = user
					}
				}

				if player == false {

					err = collection.Find(bson.M{"_id": user_id}).One(&target)

					if err != nil {
						panic(err)
					}
				}

				// Update user achievements (saw posts)
				//updateUserAchievement(target, "saw")

			}(user_bson_id, users)
		}

		if post.Solved == true {

			var best model.CommentAggregated

			pipe := database.C("posts").Pipe([]bson.M{
				{"$match": bson.M{"_id": post.Id}},
				{"$unwind": "$comments.set"},
				{"$match": bson.M{"comments.set.chosen": true}},
				{"$project": bson.M{"_id": 1, "comment": "$comments.set"}},
			})

			err := pipe.One(&best)

			if err != nil {
				panic(err)
			}

			if err == nil {

				answer := best.Comment
				post.Comments.Answer = &answer

				if _, okay := usersMap[answer.UserId]; okay {
					post.Comments.Answer.User = usersMap[answer.UserId]
				}
			}
		}

		// This will calculate the position based on the sliced array
		true_count := di.Feed.TrueCommentCount(post.Id)
		count := true_count - 10

		if count < 0 {
			count = 0
		}

		for index := range post.Comments.Set {

			comment := &post.Comments.Set[index]

			// Save the position over the comment
			post.Comments.Set[index].Position = count + index

			// Check if user liked that comment already
			for _, vote := range likes {

				if vote.NestedType == strconv.Itoa(index) {

					post.Comments.Set[index].Liked = vote.Value
				}
			}

			if _, okay := usersMap[comment.UserId]; okay {

				post.Comments.Set[index].User = usersMap[comment.UserId]
			}
		}

		// Remove deleted comments from the set
		comments := post.Comments.Set[:0]

		for _, c := range post.Comments.Set {

			if c.Deleted.IsZero() == true {

				comments = append(comments, c)
			}
		}

		post.Comments.Set = comments
		post.Comments.Total = true_count

		// Sort by created at
		sort.Sort(model.ByCommentCreatedAt(post.Comments.Set))

		// Get components information if components publication
		components := reflect.ValueOf(&post.Components).Elem()
		components_type := reflect.TypeOf(&post.Components).Elem()

		for i := 0; i < components.NumField(); i++ {

			f := components.Field(i)
			t := components_type.Field(i)

			if f.Type().String() == "model.Component" {

				component := f.Interface().(model.Component)

				for _, vote := range votes {

					if vote.NestedType == strings.ToLower(t.Name) {

						if vote.Value == 1 {

							component.Voted = "up"

						} else if vote.Value == -1 {

							component.Voted = "down"
						}
					}
				}

				if component.Elections == true {

					for option_index, option := range component.Options {

						if _, okay := usersMap[option.UserId]; okay {

							component.Options[option_index].User = usersMap[option.UserId]
						}
					}

					// Sort by created at
					sort.Sort(model.ByElectionsCreatedAt(component.Options))
				}

				f.Set(reflect.ValueOf(component))
			}
		}
	}

	// Save the activity
	signed_id, signed_in := c.Get("user_id")
	user_id := ""

	if signed_in {

		user_id = signed_id.(string)
	}

	go func(post model.Post, user_id string, signed_in bool) {

		defer di.Errors.Recover()

		post_module, _ := di.Feed.Post(post)

		if signed_in {

			by := bson.ObjectIdHex(user_id)

			post_module.Viewed(by)
		}

		post_module.UpdateRate()

		// Trigger gamification events (if needed)
		di.Gaming.Post(post_module).Review()

	}(post, user_id, signed_in)

	c.JSON(200, post)
}
