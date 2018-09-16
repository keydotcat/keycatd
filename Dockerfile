FROM alpine
ADD bin/keycatd.linux /keycatd
RUN mkdir -p /etc/keycat
ADD keycatd.toml /etc/keycat
WORKDIR /
CMD ["/keycatd", "-config", "/etc/keycat/keycatd.toml"]