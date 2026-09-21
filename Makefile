# docker compose reads .env by itself; make needs a couple of the same values
# to create bind-mount directories and to decide which compose files to use.
getenv = $(shell grep -E '^$(1)=' .env 2>/dev/null | tail -1 | cut -d= -f2-)

# Where the ~9 GB steamcmd install lives; kept in sync with docker-compose's
# ${NMRIH_DATA_DIR:-./server_data} so `prepare` creates the right directory
# (docker would otherwise create it root-owned and srcds could not write).
NMRIH_DATA_DIR := $(call getenv,NMRIH_DATA_DIR)
ifeq ($(strip $(NMRIH_DATA_DIR)),)
NMRIH_DATA_DIR := ./server_data
endif

# Single source of truth for the SourceMod version: the same pin the game
# server image is built with is what plugins are compiled against.
SOURCEMOD_VERSION := $(shell grep -E '^ARG SOURCEMOD_VERSION=' nmrih_server/Dockerfile | cut -d= -f2)

# Set PROXY_NETWORK in .env when another Traefik on this host already owns
# :80/:443 - see docker-compose.shared-proxy.yaml.
PROXY_NETWORK := $(call getenv,PROXY_NETWORK)
COMPOSE := docker compose -f docker-compose.yaml
MONITORING_SERVICES := redis log_api log_frontend
ifneq ($(strip $(PROXY_NETWORK)),)
COMPOSE := $(COMPOSE) -f docker-compose.shared-proxy.yaml
else
MONITORING_SERVICES += traefik
endif

.PHONY: help prepare up down restart ps logs \
	server-up server-update server-restart server-stop server-console server-logs server-check \
	plugins server-plugins-sync monitoring-up monitoring-logs docker-re-run docker-clean-up

help:
	@echo "make up              build + start everything (game server + dashboard)"
	@echo "make down            stop everything"
	@echo "make ps / logs       status / tail of all containers"
	@echo ""
	@echo "make server-up       start just the NMRiH game server (installs it on first run)"
	@echo "make server-update   pull the latest game build from Steam and restart"
	@echo "make server-restart  restart the game server"
	@echo "make server-console  attach to the srcds console (detach: ctrl-p ctrl-q)"
	@echo "make server-logs     tail the game server container"
	@echo "make server-check    is the server reachable from outside and listed on Steam?"
	@echo ""
	@echo "make plugins         compile sourcemod/scripting/*.sp (SourceMod $(SOURCEMOD_VERSION))"
	@echo "make server-plugins-sync  restart the game server if a compiled plugin changed"
	@echo ""
	@echo "make monitoring-up   start just the dashboard stack ($(MONITORING_SERVICES))"

# The bind mounts need to exist up front and belong to the invoking user.
prepare:
	@mkdir -p logs traefik/acme "$(NMRIH_DATA_DIR)"
	@touch traefik/acme/acme.json && chmod 600 traefik/acme/acme.json

up: prepare
	$(COMPOSE) up -d --build --remove-orphans

down:
	$(COMPOSE) down

restart:
	$(COMPOSE) restart

ps:
	$(COMPOSE) ps

logs:
	$(COMPOSE) logs -f --tail 100

server-up: prepare
	$(COMPOSE) up -d --build nmrih_server

# Runs steamcmd in a throwaway container (no published ports) while the server
# is down, so the update cannot fight the running srcds over the same files.
server-update: prepare
	$(COMPOSE) stop nmrih_server
	$(COMPOSE) run --rm --no-deps nmrih_server update
	$(COMPOSE) start nmrih_server

# Recreate rather than `restart`: a plain restart reuses the container's old
# environment, so .env edits (a new GSLT, a different map) would be ignored.
server-restart:
	$(COMPOSE) up -d --force-recreate --no-deps nmrih_server

server-stop:
	$(COMPOSE) stop nmrih_server

server-console:
	docker attach --detach-keys="ctrl-p,ctrl-q" nmrih_server

server-logs:
	$(COMPOSE) logs -f --tail 100 nmrih_server

# Needs a local Go toolchain; the check itself talks to the public address, so
# run it from somewhere outside this host to test the path players actually use.
server-check:
	$(MAKE) -C log_api server-check ARGS="$(ARGS)"

# Compiles with the compiler that ships inside the game server image, so the
# plugins are always built against exactly the SourceMod the server runs. No
# local SourcePawn toolchain needed; the .smx files land in sourcemod/plugins
# and reach the server on the next `make server-restart`.
plugins:
	@$(COMPOSE) build nmrih_server
	@docker run --rm -u "$$(id -u):$$(id -g)" -v "$(CURDIR)/sourcemod:/work" nmrih_server:latest \
		bash -c 'set -eu; sm=/opt/nmrih/addons/addons/sourcemod/scripting; \
			for sp in /work/scripting/*.sp; do \
				echo "compiling $$(basename "$$sp")"; \
				"$$sm/spcomp64" -i "$$sm/include" \
					-o "/work/plugins/$$(basename "$${sp%.sp}").smx" "$$sp"; \
			done'

# Plugins are mounted into the game container rather than baked into its image,
# so `up --build` sees no reason to recreate it when only a .smx changed and the
# entrypoint never gets to copy the new one in. Compare what is mounted with
# what is installed and restart only when they actually differ.
server-plugins-sync:
	@stale=0; \
	for plugin in sourcemod/plugins/*.smx; do \
		[ -f "$$plugin" ] || continue; \
		name=$$(basename "$$plugin"); \
		want=$$(md5sum < "$$plugin" | cut -d' ' -f1); \
		have=$$(docker exec nmrih_server md5sum \
			"/srv/nmrih/nmrih/addons/sourcemod/plugins/$$name" 2>/dev/null | cut -d' ' -f1); \
		if [ "$$want" != "$$have" ]; then echo "$$name is not the installed one"; stale=1; fi; \
	done; \
	if [ "$$stale" = "1" ]; then \
		echo "restarting the game server to install it"; \
		$(MAKE) server-restart; \
	else \
		echo "plugins already installed"; \
	fi

monitoring-up: prepare
	$(COMPOSE) up -d --build $(MONITORING_SERVICES)

monitoring-logs:
	$(COMPOSE) logs -f --tail 100 $(MONITORING_SERVICES)

# Kept for .github/workflows/deployment.yml and muscle memory.
docker-re-run: up

docker-clean-up:
	@docker images -q | sort -u | while read image; do \
		created=$$(docker inspect --format='{{.Created}}' "$$image"); \
		if [ $$(date -d "$$created" +%s) -lt $$(date -d "12 hours ago" +%s) ]; then \
			docker rmi "$$image"; \
		fi; \
	done
