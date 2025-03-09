lint:
	pylint ./**/*.py

lint-docker:
	docker compose exec log_visualizer pylint ./**/*.py

test:
	python -m unittest ./log_visualizer/*.py

test-docker:
	docker compose exec log_visualizer python -m unittest ./log_visualizer/*.py
