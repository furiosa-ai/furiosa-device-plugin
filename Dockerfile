ARG BASE_IMAGE=registry.corp.furiosa.ai/furiosa/libfuriosa-kubernetes:latest

FROM $BASE_IMAGE as build

# Build device-plugin binary
WORKDIR /
COPY . /
RUN make build

FROM gcr.io/distroless/base-debian12:nonroot

# Copy device plugin binary
WORKDIR /

# Below dynamic libraries are required.
# $ ldd /main
#     libfuriosa_smi.so => /lib/x86_64-linux-gnu/libfuriosa_smi.so (0x00007fffff4b0000)
#     libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x00007fffff2cf000)
#     libgcc_s.so.1 => /lib/x86_64-linux-gnu/libgcc_s.so.1 (0x00007fffff2af000)
#     libm.so.6 => /lib/x86_64-linux-gnu/libm.so.6 (0x00007fffff1d0000)
#     /lib64/ld-linux-x86-64.so.2 (0x00007ffffffcc000)
COPY --from=build /usr/lib/x86_64-linux-gnu/libfuriosa_smi.so /usr/lib/x86_64-linux-gnu/libfuriosa_smi.so
COPY --from=build /usr/lib/x86_64-linux-gnu/libc.so.6 /usr/lib/x86_64-linux-gnu/libc.so.6
COPY --from=build /usr/lib/x86_64-linux-gnu/libgcc_s.so.1 /usr/lib/x86_64-linux-gnu/libgcc_s.so.1
COPY --from=build /usr/lib/x86_64-linux-gnu/libm.so.6 /usr/lib/x86_64-linux-gnu/libm.so.6

COPY --from=build /main /main

CMD ["./main"]
