package graphqlType

import "github.com/graphql-go/graphql"

// ServerNumType : Graphql object type of ServerNumType
var ServerNodeNumType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "ServerNodeNumType",
		Fields: graphql.Fields{
			"number": &graphql.Field{
				Type: graphql.Int,
			},
		},
	},
)
