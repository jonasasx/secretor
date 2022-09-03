FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY bin/secretor .
RUN ls -l /
USER 65532:65532

ENTRYPOINT ["/secretor"]
