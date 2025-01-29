package db

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type date time.Time

type QSA struct {
	IdentificadorDeSocio                 *int    `json:"identificador_de_socio"`
	NomeSocio                            string  `json:"nome_socio"`
	CNPJCPFDoSocio                       string  `json:"cnpj_cpf_do_socio"`
	CodigoQualificacaoSocio              *int    `json:"codigo_qualificacao_socio"`
	QualificaoSocio                      *string `json:"qualificacao_socio"`
	DataEntradaSociedade                 *string `json:"data_entrada_sociedade"`
	CodigoPais                           *int    `json:"codigo_pais"`
	Pais                                 *string `json:"pais"`
	CPFRepresentanteLegal                string  `json:"cpf_representante_legal"`
	NomeRepresentanteLegal               string  `json:"nome_representante_legal"`
	CodigoQualificacaoRepresentanteLegal *int    `json:"codigo_qualificacao_representante_legal"`
	QualificacaoRepresentanteLegal       *string `json:"qualificacao_representante_legal"`
	CodigoFaixaEtaria                    *int    `json:"codigo_faixa_etaria"`
	FaixaEtaria                          *string `json:"faixa_etaria"`
}

type CNAESecundario struct {
	Codigo    int    `json:"codigo"`
	Descricao string `json:"descricao"`
}

type Empresa struct {
	Cnpj string   `json:"cnpj"`
	Json Detalhes `json:"json"`
}

type Detalhes struct {
	CNPJ                             string           `json:"cnpj"`
	IdentificadorMatrizFilial        *int             `json:"identificador_matriz_filial"`
	DescricaoMatrizFilial            *string          `json:"descricao_identificador_matriz_filial"`
	NomeFantasia                     string           `json:"nome_fantasia"`
	SituacaoCadastral                *int             `json:"situacao_cadastral"`
	DescricaoSituacaoCadastral       *string          `json:"descricao_situacao_cadastral"`
	DataSituacaoCadastral            *string          `json:"data_situacao_cadastral"`
	MotivoSituacaoCadastral          *int             `json:"motivo_situacao_cadastral"`
	DescricaoMotivoSituacaoCadastral *string          `json:"descricao_motivo_situacao_cadastral"`
	NomeCidadeNoExterior             string           `json:"nome_cidade_no_exterior"`
	CodigoPais                       *int             `json:"codigo_pais"`
	Pais                             *string          `json:"pais"`
	DataInicioAtividade              *string          `json:"data_inicio_atividade"`
	CNAEFiscal                       *int             `json:"cnae_fiscal"`
	CNAEFiscalDescricao              *string          `json:"cnae_fiscal_descricao"`
	DescricaoTipoDeLogradouro        string           `json:"descricao_tipo_de_logradouro"`
	Logradouro                       string           `json:"logradouro"`
	Numero                           string           `json:"numero"`
	Complemento                      string           `json:"complemento"`
	Bairro                           string           `json:"bairro"`
	CEP                              string           `json:"cep"`
	UF                               string           `json:"uf"`
	CodigoMunicipio                  *int             `json:"codigo_municipio"`
	CodigoMunicipioIBGE              *int             `json:"codigo_municipio_ibge"`
	Municipio                        *string          `json:"municipio"`
	Telefone1                        string           `json:"ddd_telefone_1"`
	Telefone2                        string           `json:"ddd_telefone_2"`
	Fax                              string           `json:"ddd_fax"`
	Email                            *string          `json:"email"`
	SituacaoEspecial                 string           `json:"situacao_especial"`
	DataSituacaoEspecial             *string          `json:"data_situacao_especial"`
	OpcaoPeloSimples                 *bool            `json:"opcao_pelo_simples"`
	DataOpcaoPeloSimples             *string          `json:"data_opcao_pelo_simples"`
	DataExclusaoDoSimples            *string          `json:"data_exclusao_do_simples"`
	OpcaoPeloMEI                     *bool            `json:"opcao_pelo_mei"`
	DataOpcaoPeloMEI                 *string          `json:"data_opcao_pelo_mei"`
	DataExclusaoDoMEI                *string          `json:"data_exclusao_do_mei"`
	RazaoSocial                      string           `json:"razao_social"`
	CodigoNaturezaJuridica           *int             `json:"codigo_natureza_juridica"`
	NaturezaJuridica                 *string          `json:"natureza_juridica"`
	QualificacaoDoResponsavel        *int             `json:"qualificacao_do_responsavel"`
	CapitalSocial                    *float32         `json:"capital_social"`
	CodigoPorte                      *int             `json:"codigo_porte"`
	Porte                            *string          `json:"porte"`
	EnteFederativoResponsavel        string           `json:"ente_federativo_responsavel"`
	DescricaoPorte                   string           `json:"descricao_porte"`
	QSA                              []QSA            `json:"qsa"`
	CNAESecundarios                  []CNAESecundario `json:"cnaes_secundarios"`
}

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
	Context  context.Context
}

// NewMongoDB inicializa uma nova conexão MongoDB encapsulada em uma estrutura.
func NewMongoDB(dbName string) (MongoDB, error) {
	uri := os.Getenv("DATABASE_URL")

	ctx, _ := context.WithCancel(context.Background())

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return MongoDB{}, fmt.Errorf("erro ao conectar ao MongoDB: %w", err)
	}

	// Verifica a conexão
	if err := client.Ping(ctx, nil); err != nil {
		return MongoDB{}, fmt.Errorf("erro ao pingar no MongoDB: %w", err)
	}

	return MongoDB{
		Client:   client,
		Database: client.Database(dbName),
		Context:  ctx,
	}, nil
}

// CreateCollection cria a coleção especificada, se ainda não existir.
func (m *MongoDB) CreateCollection() error {

	if err := m.Database.CreateCollection(m.Context, companyTableName); err != nil {
		return fmt.Errorf("erro ao criar a coleção: %w", err)
	}

	fmt.Println("Coleção criada com sucesso:", companyTableName)

	if err := m.Database.CreateCollection(m.Context, metaTableName); err != nil {
		return fmt.Errorf("erro ao criar a coleção: %w", err)
	}

	fmt.Println("Coleção criada com sucesso:", metaTableName)

	return nil
}

// CreateIndexes cria os índices na coleção especificada.
func (m *MongoDB) CreateIndexes() error {

	collection := m.Database.Collection(companyTableName)

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
	_, err := collection.Indexes().CreateMany(m.Context, indexes)
	if err != nil {
		return fmt.Errorf("erro ao criar índices: %w", err)
	}

	fmt.Println("Índices criados com sucesso na coleção:", companyTableName)
	return nil
}

// DropCollection exclui completamente uma coleção específica.
func (m *MongoDB) DropCollection() error {

	collections := []string{companyTableName, metaTableName}

	for _, v := range collections {
		collection := m.Database.Collection(v)

		if err := collection.Drop(m.Context); err != nil {
			return fmt.Errorf("erro ao excluir a coleção: %w", err)
		}

		fmt.Println("Coleção excluída com sucesso:", v)
	}
	m.Close()

	return nil
}

// CreateCompanies insere uma matriz de dados no MongoDB.
func (m *MongoDB) CreateCompanies(batch [][]string) error {

	if m == nil {
		return fmt.Errorf("conexão com o MongoDB não inicializada")
	}

	collection := m.Database.Collection(companyTableName)

	var empresas []interface{}

	for _, row := range batch {

		if len(row) < 2 {
			fmt.Println("Linha ignorada devido ao tamanho insuficiente:", row)
			continue
		}

		var empresa Empresa
		empresa.Cnpj = row[0]

		empresaJSON := row[1]
		err := json.Unmarshal([]byte(empresaJSON), &empresa.Json)
		if err != nil {
			fmt.Printf("Erro ao desserializar JSON: %s, erro: %v\n", empresaJSON, err)
			continue
		}

		empresas = append(empresas, empresa)
	}

	if len(empresas) > 0 {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		_, err := collection.InsertMany(ctx, empresas)
		if err != nil {
			return fmt.Errorf("erro ao inserir empresas no MongoDB: %w", err)
		}

	} else {
		fmt.Println("Nenhuma empresa válida para inserir.")
	}

	return nil
}

func (m *MongoDB) MetaSave(k, v string) error {

	if m == nil {
		return fmt.Errorf("conexão com o MongoDB não inicializada")
	}

	collection := m.Database.Collection(metaTableName)

	if len(k) > 16 {
		return fmt.Errorf("a chave pode ter no máximo 16 caracteres")
	}

	doc := bson.M{
		"key":   k,
		"value": v,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("erro ao salvar %s na coleção meta: %w", k, err)
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

	collection := m.Database.Collection(metaTableName)

	err := collection.FindOne(ctx, bson.M{"key": k}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", fmt.Errorf("metadata key %s not found", k)
		}
		return "", fmt.Errorf("error looking for metadata key %s: %w", k, err)
	}

	return result.Value, nil
}

// Close encerra a conexão com o MongoDB.
func (m *MongoDB) Close() error {

	if err := m.Client.Disconnect(m.Context); err != nil {
		return fmt.Errorf("erro ao desconectar do MongoDB: %w", err)
	}
	fmt.Println("Conexão com o MongoDB encerrada com sucesso.")
	return nil
}

// PreLoad runs before starting to load data into the database. Currently it
// disables autovacuum on PostgreSQL.
func (m *MongoDB) PreLoad() error {

	return nil
}

// PostLoad runs after loading data into the database. Currently it re-enables
// autovacuum on PostgreSQL.
func (m *MongoDB) PostLoad() error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	collection := m.Database.Collection(companyTableName)

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
		return fmt.Errorf("erro ao executar agregação: %w", err)
	}
	defer cursor.Close(ctx)

	// Itera pelos resultados e remove duplicados
	for cursor.Next(ctx) {
		var result struct {
			ID   string               `bson:"_id"`
			Docs []primitive.ObjectID `bson:"docs"`
		}

		if err := cursor.Decode(&result); err != nil {
			return fmt.Errorf("erro ao decodificar resultado: %w", err)
		}

		// Mantém o primeiro documento e remove os demais
		if len(result.Docs) > 1 {
			toRemove := result.Docs[1:] // Exclui o primeiro documento
			_, err := collection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": toRemove}})
			if err != nil {
				return fmt.Errorf("erro ao remover duplicados: %w", err)
			}
		}
	}

	if err := cursor.Err(); err != nil {
		return fmt.Errorf("erro ao iterar pelos resultados: %w", err)
	}

	return nil
}

func (m *MongoDB) GetCompany(cnpj string) (string, error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	collection := m.Database.Collection(companyTableName)

	filter := bson.M{"cnpj": cnpj}

	var empresa Empresa

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
