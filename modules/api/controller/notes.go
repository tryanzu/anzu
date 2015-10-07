package controller

import (
	"github.com/fernandez14/spartangeek-blacker/modules/store"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

type BuildNotesAPI struct {
	Store *store.Module `inject:""`
}

func (self BuildNotesAPI) All(c *gin.Context) {

	notes := self.Store.GetNotes()

	c.JSON(200, notes)
}

// REST handler for getting one build note
func (self BuildNotesAPI) One(c *gin.Context) {

	id := c.Param("id")

	if bson.IsObjectIdHex(id) == false {

		c.JSON(400, gin.H{"status": "error", "message": "Can't perform action. Invalid id."})
		return
	}

	note, err := self.Store.GetNote(bson.ObjectIdHex(id))

	if err != nil {

		c.JSON(404, gin.H{"status": "error", "message": "Build note not found."})
		return
	}

	c.JSON(200, note)
}

// REST handler for creating build notes
func (self BuildNotesAPI) Create(c *gin.Context) {

	var form BuildResponseForm

	if c.BindJSON(&form) == nil {

		err := self.Store.CreateNote(form.Title, form.Content)

		if err != nil {

			c.JSON(400, gin.H{"status": "error"})
			return
		}

		c.JSON(200, gin.H{"status": "okay"})
	}
}

// REST handler for updating build notes
func (self BuildNotesAPI) Update(c *gin.Context) {

	id := c.Param("id")

	if bson.IsObjectIdHex(id) == false {

		c.JSON(400, gin.H{"status": "error", "message": "Can't perform action. Invalid id."})
		return
	}

	var form BuildResponseForm

	if c.BindJSON(&form) == nil {

		err := self.Store.UpdateNote(bson.ObjectIdHex(id), form.Title, form.Content)

		if err != nil {

			c.JSON(400, gin.H{"status": "error"})
			return
		}

		c.JSON(200, gin.H{"status": "okay"})
	}
}

// REST handler for deleting build notes
func (self BuildNotesAPI) Delete(c *gin.Context) {

	id := c.Param("id")

	if bson.IsObjectIdHex(id) == false {

		c.JSON(400, gin.H{"status": "error", "message": "Can't perform action. Invalid id."})
		return
	}

	err := self.Store.DeleteNote(bson.ObjectIdHex(id))
	
	if err != nil {

		c.JSON(400, gin.H{"status": "error"})
		return
	}

	c.JSON(200, gin.H{"status": "okay"})
}

type BuildResponseForm struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}