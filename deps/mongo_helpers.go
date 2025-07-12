package deps

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Helper functions to ease migration from mgo to mongo-driver

// DefaultContext returns a context with 30 second timeout
func DefaultContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

// ShortContext returns a context with 10 second timeout
func ShortContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

// ObjectIDFromHex converts hex string to ObjectID, returns zero ObjectID on error
func ObjectIDFromHex(hex string) primitive.ObjectID {
	if id, err := primitive.ObjectIDFromHex(hex); err == nil {
		return id
	}
	return primitive.NilObjectID
}

// ConvertObjectIDs converts []bson.ObjectId to []primitive.ObjectID
func ConvertObjectIDs(oldIDs []interface{}) []primitive.ObjectID {
	newIDs := make([]primitive.ObjectID, len(oldIDs))
	for i, id := range oldIDs {
		if hexStr, ok := id.(string); ok {
			newIDs[i] = ObjectIDFromHex(hexStr)
		} else if objID, ok := id.(primitive.ObjectID); ok {
			newIDs[i] = objID
		}
	}
	return newIDs
}

// FindOneResult wraps mongo.SingleResult to provide mgo-like interface
type FindOneResult struct {
	result *mongo.SingleResult
}

func (r *FindOneResult) One(v interface{}) error {
	return r.result.Decode(v)
}

// FindResult wraps mongo.Cursor to provide mgo-like interface
type FindResult struct {
	cursor *mongo.Cursor
	ctx    context.Context
}

func (r *FindResult) All(results interface{}) error {
	defer r.cursor.Close(r.ctx)
	return r.cursor.All(r.ctx, results)
}

func (r *FindResult) One(result interface{}) error {
	defer r.cursor.Close(r.ctx)
	if r.cursor.Next(r.ctx) {
		return r.cursor.Decode(result)
	}
	return mongo.ErrNoDocuments
}

func (r *FindResult) Count() (int64, error) {
	// Note: This is a simplified version - real implementation would need the collection and filter
	return 0, nil
}

// Collection wraps mongo.Collection to provide mgo-like methods
type Collection struct {
	coll *mongo.Collection
}

func (c *Collection) FindId(id interface{}) *FindOneResult {
	ctx, _ := DefaultContext()
	filter := bson.M{"_id": id}
	result := c.coll.FindOne(ctx, filter)
	return &FindOneResult{result: result}
}

func (c *Collection) Find(query interface{}) *FindResult {
	ctx, _ := DefaultContext()
	cursor, _ := c.coll.Find(ctx, query)
	return &FindResult{cursor: cursor, ctx: ctx}
}

func (c *Collection) Insert(docs ...interface{}) error {
	ctx, cancel := DefaultContext()
	defer cancel()

	if len(docs) == 1 {
		_, err := c.coll.InsertOne(ctx, docs[0])
		return err
	}
	_, err := c.coll.InsertMany(ctx, docs)
	return err
}

func (c *Collection) UpdateId(id interface{}, update interface{}) error {
	ctx, cancel := DefaultContext()
	defer cancel()

	filter := bson.M{"_id": id}
	_, err := c.coll.UpdateOne(ctx, filter, update)
	return err
}

func (c *Collection) Update(selector interface{}, update interface{}) error {
	ctx, cancel := DefaultContext()
	defer cancel()

	_, err := c.coll.UpdateOne(ctx, selector, update)
	return err
}

func (c *Collection) UpdateAll(selector interface{}, update interface{}) (*mongo.UpdateResult, error) {
	ctx, cancel := DefaultContext()
	defer cancel()

	return c.coll.UpdateMany(ctx, selector, update)
}

func (c *Collection) UpsertId(id interface{}, update interface{}) (*mongo.UpdateResult, error) {
	ctx, cancel := DefaultContext()
	defer cancel()

	filter := bson.M{"_id": id}
	opts := options.Update().SetUpsert(true)
	return c.coll.UpdateOne(ctx, filter, update, opts)
}

func (c *Collection) Upsert(selector interface{}, update interface{}) (*mongo.UpdateResult, error) {
	ctx, cancel := DefaultContext()
	defer cancel()

	opts := options.Update().SetUpsert(true)
	return c.coll.UpdateOne(ctx, selector, update, opts)
}

func (c *Collection) Remove(selector interface{}) error {
	ctx, cancel := DefaultContext()
	defer cancel()

	_, err := c.coll.DeleteOne(ctx, selector)
	return err
}

func (c *Collection) RemoveAll(selector interface{}) (*mongo.DeleteResult, error) {
	ctx, cancel := DefaultContext()
	defer cancel()

	return c.coll.DeleteMany(ctx, selector)
}

func (c *Collection) RemoveId(id interface{}) error {
	ctx, cancel := DefaultContext()
	defer cancel()

	filter := bson.M{"_id": id}
	_, err := c.coll.DeleteOne(ctx, filter)
	return err
}

func (c *Collection) Count(query interface{}) (int64, error) {
	ctx, cancel := DefaultContext()
	defer cancel()

	return c.coll.CountDocuments(ctx, query)
}

func (c *Collection) Pipe(pipeline interface{}) *mongo.Cursor {
	ctx, _ := DefaultContext()
	cursor, _ := c.coll.Aggregate(ctx, pipeline)
	return cursor
}

// Database wraps mongo.Database to provide mgo-like methods
type Database struct {
	db *mongo.Database
}

func (d *Database) C(name string) *Collection {
	return &Collection{coll: d.db.Collection(name)}
}

func (d *Database) Collection(name string) *Collection {
	return &Collection{coll: d.db.Collection(name)}
}

// Helper to wrap mongo.Database
func WrapDatabase(db *mongo.Database) *Database {
	return &Database{db: db}
}
