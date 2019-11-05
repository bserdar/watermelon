module github.com/bserdar/watermelon

go 1.13

require (
	github.com/dustin/go-jsonpointer v0.0.0-20160814072949-ba0abeacc3dc
	github.com/dustin/gojson v0.0.0-20160307161227-2e71ec9dd5ad // indirect
	github.com/golang/protobuf v1.3.2
	github.com/google/uuid v1.1.1
	github.com/hnakamur/go-scp v0.0.0-20190410043705-badb3bf1aae2
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	golang.org/x/crypto v0.0.0-20191105034135-c7e5f84aec59
	golang.org/x/net v0.0.0-20191105084925-a882066a44e0
	google.golang.org/grpc v1.24.0
	gopkg.in/yaml.v2 v2.2.5
)

replace github.com/hnakamur/go-scp => ./vendor/github.com/hnakamur/go-scp
