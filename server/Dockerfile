FROM golang:1.22 as base

ENV HOST=0.0.0.0
ENV PORT=8080
ENV SQITCH_CONN=
ENV DB_CONN=
ENV AUTH0_DOMAIN=
ENV AUTH0_AUDIENCE=
ENV AUTH0_CLIENT_ID=
ENV AUTH0_CLIENT_SECRET=
ENV SPICEDB_ENDPOINT=
ENV SPICEDB_TOKEN=

RUN apt-get update -y
RUN apt-get install -y sqitch libdbd-pg-perl postgresql-client
RUN curl https://apt.fury.io/authzed/gpg.key | apt-key add -
RUN sh -c 'echo "deb https://apt.fury.io/authzed/ * *" > /etc/apt/sources.list.d/fury.list'
RUN apt-get update && apt-get install -y zed

# Update certificates
RUN update-ca-certificates 2>/dev/null || true

WORKDIR /app
ENV CGO_ENABLED=0
COPY go.* .
RUN go mod download
COPY . .

FROM base AS build
RUN --mount=type=cache,target=/root/.cache/go-build \
  go build -o server

FROM base AS unit-test
RUN --mount=type=cache,target=/root/.cache/go-build \
  go test -v

FROM base AS development
RUN --mount=type=cache,target=/root/.cache/go-build \
  go install github.com/air-verse/air@latest 
RUN --mount=type=cache,target=/root/.cache/go-build \
  go install github.com/go-delve/delve/cmd/dlv@latest

COPY sqitch.conf sqitch.plan ./
COPY deploy ./deploy
COPY revert ./revert
COPY verify ./verify
COPY authz ./authz
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE ${PORT}
EXPOSE 2345

CMD sqitch deploy ${SQITCH_CONN} && \
  zed schema write ./authz/schema.zed --endpoint $SPICEDB_ENDPOINT --insecure --token $SPICEDB_TOKEN && \
  air -c .air.toml

FROM base AS production

WORKDIR /
COPY sqitch.conf sqitch.plan ./
COPY deploy ./deploy
COPY revert ./revert
COPY verify ./verify
COPY authz ./authz
COPY --from=build /app/server server
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE ${PORT}

CMD sqitch deploy ${SQITCH_CONN} && \
  zed schema write ./authz/schema.zed --endpoint $SPICEDB_ENDPOINT --insecure --token $SPICEDB_TOKEN && \
  ./server
