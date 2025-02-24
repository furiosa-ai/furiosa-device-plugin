ARG BASE_IMAGE=registry.corp.furiosa.ai/furiosa/libfuriosa-kubernetes:latest

FROM $BASE_IMAGE as build

# Build device-plugin binary
WORKDIR /
COPY . /
RUN make e2e-verification

FROM registry.corp.furiosa.ai/furiosa/libfuriosa-kubernetes:latest

# Copy device plugin binary
COPY --from=build /verification /verification
WORKDIR /
CMD ["./verification"]
