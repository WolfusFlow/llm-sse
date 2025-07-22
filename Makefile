
service-build:
	docker compose build --no-cache

service-up:
	docker compose up -d

service-logs:
	docker compose logs -f

service-down:
	docker compose down

test:
	go test -race ./...

test-coverage:
	go test -coverprofile=coverage.out ./...

test-coverage-detailed: test-coverage
	go tool cover -func=coverage.out

test-coverage-visualize: test-coverage
	go tool cover -html=coverage.out
