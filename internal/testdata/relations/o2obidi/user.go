// Code generated by entimport, DO NOT EDIT.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("age"), field.String("name"), field.Int("user_spouse").Optional()}
}
func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.To("child_user", User.Type).Unique(), edge.From("parent_user", User.Type).Unique().Field("user_spouse")}
}
func (User) Annotations() []schema.Annotation {
	return nil
}
