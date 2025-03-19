docker-re-run:
	docker compose up -d --build --remove-orphans

docker-clean-up:
	@docker images -q | sort -u | while read image; do \
		created=$$(docker inspect --format='{{.Created}}' "$$image"); \
		if [ $$(date -d "$$created" +%s) -lt $$(date -d "12 hours ago" +%s) ]; then \
			docker rmi "$$image"; \
		fi; \
	done

export-csv:
	@echo "Exporting CSV files from log_api container into local /tmp dir..."
	docker-compose exec -T log_api sh -c 'cd /data && tar cf - *.csv' | tar xf - -C tmp
	@echo "Export has been finished, CSVs are already in /tmp"
