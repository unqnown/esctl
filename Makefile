main=esctl.go
index=testing

fmt:
	gofmt -w .

build:
	go build $(main)

install:
	go install $(main)

create: install
	esctl create $(index) -b .mapping/$(index).json

dump: install
	esctl dump $(index) -d .backup/$(index).json

restore: install
	esctl delete $(index)
	esctl create $(index) -b .mapping/$(index).json
	esctl restore $(index) -d .backup/$(index).json