build:
	govendor build +local

test:
	govendor test +local

test-full: example
	govendor test -race +local,^program

test-cover:
	goveralls -service=travis-ci

install-libs:
	govendor install +vendor,^program

install-deps:
	go get github.com/mattn/goveralls
	go get github.com/kardianos/govendor
	govendor sync

generate: generate-types generate-type-tests generate-joins

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
	go build ./types/...

generate-type-tests:
	go build -o ./types/gen/gen ./types/gen
	./types/gen/gen v1.Pod > types/pod/generated_test.go
	./types/gen/gen v1beta1.Ingress > types/ingress/generated_test.go
	./types/gen/gen v1.Secret > types/secret/generated_test.go
	./types/gen/gen v1.Service > types/service/generated_test.go
	./types/gen/gen v1.Event > types/event/generated_test.go
	./types/gen/gen v1.Node > types/node/generated_test.go
	./types/gen/gen v1.ReplicationController > types/replicationcontroller/generated_test.go
	./types/gen/gen v1beta1.ReplicaSet > types/replicaset/generated_test.go
	./types/gen/gen v1beta1.Deployment > types/deployment/generated_test.go
	./types/gen/gen v1beta1.DaemonSet > types/daemonset/generated_test.go
	go test ./types/...

generate-joins:
	go build -o ./join/gen/gen ./join/gen
	./join/gen/gen Service service '*v1.Service' Pod pod > ./join/generated_service_pod.go
	./join/gen/gen RC  replicationcontroller '*v1.ReplicationController' Pod pod > ./join/generated_rc_pod.go
	./join/gen/gen RS  replicaset '*v1beta1.ReplicaSet' Pod pod > ./join/generated_rs_pod.go
	./join/gen/gen Deployment deployment '*v1beta1.Deployment' Pod pod > ./join/generated_deployment_pod.go
	./join/gen/gen DaemonSet daemonset '*v1beta1.DaemonSet' Pod pod > ./join/generated_daemonset_pod.go
	./join/gen/gen Ingress ingress '*v1beta1.Ingress' Service service > ./join/generated_ingress_service.go
	go build ./join

example:
	go build -o _example/example ./_example

clean:
	rm join/gen/gen types/gen/gen _example/example 2>/dev/null || true

.PHONY: build test test-full install-deps install-libs \
	generate generate-types generate-type-tests generate-joins \
	example clean
