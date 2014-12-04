test: consul-pager.go
	go test -v .

build: cli/consul-pager.go
	gox -output "bin/consul-pager_{{.OS}}_{{.Arch}}" -os "linux darwin" -arch "amd64" ./cli


startconsul:
	rm -rf /tmp/consul
	consul agent -server -bootstrap -data-dir /tmp/consul

stopconsul:
	killall -TERM consul
