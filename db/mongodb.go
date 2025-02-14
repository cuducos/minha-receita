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
	Cnpj string            `json:"cnpj"`
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
	return MongoDB{
		client: c,
		db:     c.Database(n),
		ctx:    ctx,
	}, nil
}

// CreateCollection creates the specified collection if it does not already exist.
func (m *MongoDB) CreateCollection() error {

	if err := m.db.CreateCollection(m.ctx, companyTableName); err != nil {
		return fmt.Errorf("error creating collection: %w", err)
	}

	log.Output(1, fmt.Sprintf("Collection %s created successfully", companyTableName))

	if err := m.db.CreateCollection(m.ctx, metaTableName); err != nil {
		return fmt.Errorf("erro ao criar a coleção: %w", err)
	}

	log.Output(1, fmt.Sprintf("Collection %s created successfully", metaTableName))

	return nil
}

// CreateIndexes creates the indexes on the specified collection.
func (m *MongoDB) CreateIndexes() error {
	log.Output(1, "Creating the indexes...")
	c := m.db.Collection(companyTableName)
	i := []mongo.IndexModel{
		{Keys: bson.D{{Key: "cnpj", Value: 1}}},
	}
	_, err := c.Indexes().CreateMany(m.ctx, i)
	if err != nil {
		return fmt.Errorf("error creating indexes: %w", err)
	}
	log.Output(1, fmt.Sprintf("Indexes successfully created in the collection: %s", companyTableName))
	return nil
}

// DropCollection completely deletes a specific collection.
func (m *MongoDB) DropCollection() error {
	cs := []string{companyTableName, metaTableName}
	for _, v := range cs {
		c := m.db.Collection(v)

		if err := c.Drop(m.ctx); err != nil {
			return fmt.Errorf("error deleting collection: %w", err)
		}

		log.Output(1, fmt.Sprintf("Collection deleted successfully: %s", v))
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
		c.Cnpj = r[0]
		err := json.Unmarshal([]byte(r[1]), &c.Json)
		if err != nil {
			return fmt.Errorf("error deserializing JSON: %s, erro: %w", r[1], err)
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
	f := bson.M{
		"key": k,
	}

	//if it does not exist it creates it
	o := options.Update().SetUpsert(true)

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
func (m *MongoDB) Close() error {
	if err := m.client.Disconnect(m.ctx); err != nil {
		return fmt.Errorf("error disconnecting from MongoDB: %w", err)
	}
	return nil
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
			{"_id", "$cnpj"},
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
	if err := m.CreateIndexes(); err != nil {
		return fmt.Errorf("error creating indexes: %w", err)
	}
	return nil
}

func (m *MongoDB) GetCompany(cnpj string) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	coll := m.db.Collection(companyTableName)
	filter := bson.M{"cnpj": cnpj}
	var c company
	err := coll.FindOne(ctx, filter).Decode(&c)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", fmt.Errorf("no document found for CNPJ %s", cnpj)
		}
		return "", fmt.Errorf("error querying CNPJ %s: %w", cnpj, err)
	}
	b, err := json.Marshal(c.Json)
	if err != nil {
		return "", fmt.Errorf("error serializing JSON for CNPJ %s: %w", cnpj, err)
	}
	return string(b), nil
}
