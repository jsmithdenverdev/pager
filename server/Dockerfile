FROM golang:1.22 as base

# Update certificates
RUN update-ca-certificates 2>/dev/null || true

WORKDIR /app
ENV CGO_ENABLED=0
COPY go.* .
RUN go mod download
COPY . .

FROM base as build
RUN --mount=type=cache,target=/root/.cache/go-build \
  go build -o server

FROM base as unit-test
RUN --mount=type=cache,target=/root/.cache/go-build \
  go test -v

FROM base as development
RUN apt-get update -y
RUN apt-get install sqitch libdbd-pg-perl postgresql-client -y

RUN --mount=type=cache,target=/root/.cache/go-build \
  go install github.com/cosmtrek/air@latest 
RUN --mount=type=cache,target=/root/.cache/go-build \
  go install github.com/go-delve/delve/cmd/dlv@latest

ENV HOST=0.0.0.0
ENV PORT=8080
ENV SQITCH_CONN=
ENV DB_CONN=
ENV AUTH0_DOMAIN=
ENV AUTH0_AUDIENCE=

COPY sqitch.conf sqitch.plan ./
COPY deploy ./deploy
COPY revert ./revert
COPY verify ./verify
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE ${PORT}
EXPOSE 2345

CMD sqitch deploy ${SQITCH_CONN} && air -c .air.toml

FROM base as production
RUN apt-get update -y
RUN apt-get install sqitch libdbd-pg-perl postgresql-client -y

ENV HOST=0.0.0.0
ENV PORT=8080
ENV SQITCH_CONN=
ENV DB_CONN=

WORKDIR /
COPY sqitch.conf sqitch.plan ./
COPY deploy ./deploy
COPY revert ./revert
COPY verify ./verify
COPY --from=build /app/server server
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE ${PORT}

CMD sqitch deploy ${SQITCH_CONN} && ./server
