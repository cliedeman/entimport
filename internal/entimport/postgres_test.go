package entimport_test

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"testing"

	"ariga.io/atlas/sql/schema"

	"ariga.io/entimport/internal/entimport"

	"github.com/go-openapi/inflect"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

func TestPostgres(t *testing.T) {
	const testSchema = "public"
	r := require.New(t)
	ctx := context.Background()
	dsn := fmt.Sprintf("host=localhost port=5434 user=postgres dbname=test password=pass sslmode=disable")
	i := &entimport.ImportOptions{}
	entimport.WithDSN(dsn)(i)
	importer := &entimport.Postgres{Options: i}
	tests := []struct {
		name           string
		entities       []string
		expectedFields map[string]string
		mock           *schema.Schema
		expectedEdges  map[string]string
	}{
		{
			name: "single_table_fields",
			mock: MockPostgresSingleTableFields(),
			expectedFields: map[string]string{
				"user":
				//language=go
				`func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int16("age"), field.String("name")}
}`,
			},
			expectedEdges: map[string]string{

				`user`:
				//language=go
				`func (User) Edges() []ent.Edge {
	return nil
}`,
			},
			entities: []string{"user"},
		},
		{
			name: "fields_with_attributes",
			mock: MockPostgresTableFieldsWithAttributes(),
			expectedFields: map[string]string{
				"user":
				//language=go
				`func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id").Comment("some id"), field.Int16("age").Optional(), field.String("name").Comment("first name"), field.String("last_name").Optional().Comment("family name")}
}`,
			},
			expectedEdges: map[string]string{
				`user`:
				//language=go
				`func (User) Edges() []ent.Edge {
	return nil
}`,
			},
			entities: []string{"user"},
		},
		{
			name: "fields_with_unique_indexes",
			mock: MockPostgresTableFieldsWithUniqueIndexes(),
			expectedFields: map[string]string{
				"user":
				//language=go
				`func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id").Comment("some id"), field.Int16("age").Unique(), field.String("name").Comment("first name"), field.String("last_name").Optional().Comment("family name")}
}`,
			},
			expectedEdges: map[string]string{
				`user`:
				//language=go
				`func (User) Edges() []ent.Edge {
	return nil
}`,
			},
			entities: []string{"user"},
		},
		{
			name: "multi_table_fields",
			mock: MockPostgresMultiTableFields(),
			expectedFields: map[string]string{
				"user":
				//language=go
				`func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int16("age").Unique(), field.String("name"), field.String("last_name").Optional().Comment("not so boring")}
}`,
				"pet":
				//language=go
				`func (Pet) Fields() []ent.Field {
	return []ent.Field{field.Int("id").Comment("pet id"), field.Int16("age").Optional(), field.String("name")}
}`,
			},
			expectedEdges: map[string]string{
				`user`:
				//language=go
				`func (User) Edges() []ent.Edge {
	return nil
}`,
				`pet`: `func (Pet) Edges() []ent.Edge {
	return nil
}`,
			},
			entities: []string{"user", "pet"},
		},
		{
			name: "non_default_primary_key",
			mock: MockPostgresNonDefaultPrimaryKey(),
			expectedFields: map[string]string{
				"user":
				//language=go
				`func (User) Fields() []ent.Field {
	return []ent.Field{field.String("id").StorageKey("name"), field.String("last_name").Optional().Unique().Comment("not so boring")}
}`,
			},
			expectedEdges: map[string]string{
				`user`:
				//language=go
				`func (User) Edges() []ent.Edge {
	return nil
}`,
			},
			entities: []string{"user"},
		},
		{
			name: "relation_m2m_two_types",
			mock: MockPostgresM2MTwoTypes(),
			expectedFields: map[string]string{
				"user":
				//language=go
				`func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("age"), field.String("name")}
}`,
				"group":
				//language=go
				`func (Group) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("name")}
}`,
			},
			expectedEdges: map[string]string{
				"user":
				//language=go
				`func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.From("groups", Group.Type).Ref("users")}
}`,
				"group":
				//language=go
				`func (Group) Edges() []ent.Edge {
	return []ent.Edge{edge.To("users", User.Type)}
}`,
			},
			entities: []string{"user", "group"},
		},
		{
			name: "relation_m2m_same_type",
			mock: MockPostgresM2MSameType(),
			expectedFields: map[string]string{
				"user":
				//language=go
				`func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("age"), field.String("name")}
}`,
			},
			expectedEdges: map[string]string{
				"user":
				//language=go
				`func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.To("child_users", User.Type), edge.From("parent_users", User.Type)}
}`,
			},
			entities: []string{"user"},
		},
		{
			name: "relation_m2m_bidirectional",
			mock: MockPostgresM2MBidirectional(),
			expectedFields: map[string]string{
				"user":
				//language=go
				`func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("age"), field.String("name")}
}`,
			},
			expectedEdges: map[string]string{
				"user":
				//language=go
				`func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.To("child_users", User.Type), edge.From("parent_users", User.Type)}
}`,
			},
			entities: []string{"user"},
		},
		{
			name: "relation_o2o_two_types",
			mock: MockPostgresO2OTwoTypes(),
			expectedFields: map[string]string{
				"user":
				//language=go
				`func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("age"), field.String("name")}
}`,
				"card":
				// language=go
				`func (Card) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Time("expired"), field.String("number"), field.Int("user_card").Optional().Unique()}
}`,
			},
			expectedEdges: map[string]string{
				"user":
				//language=go
				`func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.To("card", Card.Type).Unique()}
}`,
				"card":
				// language=go
				`func (Card) Edges() []ent.Edge {
	return []ent.Edge{edge.From("user", User.Type).Ref("card").Unique().Field("user_card")}
}`,
			},
			entities: []string{"user", "card"},
		},
		{
			name: "relation_o2o_same_type",
			mock: MockPostgresO2OSameType(),
			expectedFields: map[string]string{
				"node":
				//language=go
				`func (Node) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("value"), field.Int("node_next").Optional().Unique()}
}`,
			},
			expectedEdges: map[string]string{
				"node":
				//language=go
				`func (Node) Edges() []ent.Edge {
	return []ent.Edge{edge.To("child_node", Node.Type).Unique(), edge.From("parent_node", Node.Type).Unique().Field("node_next")}
}`,
			},
			entities: []string{"node"},
		},
		{
			name: "relation_o2o_bidirectional",
			mock: MockPostgresO2OBidirectional(),
			expectedFields: map[string]string{
				"user":
				//language=go
				`func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("age"), field.String("name"), field.Int("user_spouse").Optional().Unique()}
}`,
			},
			expectedEdges: map[string]string{
				"user":
				//language=go
				`func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.To("child_user", User.Type).Unique(), edge.From("parent_user", User.Type).Unique().Field("user_spouse")}
}`,
			},
			entities: []string{"user"},
		},
		{
			name: "relation_o2m_two_types",
			mock: MockPostgresO2MTwoTypes(),
			expectedFields: map[string]string{
				"user":
				//language=go
				`func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("age"), field.String("name")}
}`,
				"pet":
				//language=go
				`func (Pet) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("name"), field.Int("user_pets").Optional()}
}`,
			},
			expectedEdges: map[string]string{
				"user":
				//language=go
				`func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.To("pets", Pet.Type)}
}`,
				"pet":
				//language=go
				`func (Pet) Edges() []ent.Edge {
	return []ent.Edge{edge.From("user", User.Type).Ref("pets").Unique().Field("user_pets")}
}`,
			},
			entities: []string{"user", "pet"},
		},
		{
			name: "relation_o2m_same_type",
			mock: MockPostgresO2MSameType(),
			expectedFields: map[string]string{
				"node":
				//language=go
				`func (Node) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("value"), field.Int("node_children").Optional()}
}`,
			},
			expectedEdges: map[string]string{
				"node":
				//language=go
				`func (Node) Edges() []ent.Edge {
	return []ent.Edge{edge.To("child_nodes", Node.Type), edge.From("parent_node", Node.Type).Unique().Field("node_children")}
}`,
			},
			entities: []string{"node"},
		},
		{
			name: "relation_o2x_other_side_ignored",
			mock: MockPostgresO2XOtherSideIgnored(),
			expectedFields: map[string]string{
				"pet":
				//language=go
				`func (Pet) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("name"), field.Int("user_pets").Optional()}
}`,
			},
			expectedEdges: map[string]string{
				"pet":
				//language=go
				`func (Pet) Edges() []ent.Edge {
	return nil
}`,
			},
			entities: []string{"pet"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemas := createTempDir(t)
			mock := &Inspector{}
			mock.On("InspectSchema", ctx, testSchema, &schema.InspectOptions{}).Return(tt.mock, nil)
			importer.Inspector = mock
			mutations, err := importer.SchemaMutations(ctx)
			r.NoError(err)
			err = entimport.WriteSchema(mutations, entimport.WithSchemaPath(schemas))
			r.NoError(err)
			actualFiles := readDir(t, schemas)
			r.EqualValues(len(tt.entities), len(actualFiles))
			for _, e := range tt.entities {
				f, err := parser.ParseFile(token.NewFileSet(), "", actualFiles[e+".go"], 0)
				r.NoError(err)
				typeName := inflect.Camelize(e)
				fieldMethod := lookupMethod(f, typeName, "Fields")
				r.NotNil(fieldMethod)
				var actualFields bytes.Buffer
				err = printer.Fprint(&actualFields, token.NewFileSet(), fieldMethod)
				r.NoError(err)
				r.EqualValues(tt.expectedFields[e], actualFields.String())
				edgeMethod := lookupMethod(f, typeName, "Edges")
				r.NotNil(edgeMethod)
				var actualEdges bytes.Buffer
				err = printer.Fprint(&actualEdges, token.NewFileSet(), edgeMethod)
				r.NoError(err)
				r.EqualValues(tt.expectedEdges[e], actualEdges.String())
			}
		})
	}
}

func TestPostgresJoinTableOnly(t *testing.T) {
	ctx := context.Background()
	importer := &entimport.Postgres{
		Options: &entimport.ImportOptions{},
	}
	mock := &Inspector{}
	mock.On("InspectSchema", ctx, "public", &schema.InspectOptions{}).Return(MockPostgresM2MJoinTableOnly(), nil)
	importer.Inspector = mock
	mutations, err := importer.SchemaMutations(ctx)
	require.Empty(t, mutations)
	require.Errorf(t, err, "join tables must be inspected with ref tables - append `tables` flag")
}

func TestPostgresNonDefaultSearchPath(t *testing.T) {
	ctx := context.Background()
	pgSchema := "non_public"
	dsn := fmt.Sprintf("host=localhost port=5434 user=postgres dbname=test password=pass sslmode=disable search_path=%s", pgSchema)
	i := &entimport.ImportOptions{}
	entimport.WithDSN(dsn)(i)
	importer := &entimport.Postgres{Options: i}
	mock := &Inspector{}
	mock.On("InspectSchema", ctx, pgSchema, &schema.InspectOptions{}).Return(&schema.Schema{}, nil)
	importer.Inspector = mock
	_, _ = importer.SchemaMutations(ctx)
	mock.AssertCalled(t, "InspectSchema", ctx, pgSchema, &schema.InspectOptions{})
}

func lookupMethod(file *ast.File, typeName string, methodName string) (m *ast.FuncDecl) {
	ast.Inspect(file, func(node ast.Node) bool {
		if decl, ok := node.(*ast.FuncDecl); ok {
			if decl.Name.Name != methodName || decl.Recv == nil || len(decl.Recv.List) != 1 {
				return true
			}
			if id, ok := decl.Recv.List[0].Type.(*ast.Ident); ok && id.Name == typeName {
				m = decl
				return false
			}
		}
		return true
	})
	return m
}
