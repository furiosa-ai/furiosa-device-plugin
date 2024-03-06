FROM ghcr.io/furiosa-ai/libfuriosa-kubernetes:latest as build

# Build device-plugin binary
WORKDIR /
COPY . /
RUN make build

FROM gcr.io/distroless/base-debian12

# Copy hwloc binaries and libraries from the builder stage
COPY --from=build /usr/local/lib/ /usr/local/lib/
COPY --from=build /usr/local/include/ /usr/local/include/

# Configure env values
ENV C_INCLUDE_PATH /usr/local/include:$C_INCLUDE_PATH
ENV LD_LIBRARY_PATH usr/local/lib:$LD_LIBRARY_PATH

# Copy device plugin binary
COPY --from=build /main /

CMD ["./main"]
