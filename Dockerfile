FROM golang:1.12.4 as builder

RUN mkdir -p /argo/
WORKDIR /argo
COPY ./ /argo

# Build argo
RUN make

FROM scratch

# Copy the argo binary generated in the builder layer
COPY --from=builder /argo/argo .
