# Build the docs for the proto3 definition.

LANGUAGES=go # cpp go csharp objc python ruby js

bindings:
	for x in ${LANGUAGES}; do \
		protoc --proto_path=. \
			--$${x}_out=. \
			--experimental_editions \
			openrtb.proto agenticrtbframework.proto; \
		protoc --proto_path=. \
			--$${x}_out=. \
			--$${x}-grpc_out=require_unimplemented_servers=false:. \
			agenticrtbframeworkservices.proto; \
	done

check:
	prototool lint

clean:
	for x in ${LANGUAGES}; do \
		rm -fr $${x}/*; \
	done

docs:
	podman run --rm \
		-v ${PWD}:${PWD} \
		-w ${PWD} \
		pseudomuto/protoc-gen-doc \
		--doc_opt=html,doc.html \
		--proto_path=${PWD} \
		openrtb.proto agenticrtbframework.proto agenticrtbframeworkservices.proto

watch:
	fswatch  -r ./ | xargs -n1 make docs
