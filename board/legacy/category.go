package handle

import (
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/board/legacy/model"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/acl"
	"gopkg.in/mgo.v2/bson"
)

type CategoryAPI struct {
	Acl *acl.Module `inject:""`
}

func (di *CategoryAPI) CategoriesGet(c *gin.Context) {
	var categories []model.Category
	err := deps.Container.Mgo().C("categories").Find(nil).All(&categories)
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

	// Check the categories the user can write if parameter is provided
	perms := c.Query("permissions")

	if _, authenticated := c.Get("token"); authenticated {
		userID := c.MustGet("userID").(bson.ObjectId)
		if perms == "write" {
			user := di.Acl.User(userID)

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
			for idx, category := range parent_categories {
				if len(category.Child) > 0 {
					continue
				}
				if len(parent_categories) > idx+1 {
					parent_categories = append(parent_categories[:idx], parent_categories[idx+1:]...)
				} else {
					parent_categories = parent_categories[:len(parent_categories)-1]
				}
			}
		}
	}

	sort.Sort(model.CategoriesOrder(parent_categories))
	c.JSON(200, parent_categories)
}
