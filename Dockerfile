FROM debian:stable-slim
COPY folkloreserver /bin/folkloreserver
COPY ./init/ /init
RUN apt-get update && apt-get --yes install sqlite3
CMD ["/bin/folkloreserver"]
