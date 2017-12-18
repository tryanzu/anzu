package feed

import (
	"github.com/tryanzu/core/modules/user"
	"gopkg.in/mgo.v2/bson"

	"time"
)

type MModelVotes struct {
	Up     int `bson:"up" json:"up"`
	Down   int `bson:"down" json:"down"`
	Rating int `bson:"rating,omitempty" json:"rating,omitempty"`
}

type MModelAuthor struct {
	Id      bson.ObjectId `bson:"id,omitempty" json:"id,omitempty"`
	Name    string        `bson:"name" json:"name"`
	Email   string        `bson:"email" json:"email"`
	Avatar  string        `bson:"avatar" json:"avatar"`
	Profile interface{}   `bson:"profile,omitempty" json:"profile,omitempty"`
}

type MModelComments struct {
	Count  int             `bson:"count" json:"count"`
	Total  int             `bson:"-" json:"total"`
	Answer *MModelComment  `bson:"-" json:"answer,omitempty"`
	Set    []MModelComment `bson:"set" json:"set"`
}

type MModelFeedComments struct {
	Count int `bson:"count" json:"count"`
}

type MModelComment struct {
	UserId   bson.ObjectId `bson:"user_id" json:"user_id"`
	Votes    MModelVotes   `bson:"votes" json:"votes"`
	User     interface{}   `bson:"-" json:"author,omitempty"`
	Position int           `bson:"position" json:"position"`
	Liked    int           `bson:"-" json:"liked,omitempty"`
	Content  string        `bson:"content" json:"content"`
	Chosen   bool          `bson:"chosen,omitempty" json:"chosen,omitempty"`
	Created  time.Time     `bson:"created_at" json:"created_at"`
	Deleted  time.Time     `bson:"deleted_at" json:"deleted_at"`
}

type MModelComponents struct {
	Cpu               MModelComponent `bson:"cpu,omitempty" json:"cpu,omitempty"`
	Motherboard       MModelComponent `bson:"motherboard,omitempty" json:"motherboard,omitempty"`
	Ram               MModelComponent `bson:"ram,omitempty" json:"ram,omitempty"`
	Storage           MModelComponent `bson:"storage,omitempty" json:"storage,omitempty"`
	Cooler            MModelComponent `bson:"cooler,omitempty" json:"cooler,omitempty"`
	Power             MModelComponent `bson:"power,omitempty" json:"power,omitempty"`
	Cabinet           MModelComponent `bson:"cabinet,omitempty" json:"cabinet,omitempty"`
	Screen            MModelComponent `bson:"screen,omitempty" json:"screen,omitempty"`
	Videocard         MModelComponent `bson:"videocard,omitempty" json:"videocard,omitempty"`
	Software          string          `bson:"software,omitempty" json:"software,omitempty"`
	Budget            string          `bson:"budget,omitempty" json:"budget,omitempty"`
	BudgetCurrency    string          `bson:"budget_currency,omitempty" json:"budget_currency,omitempty"`
	BudgetType        string          `bson:"budget_type,omitempty" json:"budget_type,omitempty"`
	BudgetFlexibility string          `bson:"budget_flexibility,omitempty" json:"budget_flexibility,omitempty"`
}

type MModelComponent struct {
	Content string      `bson:"content" json:"content"`
	Votes   MModelVotes `bson:"votes" json:"votes"`
	Status  string      `bson:"status" json:"status"`
	Voted   string      `bson:"voted,omitempty" json:"voted,omitempty"`
}

type MModelPost struct {
	Id                bson.ObjectId    `bson:"_id,omitempty" json:"id,omitempty"`
	Title             string           `bson:"title" json:"title"`
	Slug              string           `bson:"slug" json:"slug"`
	Type              string           `bson:"type" json:"type"`
	Content           string           `bson:"content" json:"content"`
	Categories        []string         `bson:"categories" json:"categories"`
	Category          bson.ObjectId    `bson:"category" json:"category"`
	Comments          MModelComments   `bson:"comments" json:"comments"`
	Author            user.UserSimple  `bson:"-" json:"author,omitempty"`
	UserId            bson.ObjectId    `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Users             []bson.ObjectId  `bson:"users,omitempty" json:"users,omitempty"`
	Votes             MModelVotes      `bson:"votes" json:"votes"`
	Components        MModelComponents `bson:"components,omitempty" json:"components,omitempty"`
	RelatedComponents []bson.ObjectId  `bson:"related_components,omitempty" json:"related_components,omitempty"`
	Following         bool             `bson:"following,omitempty" json:"following,omitempty"`
	Pinned            bool             `bson:"pinned,omitempty" json:"pinned,omitempty"`
	Lock              bool             `bson:"lock" json:"lock"`
	IsQuestion        bool             `bson:"is_question" json:"is_question"`
	Solved            bool             `bson:"solved,omitempty" json:"solved,omitempty"`
	Liked             int              `bson:"liked,omitempty" json:"liked,omitempty"`
	Created           time.Time        `bson:"created_at" json:"created_at"`
	Updated           time.Time        `bson:"updated_at" json:"updated_at"`
	Deleted           time.Time        `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

type MModelPostCommentModel struct {
	Id      bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	Title   string        `bson:"title" json:"title"`
	Slug    string        `bson:"slug" json:"slug"`
	Comment MModelComment `bson:"comment" json:"comment"`
}

type MModelPostCommentCountModel struct {
	Id    bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	Count int           `bson:"count" json:"count"`
}

type MModelCommentAggregated struct {
	Id      bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	Comment MModelComment `bson:"comment" json:"comment"`
}

type MModelCommentsPost struct {
	Id       bson.ObjectId  `bson:"_id,omitempty" json:"id,omitempty"`
	Comments MModelComments `bson:"comments" json:"comments"`
}

type MModelFeedPost struct {
	Id         bson.ObjectId      `bson:"_id,omitempty" json:"id,omitempty"`
	Title      string             `bson:"title" json:"title"`
	Slug       string             `bson:"slug" json:"slug"`
	Type       string             `bson:"type" json:"type"`
	Categories []string           `bson:"categories" json:"categories"`
	Users      []bson.ObjectId    `bson:"users,omitempty" json:"users,omitempty"`
	Category   bson.ObjectId      `bson:"category" json:"category"`
	Comments   MModelFeedComments `bson:"comments" json:"comments"`
	Author     user.UserSimple    `bson:"author,omitempty" json:"author,omitempty"`
	UserId     bson.ObjectId      `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Votes      MModelVotes        `bson:"votes" json:"votes"`
	Pinned     bool               `bson:"pinned,omitempty" json:"pinned,omitempty"`
	Solved     bool               `bson:"solved,omitempty" json:"solved,omitempty"`
	IsQuestion bool               `bson:"is_question" json:"is_question"`
	Stats      MModelFeedPostStat `bson:"stats,omitempty" json:"stats"`
	Created    time.Time          `bson:"created_at" json:"created_at"`
	Updated    time.Time          `bson:"updated_at" json:"updated_at"`
}

type MModelFeedPostStat struct {
	Viewed      int     `bson:"viewed,omitempty" json:"viewed"`
	Reached     int     `bson:"reached,omitempty" json:"reached"`
	ViewRate    float64 `bson:"view_rate,omitempty" json:"view_rate"`
	CommentRate float64 `bson:"comment_rate,omitempty" json:"comment_rate"`
	FinalRate   float64 `bson:"final_rate,omitempty" json:"final_rate"`
}

type MModelPostForm struct {
	Kind       string                 `json:"kind" binding:"required"`
	Name       string                 `json:"name" binding:"required"`
	Content    string                 `json:"content" binding:"required"`
	Budget     string                 `json:"budget"`
	Currency   string                 `json:"currency"`
	Moves      string                 `json:"moves"`
	Software   string                 `json:"software"`
	Tag        string                 `json:"tag"`
	Category   string                 `json:"category"`
	IsQuestion bool                   `json:"is_question"`
	Pinned     bool                   `json:"pinned"`
	Lock       bool                   `json:"lock"`
	Components map[string]interface{} `json:"components"`
}

// ByCommentCreatedAt implements sort.Interface for []ElectionOption based on Created field
type ByCommentCreatedAt []MModelComment

func (a ByCommentCreatedAt) Len() int           { return len(a) }
func (a ByCommentCreatedAt) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCommentCreatedAt) Less(i, j int) bool { return !a[i].Created.Before(a[j].Created) }

type ByBestRated []MModelFeedPost

func (slice ByBestRated) Len() int {
	return len(slice)
}

func (slice ByBestRated) Less(i, j int) bool {
	return slice[i].Stats.FinalRate > slice[j].Stats.FinalRate
}

func (slice ByBestRated) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
