# Build the manager binary
FROM golang:1.20 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY utils/ utils/

# Build
RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -ldflags="-s -w" -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ARG USER_ID=65532
ARG GROUP_ID=65532

ARG VERSION_LABEL=1.2.1
ARG RELEASE_LABEL=XX
ARG VCS_REF=0123456789012345678901234567890123456789
ARG VCS_URL="https://github.com/WASdev/websphere-liberty-operator"
ARG NAME="websphere-liberty-operator"
ARG SUMMARY="WebSphere Liberty Operator"
ARG DESCRIPTION="This image contains the controllers for WebSphere Liberty Operator."

LABEL name=$NAME \
      vendor=IBM \
      version=$VERSION_LABEL \
      release=$RELEASE_LABEL \
      description=$DESCRIPTION \
      summary=$SUMMARY \
      io.k8s.display-name=$SUMMARY \
      io.k8s.description=$DESCRIPTION \
      vcs-type=git \
      vcs-ref=$VCS_REF \
      vcs-url=$VCS_URL \
      url=$VCS_URL

COPY LICENSE /licenses/
WORKDIR /
COPY --from=builder /workspace/manager .

USER ${USER_ID}:${GROUP_ID}

ENTRYPOINT ["/manager"]
