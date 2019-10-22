# Dockerfile
FROM ubuntu:latest
MAINTAINER yhkim <yhki@innogrid.com>

RUN mkdir -p /violin/
WORKDIR /violin/

ADD GraphQL_violin /violin/
RUN chmod 755 /violin/violin

EXPOSE 8001

CMD ["/violin/violin"]
