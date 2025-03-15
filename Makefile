docker-re-run:
	docker compose up -d --build --remove-orphans

docker-clean-up:
	docker images -q | sort -u | while read image; do
		created=$(docker inspect --format='{{.Created}}' "$image")
		if [ $(date -d "$created" +%s) -lt $(date -d "24 hours ago" +%s) ]; then
			docker rmi "$image"
		fi
	done
