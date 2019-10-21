package graphql

import "github.com/graphql-go/graphql"

var serverNum = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "ServerNum",
		Fields: graphql.Fields{
			"number": &graphql.Field{
				Type: graphql.Int,
			},
		},
	},
)
