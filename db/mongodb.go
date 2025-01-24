package db

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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
func NewMongoDB() (MongoDB, error) {
	uri := os.Getenv("MONGO_URL")
	dbName := os.Getenv("DATABASE")

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

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
	// collections, err := m.Database.ListCollectionNames(m.Context, bson.D{{Key: "name", Value: companyTableName}, {Key: "name", Value: metaTableName}})
	// if err != nil {
	// 	return fmt.Errorf("erro ao listar coleções: %w", err)
	// }

	// for _, name := range collections {
	// 	if name == collectionName {
	// 		fmt.Println("Coleção já existe:", collectionName)
	// 		return nil
	// 	}

	if err := m.Database.CreateCollection(m.Context, companyTableName); err != nil {
		return fmt.Errorf("erro ao criar a coleção: %w", err)
	}

	fmt.Println("Coleção criada com sucesso:", companyTableName)

	if err := m.Database.CreateCollection(m.Context, metaTableName); err != nil {
		return fmt.Errorf("erro ao criar a coleção: %w", err)
	}

	fmt.Println("Coleção criada com sucesso:", metaTableName)

	// }

	return nil
}

// CreateIndexes cria os índices na coleção especificada.
func (m *MongoDB) CreateIndexes() error {
	collection := m.Database.Collection(companyTableName)

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "cnpj", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "qsa.cpf", Value: 1}}},
		{Keys: bson.D{{Key: "razao_social", Value: 1}}},
		{Keys: bson.D{{Key: "nome_fantasia", Value: 1}}},
		{Keys: bson.D{{Key: "capital_social", Value: 1}}},
		{Keys: bson.D{{Key: "ddd_telefone_1", Value: 1}}},
		{Keys: bson.D{{Key: "ddd_telefone_2", Value: 1}}},
		{Keys: bson.D{{Key: "natureza_juridica", Value: 1}}},
		{Keys: bson.D{{Key: "qsa.cnpj_cpf_do_socio", Value: 1}}},
		{Keys: bson.D{{Key: "qsa.qualificacao_socio", Value: 1}}},
		{Keys: bson.D{{Key: "qsa.qualificacao_representante_legal", Value: 1}}},
		{Keys: bson.D{{Key: "qsa.nome_socio", Value: 1}}},
		{Keys: bson.D{{Key: "bairro", Value: 1}}},
		{Keys: bson.D{{Key: "cep", Value: 1}}},
		{Keys: bson.D{{Key: "cnae_fiscal", Value: 1}}},
		{Keys: bson.D{{Key: "cnaes_secundarios", Value: 1}}},
		{Keys: bson.D{{Key: "data_inicio_atividade", Value: 1}}},
		{Keys: bson.D{{Key: "email", Value: 1}}},
		{Keys: bson.D{{Key: "codigo_municipio", Value: 1}}},
		{Keys: bson.D{{Key: "descricao_situacao_cadastral", Value: 1}}},
		{Keys: bson.D{{Key: "uf", Value: 1}}},
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
func (m *MongoDB) CreateCompaniesMongo(batch [][]string) error {
	// Recupera o nome da coleção do ambiente
	collectionName := companyTableName
	if collectionName == "" {
		return fmt.Errorf("nome da coleção não definido na variável de ambiente COLLECTION")
	}

	// Verifica se a conexão está configurada corretamente
	if m == nil {
		return fmt.Errorf("conexão com o MongoDB não inicializada")
	}

	collection := m.Database.Collection(collectionName)

	// Cria uma lista para armazenar os documentos a serem inseridos
	var empresas []interface{}

	// Itera sobre o batch para processar os dados
	for _, row := range batch {

		if len(row) < 2 {
			// Ignora linhas que não tenham pelo menos dois elementos
			fmt.Println("Linha ignorada devido ao tamanho insuficiente:", row)
			continue
		}

		// O segundo elemento do batch é o JSON que será convertido para a estrutura Empresa
		empresaJSON := row[1]
		var empresa Empresa

		// Deserializa o JSON para a estrutura Empresa
		err := json.Unmarshal([]byte(empresaJSON), &empresa)
		if err != nil {
			fmt.Printf("Erro ao desserializar JSON: %s, erro: %v\n", empresaJSON, err)
			continue
		}

		// Adiciona a empresa convertida à lista
		empresas = append(empresas, empresa)
	}

	// Insere as empresas no MongoDB
	if len(empresas) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := collection.InsertMany(ctx, empresas)
		if err != nil {
			return fmt.Errorf("erro ao inserir empresas no MongoDB: %w", err)
		}
		fmt.Println("Empresas inseridas com sucesso no MongoDB!")
	} else {
		fmt.Println("Nenhuma empresa válida para inserir.")
	}

	return nil
}

func (m *MongoDB) MetaSaveMongo(k, v string) error {
	// Verifica se a conexão com o MongoDB está inicializada
	if m == nil {
		return fmt.Errorf("conexão com o MongoDB não inicializada")
	}

	// Recupera o nome da coleção `meta` do ambiente
	collectionName := "meta"
	if collectionName == "" {
		return fmt.Errorf("nome da coleção não definido na variável de ambiente META_COLLECTION")
	}

	// Obtém a coleção
	collection := m.Database.Collection(collectionName)

	// Valida o tamanho da chave
	if len(k) > 16 {
		return fmt.Errorf("a chave pode ter no máximo 16 caracteres")
	}

	// Cria o documento a ser inserido
	doc := bson.M{
		"key":   k,
		"value": v,
	}

	// Insere o documento na coleção
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("erro ao salvar %s na coleção meta: %w", k, err)
	}

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
