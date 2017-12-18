package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/board/legacy/model"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/acl"
	"gopkg.in/mgo.v2/bson"
	"sort"
)

type CategoryAPI struct {
	Acl *acl.Module `inject:""`
}

func (di *CategoryAPI) CategoriesGet(c *gin.Context) {

	// Get the database interface from the DI
	database := deps.Container.Mgo()

	var categories []model.Category

	// Get the categories collection to perform there
	collection := database.C("categories")

	err := collection.Find(nil).All(&categories)

	if err != nil {
		panic(err)
	}

	// Filter the parent categories without allocating
	parent_categories := []model.Category{}
	child_categories := []model.Category{}

	for _, category := range categories {

		if category.Parent.Hex() == "" {

			parent_categories = append(parent_categories, category)
		} else {

			child_categories = append(child_categories, category)
		}
	}

	for category_index, category := range parent_categories {

		for _, child := range child_categories {

			if child.Parent == category.Id {

				parent_categories[category_index].Child = append(parent_categories[category_index].Child, child)
			}
		}
	}

	// Check whether auth or not
	_, signed_in := c.Get("token")

	// Check the categories the user can write if parameter is provided
	perms := c.Query("permissions")

	if signed_in {

		user_id := c.MustGet("user_id").(string)

		if perms == "write" {

			user := di.Acl.User(bson.ObjectIdHex(user_id))

			// Remove child categories with no write permissions
			for category_index, _ := range parent_categories {

				for child_index, child := range parent_categories[category_index].Child {

					if user.CanWrite(child) == false {

						if len(parent_categories[category_index].Child) > child_index+1 {

							parent_categories[category_index].Child = append(parent_categories[category_index].Child[:child_index], parent_categories[category_index].Child[child_index+1:]...)
						} else {

							parent_categories[category_index].Child = parent_categories[category_index].Child[:len(parent_categories[category_index].Child)-1]
						}
					}
				}
			}

			// Clean up parent categories with no subcategories
			for category_index, category := range parent_categories {

				//log.Printf("%v has %v \n", category.Slug, len(category.Child))

				if len(category.Child) == 0 {

					if len(parent_categories) > category_index+1 {

						parent_categories = append(parent_categories[:category_index], parent_categories[category_index+1:]...)
					} else {

						parent_categories = parent_categories[:len(parent_categories)-1]
					}
				}
			}
		}
	}

	sort.Sort(model.CategoriesOrder(parent_categories))

	c.JSON(200, parent_categories)
}
