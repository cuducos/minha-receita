package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/cuducos/minha-receita/transform"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoRecord struct {
	Id   string            `json:"id" bson:"id"`
	Json transform.Company `json:"json" bson:"json"`
}

type MongoDB struct {
	client *mongo.Client
	db     *mongo.Database
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
	return MongoDB{client: c, db: c.Database(n)}, nil
}

// Create creates the required collections.
func (m *MongoDB) Create() error {
	for _, c := range []string{companyTableName, metaTableName} {
		slog.Info("Creating", "collection", c)
		if err := m.db.CreateCollection(context.Background(), c); err != nil {
			return fmt.Errorf("error creating collection %s: %w", c, err)
		}
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
		_, err := c.Indexes().CreateMany(context.Background(), i)
		if err != nil {
			return fmt.Errorf("error creating index for %s in %s: %w", k, n, err)
		}
	}
	return nil
}

// Drop deletes the collectiosn created by `Create`.
func (m *MongoDB) Drop() error {
	for _, n := range []string{companyTableName, metaTableName} {
		slog.Info("Deleting", "collection", n)
		c := m.db.Collection(n)
		if err := c.Drop(context.Background()); err != nil {
			return fmt.Errorf("error deleting collection %s: %w", n, err)
		}
	}
	return nil
}

// CreateCompanies writes a batch of company data to MongoDB
func (m *MongoDB) CreateCompanies(batch [][]string) error {
	if m == nil {
		return fmt.Errorf("mongodb connection not initialized")
	}
	coll := m.db.Collection(companyTableName)
	var cs []any // required by MongoDb pkg
	for _, c := range batch {
		if len(c) < 2 {
			return fmt.Errorf("line skipped due to insufficient length: %s", c)
		}
		var r mongoRecord
		r.Id = c[0]
		err := json.Unmarshal([]byte(c[1]), &r.Json)
		if err != nil {
			return fmt.Errorf("error deserializing JSON: %s\nerror: %w", c[1], err)
		}
		cs = append(cs, r)
	}
	if len(cs) == 0 {
		return nil
	}
	_, err := coll.InsertMany(context.Background(), cs)
	if err != nil {
		return fmt.Errorf("error inserting companies into MongoDB: %w", err)
	}
	return nil
}

// MetaSave inserts if the key doesn't exist, or updates the value if it does.
func (m *MongoDB) MetaSave(k, v string) error {
	c := m.db.Collection(metaTableName)
	if len(k) > 16 {
		return fmt.Errorf("the key can have a maximum of 16 characters")
	}
	f := bson.M{"key": k}
	o := options.Update().SetUpsert(true) // if it does not exist, creates it
	upd := bson.M{"$set": bson.M{"key": k, "value": v}}
	_, err := c.UpdateOne(context.Background(), f, upd, o)
	if err != nil {
		return fmt.Errorf("error saving %s in the meta collection: %w", k, err)
	}
	return nil
}

// MetaRead reads a key/value pair from the metadata collection.
func (m *MongoDB) MetaRead(k string) (string, error) {
	var result struct {
		Value string `bson:"value"`
	}
	c := m.db.Collection(metaTableName)
	err := c.FindOne(context.Background(), bson.M{"key": k}).Decode(&result)
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
	if err := m.client.Disconnect(context.Background()); err != nil {
		slog.Error("Error disconnecting from MongoDB", "error", err)
	}
}

// PreLoad runs before starting to load data into the database.
func (m *MongoDB) PreLoad() error {
	return nil
}

// PostLoad runs after loading data into the database. Removes duplicates and
// creates indexes.
func (m *MongoDB) PostLoad() error {
	ctx := context.Background()
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
	coll := m.db.Collection(companyTableName)
	var r mongoRecord
	err := coll.FindOne(context.Background(), bson.M{idFieldName: id}).Decode(&r)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", fmt.Errorf("no document found for CNPJ %s", id)
		}
		return "", fmt.Errorf("error querying CNPJ %s: %w", id, err)
	}
	b, err := json.Marshal(r.Json)
	if err != nil {
		return "", fmt.Errorf("error serializing JSON for CNPJ %s: %w", id, err)
	}
	return string(b), nil
}

// Search returns paginated results with JSON for companies bases on a search
// query
func (m *MongoDB) Search(q *Query) (string, error) {
	coll := m.db.Collection(companyTableName)
	f := bson.M{}
	if len(q.UF) > 0 {
		if len(q.UF) == 1 {
			f["json.uf"] = q.UF[0]
		} else {
			f["json.uf"] = bson.M{"$in": q.UF}
		}
	}
	if len(q.CNAEFiscal) > 0 {
		if len(q.CNAEFiscal) == 1 {
			f["json.cnae_fiscal"] = q.CNAEFiscal[0]
		} else {
			f["json.cnae_fiscal"] = bson.M{"$in": q.UF}
		}
	}
	if len(q.CNAE) > 0 {
		if len(q.CNAE) == 1 {
			f["$or"] = []bson.M{
				{"cnae_fiscal": q.CNAE[0]},
				{"cnaes_secundarios.codigo": bson.M{"$in": q.CNAE}},
			}
		} else {
			f["$or"] = []bson.M{
				{"cnae_fiscal": bson.M{"$in": q.CNAE}},
				{"cnaes_secundarios.codigo": bson.M{"$in": q.CNAE}},
			}
		}
	}
	if q.Cursor != nil {
		f["id"] = bson.M{"$gt": *q.Cursor}
	}
	opts := options.Find().SetSort(bson.D{{Key: "id", Value: 1}}).SetLimit(int64(q.Limit))
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	c, err := coll.Find(ctx, f, opts)
	if err != nil {
		return "", fmt.Errorf("error running query %#v: %w", q, err)
	}
	defer c.Close(ctx)
	var r []mongoRecord
	if err := c.All(ctx, &r); err != nil {
		return "", fmt.Errorf("error decoding results: %w", err)
	}
	var cs []transform.Company
	var cur string
	for i, c := range r {
		cs = append(cs, c.Json)
		if i == len(r)-1 {
			cur = c.Id
		}
	}
	p := newPage(cs, cur)
	b, err := json.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("error serializing the query result: %w", err)
	}
	return string(b), nil
}

func (m *MongoDB) CreateExtraIndexes(idxs []string) error {
	if err := transform.ValidateIndexes(idxs); err != nil {
		return fmt.Errorf("index name error: %w", err)
	}
	slog.Info("Creating the indexes…")
	c := m.db.Collection(companyTableName)
	var i []mongo.IndexModel
	for _, v := range idxs {
		i = append(i, mongo.IndexModel{
			Keys:    bson.D{{Key: fmt.Sprintf("json.%s", v), Value: 1}},
			Options: options.Index().SetName(fmt.Sprintf("idx_json.%s", v)),
		})
	}
	r, err := c.Indexes().CreateMany(context.Background(), i)
	if err != nil {
		return fmt.Errorf("error creating indexes: %w", err)
	}
	l := "index"
	if len(i) > 1 {
		l = "indexes"
	}
	slog.Info(fmt.Sprintf("%d %s successfully created in the collection %s", len(r), l, companyTableName))
	return nil
}
