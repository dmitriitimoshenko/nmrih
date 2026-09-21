# NMRiH dedicated server (Steam app 317670, "No More Room in Hell Dedicated
# Server") on top of the community steamcmd image, which already ships
# steamcmd itself plus the unprivileged `steam` user (uid 1000) it wants to
# run as.
FROM cm2network/steamcmd:root

# srcds_linux is a 32-bit Source binary, so everything it links against has to
# come from the i386 arch (the base image has already enabled it for steamcmd).
# gettext-base provides envsubst, used to render cfg/server.cfg from the
# template at container start.
RUN dpkg --add-architecture i386 \
    && apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        gettext-base \
        lib32gcc-s1 \
        lib32stdc++6 \
        libcurl3t64-gnutls:i386 \
        libncurses6:i386 \
        libtinfo6:i386 \
    && rm -rf /var/lib/apt/lists/*

# Metamod:Source and SourceMod are pinned rather than tracked as "latest": two
# builds of the same commit have to produce the same server. Both are staged
# into the image here and copied into the game directory at start, so a restart
# never has to reach the network for them.
#
# Bump by editing these, or one-off with
#   docker compose build --build-arg SOURCEMOD_VERSION=1.13.0-gitXXXX nmrih_server
# Available builds: https://www.sourcemm.net/downloads.php and
# https://www.sourcemod.net/downloads.php
ARG METAMOD_VERSION=1.12.0-git1226
ARG SOURCEMOD_VERSION=1.12.0-git7253

ENV STEAM_APP_ID=317670 \
    SERVER_DIR=/srv/nmrih \
    ADDONS_STAGE=/opt/nmrih/addons \
    PLUGINS_DIR=/opt/nmrih/plugins \
    METAMOD_VERSION=${METAMOD_VERSION} \
    SOURCEMOD_VERSION=${SOURCEMOD_VERSION}

# Both archives unpack straight into the mod directory: they carry addons/ and
# cfg/ at their root, including the metamod.vdf that makes the engine load
# Metamod and the sourcemod.vdf that makes Metamod load SourceMod. The branch
# is the first two components of the version (1.12.0-git1226 -> 1.12).
RUN set -eux; \
    mm_branch="$(echo "${METAMOD_VERSION}" | cut -d. -f1,2)"; \
    sm_branch="$(echo "${SOURCEMOD_VERSION}" | cut -d. -f1,2)"; \
    mkdir -p "${ADDONS_STAGE}" "${PLUGINS_DIR}"; \
    curl -fsSL "https://mms.alliedmods.net/mmsdrop/${mm_branch}/mmsource-${METAMOD_VERSION}-linux.tar.gz" \
        | tar -xz -C "${ADDONS_STAGE}"; \
    curl -fsSL "https://sm.alliedmods.net/smdrop/${sm_branch}/sourcemod-${SOURCEMOD_VERSION}-linux.tar.gz" \
        | tar -xz -C "${ADDONS_STAGE}"; \
    test -f "${ADDONS_STAGE}/addons/metamod.vdf"; \
    test -f "${ADDONS_STAGE}/addons/metamod/sourcemod.vdf"; \
    test -d "${ADDONS_STAGE}/addons/sourcemod/bin"

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
COPY cfg /opt/nmrih/cfg
RUN chmod +x /usr/local/bin/entrypoint.sh \
    && mkdir -p "${SERVER_DIR}" \
    && chown steam:steam "${SERVER_DIR}"

USER steam
WORKDIR /srv/nmrih

# Game traffic (udp) and rcon (tcp) share 27015; SourceTV listens on 27020/udp.
EXPOSE 27015/udp 27015/tcp 27020/udp

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["run"]
