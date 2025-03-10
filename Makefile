docker-re-run:
	docker compose up -d --build --remove-orphans

docker-clean-up:
	docker container prune -f
	docker image prune -f
