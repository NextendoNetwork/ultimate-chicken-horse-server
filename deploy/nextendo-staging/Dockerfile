FROM scratch
COPY --chmod=0555 uch-server /uch-server
USER 65532:65532
ENTRYPOINT ["/uch-server"]
CMD ["-config", "/run/uch/config.json", "-addr", "0.0.0.0:8443", "-go-transport", "-udp-addr", "0.0.0.0:19889", "-stats-addr", "127.0.0.1:8089"]
