.PHONY: init up down infra-up migrate logs seed

init:
	$(MAKE) -C shop-infra init

up:
	$(MAKE) -C shop-infra up

down:
	$(MAKE) -C shop-infra down

infra-up:
	$(MAKE) -C shop-infra infra-up

migrate:
	$(MAKE) -C shop-infra migrate

logs:
	$(MAKE) -C shop-infra logs

seed:
	$(MAKE) -C shop-infra seed
