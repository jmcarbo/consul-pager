test: consul-pager.go
	go test -v .

build: cli/consul-pager.go
	gox -output "bin/consul-pager_{{.OS}}_{{.Arch}}" -os "linux darwin" -arch "amd64" ./cli
