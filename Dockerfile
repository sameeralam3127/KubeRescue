FROM golang:1.26 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags "-s -w -X github.com/sameeralam3127/kuberescue/internal/cli.version=${VERSION}" \
    -o /out/kuberescue ./cmd/kuberescue

FROM gcr.io/distroless/static:nonroot

COPY --from=build /out/kuberescue /kuberescue

USER nonroot:nonroot

ENTRYPOINT ["/kuberescue"]
CMD ["--help"]
