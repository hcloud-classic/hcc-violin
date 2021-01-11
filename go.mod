module hcc/violin

go 1.13

require (
	github.com/Terry-Mao/goconf v0.0.0-20161115082538-13cb73d70c44
	github.com/TylerBrock/colorjson v0.0.0-20200706003622-8a50f05110d2
	github.com/apparentlymart/go-cidr v1.1.0
	github.com/fatih/color v1.10.0 // indirect
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/protobuf v1.4.3
	github.com/hcloud-classic/hcc_errors v1.1.2
	github.com/hcloud-classic/pb v0.0.0
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d
	github.com/streadway/amqp v1.0.0
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b // indirect
	golang.org/x/sys v0.0.0-20210110051926-789bb1bd4061 // indirect
	golang.org/x/text v0.3.5 // indirect
	google.golang.org/appengine v1.6.2 // indirect
	google.golang.org/genproto v0.0.0-20210108203827-ffc7fda8c3d7 // indirect
	google.golang.org/grpc v1.34.0
	google.golang.org/protobuf v1.25.0 // indirect
)

replace github.com/hcloud-classic/pb => ../pb
