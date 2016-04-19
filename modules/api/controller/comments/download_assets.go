package comments

func (this API) saveCommentImages(from string, post_id bson.ObjectId) error {

	// Recover from any panic even inside this goroutine
	defer di.Errors.Recover()

	// Get the database interface from the DI
	database := di.DataService.Database
	amazon_url, err := di.ConfigService.String("amazon.url")

	if err != nil {
		panic(err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// Download the file
	response, err := client.Get(from)
	if err != nil {
		return errors.New(fmt.Sprint("Error while downloading", from, "-", err))
	}

	// Read all the bytes to the image
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return errors.New(fmt.Sprint("Error while downloading", from, "-", err))
	}

	// Detect the downloaded file type
	dataType := http.DetectContentType(data)

	if dataType[0:5] == "image" {

		var extension, name string

		// Parse the filename
		u, err := url.Parse(from)

		if err != nil {
			return errors.New(fmt.Sprint("Error while parsing url", from, "-", err))
		}

		extension = filepath.Ext(u.Path)
		name = bson.NewObjectId().Hex()

		if extension != "" {

			name = name + extension
		} else {

			// If no extension is provided on the url then add a dummy one
			name = name + ".jpg"
		}

		path := "posts/" + name
		err = di.S3Bucket.Put(path, data, dataType, s3.ACL("public-read"))

		if err != nil {
			panic(err)
		}

		var post model.Post

		err = database.C("posts").Find(bson.M{"_id": post_id}).One(&post)

		if err == nil {

			for index := range post.Comments.Set {

				comment := post.Comments.Set[index].Content

				// Replace the url on the comment
				if strings.Contains(comment, from) {

					var rem bytes.Buffer

					// Make the push string
					rem.WriteString("comments.set.")
					rem.WriteString(strconv.Itoa(index))
					rem.WriteString(".content")

					ctc := rem.String()

					content := strings.Replace(comment, from, amazon_url+path, -1)

					// Update the comment
					di.DataService.Database.C("posts").Update(bson.M{"_id": post_id}, bson.M{"$set": bson.M{ctc: content}})
				}
			}
		}
	}

	response.Body.Close()

	return nil
}
