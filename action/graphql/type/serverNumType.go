package graphqlType

import "github.com/graphql-go/graphql"

// ServerNumType : Graphql object type of ServerNumType
var ServerNumType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "ServerNumType",
		Fields: graphql.Fields{
			"number": &graphql.Field{
				Type: graphql.Int,
			},
		},
	},
)
