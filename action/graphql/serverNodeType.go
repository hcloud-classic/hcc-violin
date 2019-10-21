package graphql

import "github.com/graphql-go/graphql"

var serverNodeType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "ServerNode",
		Fields: graphql.Fields{
			"uuid": &graphql.Field{
				Type: graphql.String,
			},
			"server_uuid": &graphql.Field{
				Type: graphql.String,
			},
			"node_uuid": &graphql.Field{
				Type: graphql.String,
			},
			"created_at": &graphql.Field{
				Type: graphql.DateTime,
			},
		},
	},
)
