docker-re-run:
	docker compose up -d --build --remove-orphans

docker-clean-up:
	docker system prune --force
