package mongo

import (
	"gopkg.in/mgo.v2"
)

type Service struct {
	Session  *mgo.Session
	Database *mgo.Database
	Name     string
}

func NewService(connection string, name string) *Service {

	// Start a session with out replica set
	session, err := mgo.Dial(connection)

	if err != nil {

		// There has been an error connection to the database
		panic(err)
	}

	database := session.DB(name)

	// Set monotonic session behavior
	//session.SetMode(mgo.Monotonic, true)

	return &Service{Session: session, Database: database, Name: name}
}

func (s *Service) Save(o SaveOp) error {

	err := o.Save(s.Database)

	return err
}

type SaveOp interface {
	Save(*mgo.Database) error
}
