FROM ghcr.io/furiosa-ai/libfuriosa-kubernetes:buildbase as build

# Build device-plugin binary
WORKDIR /
COPY . /
RUN make build

FROM ghcr.io/furiosa-ai/libfuriosa-kubernetes:base

# Copy device plugin binary
COPY --from=build /main /

CMD ["./main"]
