#!/usr/bin/env bash
# Installs/updates the NMRiH dedicated server through steamcmd, renders
# cfg/server.cfg from the environment and hands over to srcds.
#
#   entrypoint.sh run      install/update, then run the server (default)
#   entrypoint.sh update   install/update only, then exit (used by `make server-update`)
#   entrypoint.sh <cmd>    run <cmd> instead (debugging: `bash`, `ls`, ...)
set -euo pipefail

STEAM_APP_ID="${STEAM_APP_ID:-317670}"
SERVER_DIR="${SERVER_DIR:-/srv/nmrih}"
GAME_DIR="${SERVER_DIR}/nmrih"
STEAMCMD="${STEAMCMD:-/home/steam/steamcmd/steamcmd.sh}"
UPDATE_RETRIES="${NMRIH_UPDATE_RETRIES:-3}"

log() { echo "[entrypoint] $*"; }

# $1: "validate" to re-hash all ~6 GB of game files (slow, minutes), anything
# else for the plain update check a normal start does (seconds when current).
update_server() {
    local validate="${1:-}" attempt=1
    local -a app_update_args=("${STEAM_APP_ID}")
    if [ "${validate}" = "validate" ]; then
        app_update_args+=(validate)
    fi
    while [ "${attempt}" -le "${UPDATE_RETRIES}" ]; do
        log "steamcmd app_update ${app_update_args[*]} (attempt ${attempt}/${UPDATE_RETRIES})"
        if "${STEAMCMD}" \
            +@sSteamCmdForcePlatformType linux \
            +force_install_dir "${SERVER_DIR}" \
            +login anonymous \
            +app_update "${app_update_args[@]}" \
            +quit; then
            log "steamcmd finished"
            return 0
        fi
        # steamcmd regularly dies on transient CDN errors (0x202, 0x606);
        # retrying the same command just resumes the download.
        log "steamcmd failed, retrying in 15s"
        attempt=$((attempt + 1))
        sleep 15
    done
    log "steamcmd did not succeed after ${UPDATE_RETRIES} attempts"
    return 1
}

# srcds looks for steamclient.so under ~/.steam/sdk32 when it registers with
# the Steam master servers (that is also what a GSLT login needs).
link_steamclient() {
    mkdir -p "${HOME}/.steam/sdk32" "${HOME}/.steam/sdk64"
    [ -f "${SERVER_DIR}/bin/steamclient.so" ] \
        && ln -sf "${SERVER_DIR}/bin/steamclient.so" "${HOME}/.steam/sdk32/steamclient.so"
    [ -f "/home/steam/steamcmd/linux64/steamclient.so" ] \
        && ln -sf "/home/steam/steamcmd/linux64/steamclient.so" "${HOME}/.steam/sdk64/steamclient.so"
    return 0
}

render_cfg() {
    mkdir -p "${GAME_DIR}/cfg" "${GAME_DIR}/logs"

    export NMRIH_HOSTNAME="${NMRIH_HOSTNAME:-NMRiH Server}"
    export NMRIH_SV_PASSWORD="${NMRIH_SV_PASSWORD:-}"
    export NMRIH_RCON_PASSWORD="${NMRIH_RCON_PASSWORD:-}"
    export NMRIH_SV_REGION="${NMRIH_SV_REGION:-3}"
    export NMRIH_SV_CONTACT="${NMRIH_SV_CONTACT:-}"
    export NMRIH_TV_ENABLE="${NMRIH_TV_ENABLE:-0}"

    envsubst \
        '${NMRIH_HOSTNAME} ${NMRIH_SV_PASSWORD} ${NMRIH_RCON_PASSWORD} ${NMRIH_SV_REGION} ${NMRIH_SV_CONTACT} ${NMRIH_TV_ENABLE}' \
        < /opt/nmrih/cfg/server.cfg.template \
        > "${GAME_DIR}/cfg/server.cfg"

    # The GSLT deliberately does not go into server.cfg: that file is executed
    # after srcds has already logged into Steam, so the token only counts when
    # it is passed on the command line (see run_server).
    if [ -z "${NMRIH_GSLT:-}" ]; then
        log "NMRIH_GSLT is empty: the server stays reachable by direct connect but is not listed publicly"
    fi

    if [ -n "${NMRIH_RCON_PASSWORD:-}" ]; then
        log "rcon enabled on ${NMRIH_PORT:-27015}/tcp"
    else
        log "NMRIH_RCON_PASSWORD is empty: rcon stays disabled"
    fi

    log "rendered ${GAME_DIR}/cfg/server.cfg"
}

run_server() {
    cd "${SERVER_DIR}"

    if [ ! -x "${SERVER_DIR}/srcds_linux" ]; then
        log "srcds_linux missing in ${SERVER_DIR} - the steamcmd install did not complete"
        return 1
    fi

    local args=(
        -game nmrih
        -console
        -usercon
        # Docker's restart policy is the supervisor here: without -norestart
        # srcds_run/srcds respawns itself and a broken server looks healthy.
        -norestart
        -ip 0.0.0.0
        -port "${NMRIH_PORT:-27015}"
        +clientport "${NMRIH_CLIENT_PORT:-27005}"
        +tv_port "${NMRIH_TV_PORT:-27020}"
        +maxplayers "${NMRIH_MAXPLAYERS:-8}"
        +map "${NMRIH_MAP:-nmo_broadway}"
    )
    if [ -n "${NMRIH_EXTRA_ARGS:-}" ]; then
        # shellcheck disable=SC2206 # deliberate word splitting: free-form args
        local extra=(${NMRIH_EXTRA_ARGS})
        args+=("${extra[@]}")
    fi

    # Has to be a launch parameter: Steam logon happens before server.cfg runs.
    local printable_args=("${args[@]}")
    if [ -n "${NMRIH_GSLT:-}" ]; then
        args+=(+sv_setsteamaccount "${NMRIH_GSLT}")
        printable_args+=(+sv_setsteamaccount "<gslt>")
        log "GSLT configured, the server should show up in the public server browser"
    fi

    export LD_LIBRARY_PATH="${SERVER_DIR}:${SERVER_DIR}/bin:${LD_LIBRARY_PATH:-}"
    log "starting: srcds_linux ${printable_args[*]}"
    exec ./srcds_linux "${args[@]}"
}

case "${1:-run}" in
    run)
        if [ "${NMRIH_SKIP_UPDATE:-0}" = "1" ]; then
            log "NMRIH_SKIP_UPDATE=1, skipping the steamcmd update"
        else
            # A restart should not cost a full re-validation of the install;
            # `make server-update` (and NMRIH_VALIDATE=1) is the way to ask for
            # one.
            update_server "${NMRIH_VALIDATE:+validate}"
        fi
        link_steamclient
        render_cfg
        run_server
        ;;
    update)
        update_server validate
        link_steamclient
        ;;
    *)
        exec "$@"
        ;;
esac
