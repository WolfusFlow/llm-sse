
service-build:
	docker compose build --no-cache

service-up:
	docker compose up -d

service-logs:
	docker compose logs -f

service-down:
	docker compose down
