ARG BASE_IMAGE=registry.corp.furiosa.ai/furiosa/libfuriosa-kubernetes:latest

FROM $BASE_IMAGE as build

# Build device-plugin binary
WORKDIR /
COPY . /
RUN make build

FROM gcr.io/distroless/base-debian12:latest

# Copy device plugin binary
WORKDIR /

# Below dynamic libraries are required due to `furiosa-smi` and Rust dependencies.
COPY --from=build /usr/lib/x86_64-linux-gnu/libfuriosa_smi.so /usr/lib/x86_64-linux-gnu/libfuriosa_smi.so
COPY --from=build /usr/lib/x86_64-linux-gnu/libgcc_s.so.1 /usr/lib/x86_64-linux-gnu/libgcc_s.so.1

COPY --from=build /main /main

CMD ["./main"]
