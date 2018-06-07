FROM alpine

RUN apk update && apk add ca-certificates

COPY waf_agent /
RUN chmod +x /waf_agent

ENTRYPOINT ["./waf_agent"]
