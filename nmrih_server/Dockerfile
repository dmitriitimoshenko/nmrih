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
        gettext-base \
        lib32gcc-s1 \
        lib32stdc++6 \
        libcurl3t64-gnutls:i386 \
        libncurses6:i386 \
        libtinfo6:i386 \
    && rm -rf /var/lib/apt/lists/*

ENV STEAM_APP_ID=317670 \
    SERVER_DIR=/srv/nmrih

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
