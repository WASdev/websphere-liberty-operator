FROM registry.redhat.io/openshift4/ose-operator-registry:v4.14 AS builder
FROM registry.redhat.io/ubi8/ubi-minimal

# Add label for location of Declarative Config root directory & required OpenShift labels
ARG VERSION_LABEL=1.3.1
ARG RELEASE_LABEL=XX
ARG VCS_REF=0123456789012345678901234567890123456789
ARG VCS_URL="https://github.com/WASdev/websphere-liberty-operator"
ARG NAME="websphere-liberty-operator-catalog"
ARG SUMMARY="WebSphere Liberty Operator Catalog"
ARG DESCRIPTION="This image contains the catalog for WebSphere Liberty Operator."

# Set DC-specific label for the location of the DC root directory in the image
LABEL operators.operatorframework.io.index.configs.v1=/configs \
      name=$NAME \
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

## Copy Apache license
COPY LICENSE /licenses

USER root

# Pick up any latest fixes
RUN microdnf update && microdnf clean all

# Copy required tooling, licenses, and declarative config into defined location
COPY --from=builder /bin/opm /bin/opm
COPY --from=builder /bin/grpc_health_probe /bin/grpc_health_probe

ARG ARCH=linux-amd64
RUN mkdir /catalog
# Note: the COPY directive can also point to a directory structure and it will recurse thru the directory structure and use any yaml/json files it locates
COPY --chown=1001:0 catalog /catalog
RUN cp -r "catalog/linux-$ARCH" /configs

# Validate catalog file
RUN ["/bin/opm", "validate", "/configs"]

EXPOSE 50051

USER 1001

WORKDIR /tmp
ENTRYPOINT ["/bin/opm"]
CMD ["serve", "/configs", "--cache-dir=/tmp/cache"]
RUN ["/bin/opm", "serve", "/configs", "--cache-dir=/tmp/cache",  "--cache-only"]