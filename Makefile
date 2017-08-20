build:
	govendor build +local

test:
	govendor test +local

test-full: example
	govendor test -race +local

test-cover:
	goveralls -service=travis-ci

install-libs:
	govendor install +vendor,^program

install-deps:
	go get github.com/mattn/goveralls
	go get github.com/kardianos/govendor
	govendor sync

generate-types:
	genny -in=types/gen/template.go -out=types/pod/generated.go -pkg=pod gen 'ObjectType=*v1.Pod'
	genny -in=types/gen/template.go -out=types/ingress/generated.go -pkg=ingress gen 'ObjectType=*v1beta1.Ingress'
	genny -in=types/gen/template.go -out=types/secret/generated.go -pkg=secret gen 'ObjectType=*v1.Secret'
	genny -in=types/gen/template.go -out=types/service/generated.go -pkg=service gen 'ObjectType=*v1.Service'
	genny -in=types/gen/template.go -out=types/event/generated.go -pkg=event gen 'ObjectType=*v1.Event'
	genny -in=types/gen/template.go -out=types/node/generated.go -pkg=node gen 'ObjectType=*v1.Node'
	genny -in=types/gen/template.go -out=types/replicationcontroller/generated.go -pkg=replicationcontroller gen 'ObjectType=*v1.ReplicationController'
	genny -in=types/gen/template.go -out=types/replicaset/generated.go -pkg=replicaset gen 'ObjectType=*v1beta1.ReplicaSet'
	genny -in=types/gen/template.go -out=types/deployment/generated.go -pkg=deployment gen 'ObjectType=*v1beta1.Deployment'
	genny -in=types/gen/template.go -out=types/daemonset/generated.go -pkg=daemonset gen 'ObjectType=*v1beta1.DaemonSet'

example:
	go build -o _example/example ./_example

.PHONY: build test test-full install-deps install-libs example
