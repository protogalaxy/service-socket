FROM debian:wheezy

MAINTAINER The Protogalaxy Project

EXPOSE 10100

ENTRYPOINT ["./main", "-logtostderr", "-v=4"]

COPY ./target/bin/main .
