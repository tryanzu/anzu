package store

import (
	"bytes"
	"errors"
	"html/template"
	"sort"
	"strings"
	"time"

	"github.com/fernandez14/spartangeek-blacker/deps"
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var InvalidLeadAnswer = errors.New("Invalid lead answer type.")

type Lead struct {
	Id         bson.ObjectId   `bson:"_id,omitempty" json:"id,omitempty"`
	User       OrderUserModel  `bson:"user" json:"user"`
	Content    string          `bson:"content" json:"content"`
	Budget     int             `bson:"budget" json:"budget"`
	Currency   string          `bson:"currency" json:"currency"`
	State      string          `bson:"state" json:"state"`
	Usage      string          `bson:"usage" json:"usage"`
	Games      []string        `bson:"games" json:"games"`
	Extra      []string        `bson:"extras" json:"extra"`
	BuyDelay   int             `bson:"buydelay" json:"buydelay"`
	Readed     []bson.ObjectId `bson:"readed" json:"-"`
	UserReaded bool            `bson:"-" json:"readed"`
	Messages   Messages        `bson:"messages,omitempty" json:"messages"`
	Tags       []TagModel      `bson:"tags,omitempty" json:"tags"`
	Activities []ActivityModel `bson:"activities,omitempty" json:"activities"`
	Pipeline   PipelineModel   `bson:"pipeline,omitempty" json:"pipeline"`
	Trusted    bool            `bson:"trusted_flag" json:"trusted_flag"`
	Favorite   bool            `bson:"favorite_flag" json:"favorite_flag"`
	Lead       bool            `bson:"-" json:"lead"`
	Created    time.Time       `bson:"created_at" json:"created_at"`
	Updated    time.Time       `bson:"updated_at" json:"updated_at"`

	// Runtime generated and not persisted in database
	RelatedUsers interface{}  `bson:"-" json:"related_users,omitempty"`
	Duplicates   []OrderModel `bson:"-" json:"duplicates,omitempty"`
	Invoice      *Invoice     `bson:"-" json:"invoice,omitempty"`

	readed    bool
	readIndex int64
}

func (lead Lead) HadRead(userId bson.ObjectId) Lead {
	for _, r := range lead.Readed {
		if r == userId {
			lead.UserReaded = true
			break
		}
	}

	return lead
}

type Leads []Lead

func (list Leads) ToMap() map[string]Lead {
	m := map[string]Lead{}
	for _, item := range list {
		m[item.Id.Hex()] = item
	}

	return m
}

func (list Leads) IDList() []string {
	keys := make([]string, len(list))
	for k, v := range list {
		keys[k] = v.Id.Hex()
	}

	return keys
}

func (list Leads) HadRead(userId bson.ObjectId) Leads {
	for k, v := range list {
		list[k] = v.HadRead(userId)
	}

	return list
}

// Reply logic over a lead.
func (lead *Lead) Reply(answer, kind string) (string, error) {
	db := deps.Container.Mgo()
	mailer := deps.Container.Mailer()

	if kind != "text" && kind != "note" {
		return "", InvalidLeadAnswer
	}

	var id, subject string
	if kind == "text" {
		subject = "PC Spartana"

		if len(lead.Messages) > 0 {
			subject = "RE: " + subject
		}

		sort.Sort(lead.Messages)
		data := struct {
			Reply string
			Lead  *Lead
		}{
			answer,
			lead,
		}

		funcs := template.FuncMap{
			"trust": func(html string) template.HTML {
				return template.HTML(html)
			},
			"nl2br": func(html template.HTML) template.HTML {
				return template.HTML(strings.Replace(string(html), "\n", "<br>", -1))
			},
		}

		buf := new(bytes.Buffer)
		t := template.New("lead-reply").Funcs(funcs)
		t, _ = t.Parse(`
            {{ .Reply | trust | nl2br }}
			<br><br><br><br>
			<div class="gmail_quote">
			{{with .Lead -}}
				{{ range .Messages -}}
					{{ if eq .Type "inbound" -}} 
						El {{ .Created.Format "02/01/2006, 03:04:05 PM" }}, {{ $.Lead.User.Name }} &lt;{{ $.Lead.User.Email }}&gt; escribió:<br><br>
						<blockquote class="gmail_quote" style="margin:0 0 0 .8ex;border-left:1px #ccc solid;padding-left:1ex">
							<div>
							{{ .Content | trust | nl2br }}<br><br>
					{{ else }}
						El {{ .Created.Format "02/01/2006, 03:04:05 PM" }}, Drak Spartan &lt;pedidos@spartangeek.com&gt; escribió:<br><br>
						<blockquote class="gmail_quote" style="margin:0 0 0 .8ex;border-left:1px #ccc solid;padding-left:1ex">
							<div>
							{{ .Content | trust | nl2br }}<br><br>
					{{- end }}
				{{- end }}

				{{ range .Messages -}}
						</div>
					</blockquote>
				{{- end }}
			{{- end }}
			</div>
        `)
		err := t.Execute(buf, data)
		if err != nil {
			panic(err)
		}

		compose := mail.Raw{
			mail.MailBase{
				Subject: subject,
				Recipient: []mail.MailRecipient{
					{
						Name:  lead.User.Name,
						Email: lead.User.Email,
					},
				},
				FromEmail: "pc@spartangeek.com",
				FromName:  "Drak Spartan",
				Variables: map[string]interface{}{
					"content": answer,
					"subject": subject,
				},
			},
			buf,
		}

		id = mailer.SendRaw(compose)
	}

	message := MessageModel{
		Content:   answer,
		Type:      kind,
		MessageID: id,
		Created:   time.Now(),
		Updated:   time.Now(),
	}

	err := db.C("orders").Update(
		bson.M{"_id": lead.Id},
		bson.M{
			"$push": bson.M{"messages": message},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)

	if err != nil {
		return "", err
	}

	lead.Messages = append(lead.Messages, message)

	return id, nil
}

// Find lead by ID
func FindLead(deps Deps, id bson.ObjectId) (*Lead, error) {
	data := &Lead{}
	err := deps.Mgo().C("orders").FindId(id).One(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Fetch multiple leads by conditions
func FetchLeads(deps Deps, query Query) (list Leads, err error) {
	err = query(deps.Mgo().C("orders")).All(&list)
	if err != nil {
		return
	}

	// Sort inner message lists.
	for index, _ := range list {
		sort.Sort(list[index].Messages)
	}
	return
}

// Take new leads query.
func NewLeads(take, skip int) Query {
	return func(col *mgo.Collection) *mgo.Query {
		return col.Find(bsonq(SoftDelete, FieldExists("messages", false))).Limit(take).Skip(skip).Sort("-updated_at")
	}
}

// Take new leads query.
func AllLeads(take, skip int) Query {
	return func(col *mgo.Collection) *mgo.Query {
		return col.Find(bsonq(SoftDelete)).Limit(take).Skip(skip).Sort("-updated_at")
	}
}

// Take leads that match search query.
func SearchLeads(search string, take, skip int) Query {
	return func(col *mgo.Collection) *mgo.Query {
		query := col.Find(bsonq(SoftDelete, FulltextSearch(search)))
		return query.Select(bson.M{"score": bson.M{"$meta": "textScore"}}).Limit(take).Skip(skip).Sort("-updated_at", "$textScore:score")
	}
}

// Take next to answer leads query.
func NextUpLeads(deps Deps, take, skip int) Query {
	var boundaries []struct {
		Id bson.ObjectId `bson:"_id"`
	}

	err := deps.Mgo().C("orders").Pipe([]bson.M{
		{
			"$match": bson.M{"deleted_at": bson.M{"$exists": false}, "messages": bson.M{"$exists": true}},
		},
		{
			"$sort": bson.M{"updated_at": -1},
		},
		{
			"$project": bson.M{"lastMessage": bson.M{"$arrayElemAt": []interface{}{"$messages", -1}}},
		},
		{
			"$match": bson.M{"lastMessage.type": "inbound"},
		},
		{
			"$skip": skip,
		},
		{
			"$limit": take,
		},
	}).All(&boundaries)

	// Panic this weird exception.
	if err != nil {
		panic(err)
	}

	list := make([]bson.ObjectId, len(boundaries))
	for index, item := range boundaries {
		list[index] = item.Id
	}

	return func(col *mgo.Collection) *mgo.Query {
		return col.Find(bsonq(WithinID(list))).Limit(take).Skip(skip).Sort("-updated_at")
	}
}
