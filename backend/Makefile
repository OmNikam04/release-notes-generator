# make run	Start (no rebuild)	⚡ Fast
# make rebuild	Force rebuild	🐢 Slow
# make restart	Restart API	⚡⚡ Super fast
# make logs	View API logs	-
# make shell	Access API shell	-
# make clean	Clean everything	-
# Start containers (fast - no rebuild unless necessary)
run:
	docker-compose up

# Force rebuild and start (use when dependencies change)
rebuild:
	docker-compose up --build

# Stop and remove containers
down:
	docker-compose down

# Stop, remove containers and volumes (clean slate)
clean:
	docker-compose down -v

# Restart just the API service (useful for quick restart)
restart:
	docker-compose restart api

# View logs (follow mode)
logs:
	docker-compose logs -f api

# Run tests
test:
	go test ./internal/tests/...

# Access API container shell
shell:
	docker exec -it release-notest-generator-api sh

# Run with migrations enabled
run-migrate:
	RUN_MIGRATIONS=true docker-compose up