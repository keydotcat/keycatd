FROM golang:latest 
ADD bin/keycatd.linux /
RUN mkdir -p /etc/keycat
ADD keycatd.toml /etc/keycat
WORKDIR /
CMD ["/keycatd.linux", "-config", "/etc/keycat/keycatd.toml"]