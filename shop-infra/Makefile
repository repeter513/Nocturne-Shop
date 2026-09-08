.PHONY: up down infra-up migrate logs seed env init

up: env
	docker compose up -d --build

envoy-reload:
	docker compose exec envoy envoy --restart

migrate:
	./scripts/migrate-all.sh

down:
	docker compose down

infra-up: env
	docker compose up -d postgres adminer

logs:
	docker compose logs -f

seed:
	cd ../shop-seed && $(MAKE) seed

env:
	@test -f .env || cp .env.example .env
	@for f in env/*.env.example; do \
		target="$${f%.example}"; \
		test -f "$$target" || cp "$$f" "$$target"; \
	done

init: env
