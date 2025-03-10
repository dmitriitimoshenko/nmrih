docker-re-run:
	docker compose up -d --build --remove-orphans

docker-clean-up:
	docker container prune -f
	docker image prune -f

export-csv:
	docker run --rm -v "$(pwd)/tmp":/tmp -v shared_data:/data alpine sh -c "cp /data/*.csv /tmp"
