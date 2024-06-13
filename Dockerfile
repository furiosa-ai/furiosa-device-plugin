FROM ghcr.io/furiosa-ai/libfuriosa-kubernetes:latest as build

# Build device-plugin binary
WORKDIR /
COPY . /
RUN make build

FROM ghcr.io/furiosa-ai/libfuriosa-kubernetes:latest

# Copy device plugin binary
WORKDIR /
COPY --from=build /main /main
CMD ["./main"]
