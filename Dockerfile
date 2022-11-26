# Build the manager binary
FROM gcr.io/distroless/static:nonroot as builder

WORKDIR /workspace

# Copy binary and charts
COPY ./bin/manager ./manager
COPY ./module-chart ./module-chart

# Copy the Go Modules manifests
#COPY go.mod go.sum ./

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
#RUN go mod download

# Copy the go source
#COPY . ./

# Build
#RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --chown=65532:65532 --from=builder /workspace/manager ./manager
COPY --chown=65532:65532 --from=builder /workspace/module-chart ./module-chart
USER 65532:65532

ENTRYPOINT ["/manager"]
