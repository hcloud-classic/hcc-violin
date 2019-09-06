# Dockerfile
FROM ubuntu:latest
MAINTAINER yhkim <yhki@innogrid.com>

RUN mkdir -p /GraphQL_violin/
WORKDIR /GraphQL_violin/

ADD GraphQL_violin /GraphQL_violin/
RUN chmod 755 /GraphQL_violin/GraphQL_violin

EXPOSE 8001

CMD ["/GraphQL_violin/GraphQL_violin"]
