FROM scratch
ADD keycatd /keycatd
ADD keycatd.toml /etc/keycat/keycatd.toml
WORKDIR /
CMD ["/keycatd", "--config", "/etc/keycat/keycatd.toml"]