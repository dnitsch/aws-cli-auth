OWNER := dnitsch
NAME := aws-cli-auth
VERSION := v0.7.2
REVISION := $(shell git rev-parse --short HEAD)


LDFLAGS := -ldflags="-s -w -X \"github.com/dnitsch/aws-cli-auth/cmd.Version=$(VERSION)\" -X \"github.com/dnitsch/aws-cli-auth/cmd.Revision=$(REVISION)\" -extldflags -static"

.PHONY: test test_ci tidy install buildprep build buildmac buildwin

test: test_prereq
	go test `go list ./... | grep -v */generated/` -v -mod=readonly -coverprofile=.coverage/out | go-junit-report > .coverage/report-junit.xml && \
	gocov convert .coverage/out | gocov-xml > .coverage/report-cobertura.xml

test_ci:
	go test ./... -mod=readonly

test_prereq: 
	mkdir -p .coverage
	go install github.com/jstemmer/go-junit-report@v0.9.1 && \
	go install github.com/axw/gocov/gocov@v1.0.0 && \
	go install github.com/AlekSi/gocov-xml@v1.0.0

tidy: install 
	go mod tidy

install:
	go mod vendor

.PHONY: clean
clean:
	rm -rf bin/*
	rm -rf dist/*
	rm -rf vendor/*

.PHONY: cross-build
cross-build:
	for os in darwin linux windows; do \
	    [ $$os = "windows" ] && EXT=".exe"; \
		GOOS=$$os CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o dist/$(NAME)-$$os$$EXT .; \
	done

release: cross-build
	git tag $(VERSION)
	git push origin $(VERSION)
	id=$(curl \
	-X POST \
	-u $(OWNER):$(PAT) \
	-H "Accept: application/vnd.github.v3+json" \
	https://api.github.com/repos/$(OWNER)/$(NAME)/releases \
	-d '{"tag_name":"$(VERSION)","generate_release_notes":true,"prerelease":false}' | jq -r .id) \
	upload_url=https://uploads.github.com/repos/$(OWNER)/$(NAME)/releases/$$id/assets; \
	for asset in dist/*; do \
		echo $$asset; \
		name=$(echo $$asset | cut -c 6-); \
		curl -u $(OWNER):$(PAT) -H "Content-Type: application/x-binary" -X POST --data-binary "@$$asset" "$$upload_url?name=$$name"; \
	done 

.PHONY: deps
deps:
	GO111MODULE=on go mod vendor

.PHONY: dist
dist:
	cd dist && \
	$(DIST_DIRS) cp ../LICENSE {} \; && \
	$(DIST_DIRS) cp ../README.md {} \; && \
	$(DIST_DIRS) tar -zcf $(NAME)-$(VERSION)-{}.tar.gz {} \; && \
	$(DIST_DIRS) zip -r $(NAME)-$(VERSION)-{}.zip {} \; && \
	cd ..


