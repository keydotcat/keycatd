FROM scratch
ADD bin/keycatd.linux /keycatd
ADD keycatd.toml /etc/keycat/keycatd.toml
WORKDIR /
CMD ["/keycatd", "--config", "/etc/keycat/keycatd.toml"]