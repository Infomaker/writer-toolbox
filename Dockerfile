FROM alpine

MAINTAINER tobias.sodergren@infomaker.se

RUN apk --update upgrade && apk add ca-certificates && rm -rf /var/cache/apk/*

ADD writer-tool /

ENTRYPOINT ["/writer-tool"]
