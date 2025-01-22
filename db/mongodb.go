package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type QSA struct {
	Pais                                 *string `json:"pais"`
	NomeSocio                            string  `json:"nome_socio"`
	CodigoPais                           *string `json:"codigo_pais"`
	FaixaEtaria                          string  `json:"faixa_etaria"`
	CnpjCpfDoSocio                       string  `json:"cnpj_cpf_do_socio"`
	QualificacaoSocio                    string  `json:"qualificacao_socio"`
	CodigoFaixaEtaria                    int     `json:"codigo_faixa_etaria"`
	DataEntradaSociedade                 string  `json:"data_entrada_sociedade"`
	IdentificadorDeSocio                 int     `json:"identificador_de_socio"`
	CpfRepresentanteLegal                string  `json:"cpf_representante_legal"`
	NomeRepresentanteLegal               string  `json:"nome_representante_legal"`
	CodigoQualificacaoSocio              int     `json:"codigo_qualificacao_socio"`
	QualificacaoRepresentanteLegal       string  `json:"qualificacao_representante_legal"`
	CodigoQualificacaoRepresentanteLegal int     `json:"codigo_qualificacao_representante_legal"`
}

type CNAESecundario struct {
	Codigo    int    `json:"codigo"`
	Descricao string `json:"descricao"`
}

type Empresa struct {
	UF                                 string           `json:"uf"`
	CEP                                string           `json:"cep"`
	QSA                                []QSA            `json:"qsa"`
	CNPJ                               string           `json:"cnpj"`
	Pais                               *string          `json:"pais"`
	Email                              string           `json:"email"`
	Porte                              string           `json:"porte"`
	Bairro                             string           `json:"bairro"`
	Numero                             string           `json:"numero"`
	DDDFax                             string           `json:"ddd_fax"`
	Municipio                          string           `json:"municipio"`
	Logradouro                         string           `json:"logradouro"`
	CNAEFiscal                         int              `json:"cnae_fiscal"`
	CodigoPais                         *string          `json:"codigo_pais"`
	Complemento                        string           `json:"complemento"`
	CodigoPorte                        int              `json:"codigo_porte"`
	RazaoSocial                        string           `json:"razao_social"`
	NomeFantasia                       string           `json:"nome_fantasia"`
	CapitalSocial                      float64          `json:"capital_social"`
	DDDTelefone1                       string           `json:"ddd_telefone_1"`
	DDDTelefone2                       string           `json:"ddd_telefone_2"`
	OpcaoPeloMEI                       bool             `json:"opcao_pelo_mei"`
	DescricaoPorte                     string           `json:"descricao_porte"`
	CodigoMunicipio                    int              `json:"codigo_municipio"`
	CNAESecundarios                    []CNAESecundario `json:"cnaes_secundarios"`
	NaturezaJuridica                   string           `json:"natureza_juridica"`
	SituacaoEspecial                   string           `json:"situacao_especial"`
	OpcaoPeloSimples                   bool             `json:"opcao_pelo_simples"`
	SituacaoCadastral                  int              `json:"situacao_cadastral"`
	DataOpcaoPeloMEI                   *time.Time       `json:"data_opcao_pelo_mei"`
	DataExclusaoDoMEI                  *time.Time       `json:"data_exclusao_do_mei"`
	CNAEFiscalDescricao                string           `json:"cnae_fiscal_descricao"`
	CodigoMunicipioIBGE                int              `json:"codigo_municipio_ibge"`
	DataInicioAtividade                string           `json:"data_inicio_atividade"`
	DataSituacaoEspecial               *time.Time       `json:"data_situacao_especial"`
	DataOpcaoPeloSimples               string           `json:"data_opcao_pelo_simples"`
	DataSituacaoCadastral              string           `json:"data_situacao_cadastral"`
	NomeCidadeNoExterior               string           `json:"nome_cidade_no_exterior"`
	CodigoNaturezaJuridica             int              `json:"codigo_natureza_juridica"`
	DataExclusaoDoSimples              *time.Time       `json:"data_exclusao_do_simples"`
	MotivoSituacaoCadastral            int              `json:"motivo_situacao_cadastral"`
	EnteFederativoResponsavel          string           `json:"ente_federativo_responsavel"`
	IdentificadorMatrizFilial          int              `json:"identificador_matriz_filial"`
	QualificacaoDoResponsavel          int              `json:"qualificacao_do_responsavel"`
	DescricaoSituacaoCadastral         string           `json:"descricao_situacao_cadastral"`
	DescricaoTipoDeLogradouro          string           `json:"descricao_tipo_de_logradouro"`
	DescricaoMotivoSituacaoCadastral   string           `json:"descricao_motivo_situacao_cadastral"`
	DescricaoIdentificadorMatrizFilial string           `json:"descricao_identificador_matriz_filial"`
}

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
	Context  context.Context
}

// NewMongoDB inicializa uma nova conexão MongoDB encapsulada em uma estrutura.
func NewMongoDB() (*MongoDB, error) {
	uri := os.Getenv("MONGO_RL")
	dbName := os.Getenv("DATABASE")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao MongoDB: %w", err)
	}

	// Verifica a conexão
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("erro ao pingar no MongoDB: %w", err)
	}

	return &MongoDB{
		Client:   client,
		Database: client.Database(dbName),
		Context:  ctx,
	}, nil
}

// CreateCollection cria a coleção especificada, se ainda não existir.
func (m *MongoDB) CreateCollection(collectionName string) error {
	collections, err := m.Database.ListCollectionNames(m.Context, bson.D{{Key: "name", Value: collectionName}})
	if err != nil {
		return fmt.Errorf("erro ao listar coleções: %w", err)
	}

	for _, name := range collections {
		if name == collectionName {
			fmt.Println("Coleção já existe:", collectionName)
			return nil
		}
	}

	if err := m.Database.CreateCollection(m.Context, collectionName); err != nil {
		return fmt.Errorf("erro ao criar a coleção: %w", err)
	}

	fmt.Println("Coleção criada com sucesso:", collectionName)
	return nil
}

// CreateIndexes cria os índices na coleção especificada.
func (m *MongoDB) CreateIndexes(collectionName string) error {
	collection := m.Database.Collection(collectionName)

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "cnpj", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "razao_social", Value: 1}}},
		{Keys: bson.D{{Key: "nome_fantasia", Value: 1}}},
		{Keys: bson.D{{Key: "capital_social", Value: 1}}},
		{Keys: bson.D{{Key: "qsa.nome_socio", Value: 1}}},
	}

	_, err := collection.Indexes().CreateMany(m.Context, indexes)
	if err != nil {
		return fmt.Errorf("erro ao criar índices: %w", err)
	}

	fmt.Println("Índices criados com sucesso na coleção:", collectionName)
	return nil
}

// DropCollection exclui completamente uma coleção específica.
func (m *MongoDB) DropCollection(collectionName string) error {
	collection := m.Database.Collection(collectionName)

	if err := collection.Drop(m.Context); err != nil {
		return fmt.Errorf("erro ao excluir a coleção: %w", err)
	}

	fmt.Println("Coleção excluída com sucesso:", collectionName)
	return nil
}

// CreateCompanies insere uma matriz de dados no MongoDB.
func (m *MongoDB) CreateCompanies(collectionName string, empresas []Empresa) error {
	collection := m.Database.Collection(collectionName)

	docs := make([]interface{}, len(empresas))
	for i, empresa := range empresas {
		docs[i] = empresa
	}

	_, err := collection.InsertMany(m.Context, docs)
	if err != nil {
		return fmt.Errorf("erro ao inserir empresas no MongoDB: %w", err)
	}

	fmt.Println("Empresas inseridas com sucesso no MongoDB!")
	return nil
}

// Close encerra a conexão com o MongoDB.
func (m *MongoDB) Close() error {
	if err := m.Client.Disconnect(m.Context); err != nil {
		return fmt.Errorf("erro ao desconectar do MongoDB: %w", err)
	}
	fmt.Println("Conexão com o MongoDB encerrada com sucesso.")
	return nil
}
