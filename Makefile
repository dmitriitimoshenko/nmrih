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
	@echo "Экспорт CSV файлов из контейнера log_api в локальную директорию /tmp..."
	docker-compose exec log_api tar cf - -C /data *.csv | tar xf - -C /tmp
	@echo "Экспорт завершён, файлы находятся в /tmp"
