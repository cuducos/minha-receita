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

type empresa struct {
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

	clientOptions := options.Client().ApplyURI(uri)

	ctx := context.Background()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return MongoDB{}, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Verify the connection
	if err := client.Ping(ctx, nil); err != nil {
		return MongoDB{}, fmt.Errorf("failed to ping to MongoDB: %w", err)
	}

	// Extract the database name from the URI
	uriWithoutParams := strings.Split(uri, "?")[0] // Remove query parameters from the URI
	parts := strings.Split(uriWithoutParams, "/")  // Split the string by "/"
	dbName := parts[len(parts)-1]                  // The last part is the database name

	// Ensure the extracted database name is valid
	if dbName == "" || strings.Contains(dbName, "@") {
		log.Fatal("No database name found in the URI")
	}

	return MongoDB{
		client: client,
		db:     client.Database(dbName),
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
	fmt.Println("Creating the indexes...")
	collection := m.db.Collection(companyTableName)
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "cnpj", Value: 1}}},
		{Keys: bson.D{{Key: "json.cnpj", Value: 1}}},
		{Keys: bson.D{{Key: "json.qsa.cpf", Value: 1}}},
		{Keys: bson.D{{Key: "json.razaosocial", Value: 1}}},
		{Keys: bson.D{{Key: "json.nomefantasia", Value: 1}}},
		{Keys: bson.D{{Key: "json.capitalsocial", Value: 1}}},
		{Keys: bson.D{{Key: "json.telefone1", Value: 1}}},
		{Keys: bson.D{{Key: "json.telefone2", Value: 1}}},
		{Keys: bson.D{{Key: "json.naturezajuridica", Value: 1}}},
		{Keys: bson.D{{Key: "json.qsa.cnpjcpfdosocio", Value: 1}}},
		{Keys: bson.D{{Key: "json.qsa.qualificaosocio", Value: 1}}},
		{Keys: bson.D{{Key: "json.qsa.qualificacaorepresentantelegal", Value: 1}}},
		{Keys: bson.D{{Key: "json.qsa.nomesocio", Value: 1}}},
		{Keys: bson.D{{Key: "json.bairro", Value: 1}}},
		{Keys: bson.D{{Key: "json.cep", Value: 1}}},
		{Keys: bson.D{{Key: "json.cnaefiscal", Value: 1}}},
		{Keys: bson.D{{Key: "json.cnaessecundarios", Value: 1}}},
		{Keys: bson.D{{Key: "json.cnaessecundarios.codigo", Value: 1}}},
		{Keys: bson.D{{Key: "json.datainicioatividade", Value: 1}}},
		{Keys: bson.D{{Key: "json.email", Value: 1}}},
		{Keys: bson.D{{Key: "json.codigomunicipio", Value: 1}}},
		{Keys: bson.D{{Key: "json.codigomunicipioibge", Value: 1}}},
		{Keys: bson.D{{Key: "json.descricaosituacaocadastral", Value: 1}}},
		{Keys: bson.D{{Key: "json.uf", Value: 1}}},
	}
	_, err := collection.Indexes().CreateMany(m.ctx, indexes)
	if err != nil {
		return fmt.Errorf("error creating indexes: %w", err)
	}
	fmt.Println("Indexes successfully created in the collection:", companyTableName)
	return nil
}

// DropCollection completely deletes a specific collection.
func (m *MongoDB) DropCollection() error {
	collections := []string{companyTableName, metaTableName}
	for _, v := range collections {
		collection := m.db.Collection(v)

		if err := collection.Drop(m.ctx); err != nil {
			return fmt.Errorf("error deleting collection: %w", err)
		}

		fmt.Println("Collection deleted successfully:", v)
	}
	m.Close()
	return nil
}

// CreateCompanies insere uma matriz de dados no MongoDB.
func (m *MongoDB) CreateCompanies(batch [][]string) error {
	if m == nil {
		return fmt.Errorf("mongodb connection not initialized")
	}
	collection := m.db.Collection(companyTableName)
	var empresas []interface{}
	for _, r := range batch {
		if len(r) < 2 {
			return fmt.Errorf("line skipped due to insufficient length: %s", r)
		}
		var c empresa
		c.Cnpj = r[0]
		err := json.Unmarshal([]byte(r[1]), &c.Json)
		if err != nil {
			return fmt.Errorf("error deserializing JSON: %s, erro: %v", r[1], err)
		}
		empresas = append(empresas, c)
	}
	if len(empresas) > 0 {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		_, err := collection.InsertMany(ctx, empresas)
		if err != nil {
			return fmt.Errorf("error inserting companies into MongoDB: %w", err)
		}
	} else {
		fmt.Println("No valid company to insert.")
	}
	return nil
}

func (m *MongoDB) MetaSave(k, v string) error {
	if m == nil {
		return fmt.Errorf("MongoDB connection not initialized")
	}
	collection := m.db.Collection(metaTableName)
	if len(k) > 16 {
		return fmt.Errorf("the key can have a maximum of 16 characters")
	}
	doc := bson.M{
		"key":   k,
		"value": v,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("error saving %s to the meta collection: %w", k, err)
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
	collection := m.db.Collection(metaTableName)
	err := collection.FindOne(ctx, bson.M{"key": k}).Decode(&result)
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
	fmt.Println("Successfully disconnected from MongoDB")
	return nil
}

// PreLoad runs before starting to load data into the database.
func (m *MongoDB) PreLoad() error {
	return nil
}

// PostLoad runs after loading data into the database.
// Removes duplicates and creates indexes.
func (m *MongoDB) PostLoad() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	collection := m.db.Collection(companyTableName)
	pipeline := mongo.Pipeline{
		{{"$group", bson.D{
			{"_id", "$cnpj"},
			{"docs", bson.D{{"$push", "$_id"}}},
			{"count", bson.D{{"$sum", 1}}},
		}}},
		{{"$match", bson.D{{"count", bson.D{{"$gt", 1}}}}}},
	}
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return fmt.Errorf("error executing aggregation: %w", err)
	}
	defer cursor.Close(ctx)
	// Iterates through the results and removes duplicates
	for cursor.Next(ctx) {
		var result struct {
			ID   string               `bson:"_id"`
			Docs []primitive.ObjectID `bson:"docs"`
		}
		if err := cursor.Decode(&result); err != nil {
			return fmt.Errorf("error decoding result: %w", err)
		}
		// Keep the first document and remove the others
		if len(result.Docs) > 1 {
			toRemove := result.Docs[1:] // Delete the first document
			_, err := collection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": toRemove}})
			if err != nil {
				return fmt.Errorf("error removing duplicates: %w", err)
			}
		}
	}
	if err := cursor.Err(); err != nil {
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
	collection := m.db.Collection(companyTableName)
	filter := bson.M{"cnpj": cnpj}
	var empresa empresa
	err := collection.FindOne(ctx, filter).Decode(&empresa)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", fmt.Errorf("no document found for CNPJ %s", cnpj)
		}
		return "", fmt.Errorf("error querying CNPJ %s: %w", cnpj, err)
	}
	jsonBytes, err := json.Marshal(empresa.Json)
	if err != nil {
		return "", fmt.Errorf("error serializing JSON for CNPJ %s: %w", cnpj, err)
	}
	return string(jsonBytes), nil
}
