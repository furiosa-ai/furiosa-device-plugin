FROM registry.corp.furiosa.ai/furiosa/libfuriosa-kubernetes:latest as build

# Build device-plugin binary
WORKDIR /
COPY . /
RUN make build

FROM registry.corp.furiosa.ai/furiosa/libfuriosa-kubernetes:latest

# Copy device plugin binary
WORKDIR /
COPY --from=build /main /main
CMD ["./main"]
