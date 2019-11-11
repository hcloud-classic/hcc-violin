package _type

import "github.com/graphql-go/graphql"

var ServerNum = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "ServerNum",
		Fields: graphql.Fields{
			"number": &graphql.Field{
				Type: graphql.Int,
			},
		},
	},
)
