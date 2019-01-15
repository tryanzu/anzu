package activity

import "gopkg.in/mgo.v2/bson"

func Count(d deps, q bson.M) int {
	n, err := d.Mgo().C("activity").Find(q).Count()
	if err != nil {
		panic(err)
	}
	return n
}
