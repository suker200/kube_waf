FROM alpine

COPY kube_waf /

RUN chmod +x /kube_waf

WORKDIR /

ENTRYPOINT ["./kube_waf"]