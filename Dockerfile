ARG BASE_IMAGE=asia-northeast3-docker.pkg.dev/next-gen-infra/furiosa-ai/libfuriosa-kubernetes:v2026.1.1

FROM $BASE_IMAGE as build
ARG TARGETARCH

# Build device-plugin binary
WORKDIR /
COPY . /
RUN make build

# Stage arch-specific runtime libs under their final path so the distroless
# stage (no shell) can copy them verbatim. Below dynamic libraries are required
# due to `furiosa-smi` and Rust dependencies.
RUN set -e; \
    case "$TARGETARCH" in \
        amd64) libDir='x86_64-linux-gnu' ;; \
        arm64) libDir='aarch64-linux-gnu' ;; \
        *) echo >&2 "unsupported architecture: $TARGETARCH"; exit 1 ;; \
    esac; \
    mkdir -p /staging/usr/lib/$libDir; \
    cp /usr/lib/$libDir/libfuriosa_smi.so /staging/usr/lib/$libDir/; \
    cp /usr/lib/$libDir/libgcc_s.so.1     /staging/usr/lib/$libDir/

FROM gcr.io/distroless/base-debian12:latest

# Copy device plugin binary
WORKDIR /

# Arch-specific dynamic libs (libfuriosa_smi + Rust libgcc_s), landed in the
# running arch's multiarch dir via the staging tree built above.
COPY --from=build /staging/ /

COPY --from=build /main /main

CMD ["./main"]
