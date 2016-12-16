#  grafana_bootstrap
FROM scratch

# Add ca-certificates.crt
ADD ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Add executable
ADD grafana_bootstrap /

ENTRYPOINT [ "/grafana_bootstrap" ]
