FROM busybox:latest

WORKDIR /app
COPY --chmod=555 snmpflapd /app/snmpflapd
CMD ["./snmpflapd"]
