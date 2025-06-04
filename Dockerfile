FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.24 as build

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG Version

WORKDIR /go/src/github.com/ihatemodels/alcatraz-rest
COPY . .

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 \
  cd cmd/server && \
  go build \
  -tags osusergo,netgo \
  -ldflags "-s -w -X main.version=${Version}" \
  -o /usr/bin/alcatraz-rest .

FROM --platform=${BUILDPLATFORM:-linux/amd64} gcr.io/distroless/static-debian12:latest

LABEL org.opencontainers.image.source=https://github.com/ihatemodels/alcatraz-rest
LABEL org.opencontainers.image.version=${Version}
LABEL org.opencontainers.image.authors="admins@ihatemodels.com"
LABEL org.opencontainers.image.title="Alcatraz rest"
LABEL org.opencontainers.image.description="Alcatraz rest"

COPY --from=build /usr/bin/alcatraz-rest /
EXPOSE 8080
ENTRYPOINT ["/alcatraz-rest"]