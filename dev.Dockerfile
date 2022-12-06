FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --chown=65532:65532 ./bin/manager ./manager
COPY --chown=65532:65532 ./module-chart ./module-chart
USER 65532:65532

ENTRYPOINT ["/manager"]
