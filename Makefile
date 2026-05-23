.PHONY: build build-core build-cli up down test test-integration e2e clean docker-build-core docker-build-cli docker-build-web docker-build release-build

build: build-core build-cli

build-core:
	cd server && go build -o ../bin/stellaris-core ./cmd

build-cli:
	cd agent && go build -o ../bin/stellaris-cli ./cmd

up:
	docker compose -f deploy/docker-compose.yml up -d

down:
	docker compose -f deploy/docker-compose.yml down

test:
	@for mod in agent server shared/protocol shared/token; do \
		echo "==> $$mod"; \
		(cd $$mod && go test ./... -race) || exit 1; \
	done

# 集成测试：需要先 `make up` 起 postgres / redis / emqx
test-integration:
	@for mod in server; do \
		echo "==> $$mod (integration)"; \
		(cd $$mod && go test -tags=integration ./... -race) || exit 1; \
	done

e2e:
	cd tests/e2e && go test ./... -v -tags=e2e -timeout 120s

clean:
	rm -rf bin/
	go clean -testcache

docker-build-core:
	docker build -f deploy/Dockerfile.core -t stellaris-core:dev .

docker-build-cli:
	docker build -f deploy/Dockerfile.cli -t stellaris-cli:dev .

docker-build-web:
	cd web && npm ci --silent && npm run build
	docker build -f deploy/Dockerfile.web -t stellaris-web:dev .

docker-build: docker-build-core docker-build-cli docker-build-web

release-build:
	@VER=$$(git describe --tags --always); \
	for pair in linux-amd64 linux-arm64 darwin-amd64 darwin-arm64; do \
		OS=$${pair%-*}; ARCH=$${pair#*-}; \
		OUT="stellaris-cli-$$VER-$$OS-$$ARCH"; \
		mkdir -p dist/$$OUT; \
		GOOS=$$OS GOARCH=$$ARCH CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/$$OUT/stellaris-cli ./agent/cmd; \
		(cd dist && tar -czf $$OUT.tar.gz $$OUT && shasum -a 256 $$OUT.tar.gz > $$OUT.tar.gz.sha256); \
		echo "✓ dist/$$OUT.tar.gz"; \
	done
