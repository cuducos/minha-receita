package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/cuducos/minha-receita/transform"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type company struct {
	Id   string            `json:"id"`
	Json transform.Company `json:"json"`
}

type MongoDB struct {
	client *mongo.Client
	db     *mongo.Database
	ctx    context.Context
}

// NewMongoDB initializes a new MongoDB connection wrapped in a structure.
func NewMongoDB(uri string) (MongoDB, error) {
	opts := options.Client().ApplyURI(uri)
	ctx := context.Background()
	c, err := mongo.Connect(ctx, opts)
	if err != nil {
		return MongoDB{}, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	if err := c.Ping(ctx, nil); err != nil {
		return MongoDB{}, fmt.Errorf("failed to ping to MongoDB: %w", err)
	}
	u := strings.Split(uri, "?")[0] // Remove query parameters from the URI
	ps := strings.Split(u, "/")
	n := ps[len(ps)-1]
	if n == "" || strings.Contains(n, "@") { // ensure the database name is valid
		return MongoDB{}, fmt.Errorf("no database name found in the uri")
	}
	return MongoDB{client: c, db: c.Database(n), ctx: ctx}, nil
}

// Create creates the required collections.
func (m *MongoDB) Create() error {
	for _, c := range []string{companyTableName, metaTableName} {
		if err := m.db.CreateCollection(m.ctx, c); err != nil {
			return fmt.Errorf("error creating collection %s: %w", c, err)
		}
		log.Output(1, fmt.Sprintf("Collection %s created successfully", c))
	}
	return nil
}

func (m *MongoDB) createIndexes() error {
	for _, n := range []string{companyTableName, metaTableName} {
		c := m.db.Collection(n)
		var k string
		if n == metaTableName {
			k = keyFieldName
		} else {
			k = idFieldName
		}
		i := []mongo.IndexModel{{Keys: bson.D{{Key: k, Value: 1}}}}
		_, err := c.Indexes().CreateMany(m.ctx, i)
		if err != nil {
			return fmt.Errorf("error creating index for %s in %s: %w", k, n, err)
		}
	}
	return nil
}

// Drop deletes the collectiosn created by `Create`.
func (m *MongoDB) Drop() error {
	for _, n := range []string{companyTableName, metaTableName} {
		c := m.db.Collection(n)
		if err := c.Drop(m.ctx); err != nil {
			return fmt.Errorf("error deleting collection %s: %w", n, err)
		}
		log.Output(1, fmt.Sprintf("Collection %s deleted successfully", n))
	}
	return nil
}

// CreateCompanies writes a batch of company data to MongoDB
func (m *MongoDB) CreateCompanies(batch [][]string) error {
	if m == nil {
		return fmt.Errorf("mongodb connection not initialized")
	}
	coll := m.db.Collection(companyTableName)
	var cs []interface{} // required by MongoDb pkg
	for _, r := range batch {
		if len(r) < 2 {
			return fmt.Errorf("line skipped due to insufficient length: %s", r)
		}
		var c company
		c.Id = r[0]
		err := json.Unmarshal([]byte(r[1]), &c.Json)
		if err != nil {
			return fmt.Errorf("error deserializing JSON: %s\nerror: %w", r[1], err)
		}
		cs = append(cs, c)
	}
	if len(cs) == 0 {
		return nil
	}
	ctx := context.Background()
	_, err := coll.InsertMany(ctx, cs)
	if err != nil {
		return fmt.Errorf("error inserting companies into MongoDB: %w", err)
	}
	return nil
}

// MetaSave inserts if the key doesn't exist, or updates the value if it does.
func (m *MongoDB) MetaSave(k, v string) error {
	ctx := context.Background()
	c := m.db.Collection(metaTableName)
	if len(k) > 16 {
		return fmt.Errorf("the key can have a maximum of 16 characters")
	}
	f := bson.M{"key": k}
	o := options.Update().SetUpsert(true) // if it does not exist, creates it
	upd := bson.M{"$set": bson.M{"key": k, "value": v}}
	_, err := c.UpdateOne(ctx, f, upd, o)
	if err != nil {
		return fmt.Errorf("error saving %s in the meta collection: %w", k, err)
	}
	return nil
}

// MetaRead reads a key/value pair from the metadata collection.
func (m *MongoDB) MetaRead(k string) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var result struct {
		Value string `bson:"value"`
	}
	c := m.db.Collection(metaTableName)
	err := c.FindOne(ctx, bson.M{"key": k}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", fmt.Errorf("metadata key %s not found", k)
		}
		return "", fmt.Errorf("error looking for metadata key %s: %w", k, err)
	}
	return result.Value, nil
}

// Close terminates the connection to MongoDB.
func (m *MongoDB) Close() {
	if err := m.client.Disconnect(m.ctx); err != nil {
		log.Output(1, fmt.Sprintf("error disconnecting from MongoDB: %s", err))
	}
}

// PreLoad runs before starting to load data into the database.
func (m *MongoDB) PreLoad() error {
	return nil
}

// PostLoad runs after loading data into the database. Removes duplicates and
// creates indexes.
func (m *MongoDB) PostLoad() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	coll := m.db.Collection(companyTableName)
	p := mongo.Pipeline{
		{{"$group", bson.D{
			{"_id", fmt.Sprintf("$%s", idFieldName)},
			{"docs", bson.D{{"$push", "$_id"}}},
			{"count", bson.D{{"$sum", 1}}},
		}}},
		{{"$match", bson.D{{"count", bson.D{{"$gt", 1}}}}}},
	}
	c, err := coll.Aggregate(ctx, p)
	if err != nil {
		return fmt.Errorf("error executing aggregation: %w", err)
	}
	defer c.Close(ctx)
	for c.Next(ctx) {
		var result struct {
			ID   string               `bson:"_id"`
			Docs []primitive.ObjectID `bson:"docs"`
		}
		if err := c.Decode(&result); err != nil {
			return fmt.Errorf("error decoding result: %w", err)
		}
		// Keep the first document and remove the others
		if len(result.Docs) > 1 {
			toRemove := result.Docs[1:] // Delete all but the first document
			_, err := coll.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": toRemove}})
			if err != nil {
				return fmt.Errorf("error removing duplicates: %w", err)
			}
		}
	}
	if err := c.Err(); err != nil {
		return fmt.Errorf("error when iterating through results: %w", err)
	}
	if err := m.createIndexes(); err != nil {
		return fmt.Errorf("error creating indexes: %w", err)
	}
	return nil
}

func (m *MongoDB) GetCompany(id string) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	coll := m.db.Collection(companyTableName)
	var c company
	err := coll.FindOne(ctx, bson.M{idFieldName: id}).Decode(&c)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", fmt.Errorf("no document found for CNPJ %s", id)
		}
		return "", fmt.Errorf("error querying CNPJ %s: %w", id, err)
	}
	b, err := json.Marshal(c.Json)
	if err != nil {
		return "", fmt.Errorf("error serializing JSON for CNPJ %s: %w", id, err)
	}
	return string(b), nil
}

func (m *MongoDB) ExtraIndexes(idxs []string) error {
	if err := transform.ValidateIndexes(idxs); err != nil {
		return fmt.Errorf("index name error: %w", err)
	}
	log.Output(1, fmt.Sprintf("indexes: %s", strings.Join(idxs, ", ")))
	return fmt.Errorf("extra-indexes not implemented (yet)")
}
