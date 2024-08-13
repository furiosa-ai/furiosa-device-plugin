FROM registry.corp.furiosa.ai/furiosa/libfuriosa-kubernetes:buildbase as build

# Build device-plugin binary
WORKDIR /
COPY . /
RUN make e2e-verification

FROM registry.corp.furiosa.ai/furiosa/libfuriosa-kubernetes:base

# Copy device plugin binary
COPY --from=build /verification /

CMD ["./verification"]
