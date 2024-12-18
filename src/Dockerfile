FROM heroiclabs/nakama-pluginbuilder:3.25.0 AS builder

ENV GO111MODULE=on
ENV CGO_ENABLED=1

WORKDIR /backend
COPY . .

RUN apt-get update && \
    apt-get -y upgrade && \
    apt-get install -y --no-install-recommends gcc libc6-dev

RUN go build --trimpath --gcflags "all=-N -l" --mod=vendor --buildmode=plugin -o ./backend.so
RUN go install github.com/go-delve/delve/cmd/dlv@latest

FROM registry.heroiclabs.com/heroiclabs/nakama-dsym:3.25.0

COPY --from=builder /go/bin/dlv /nakama
COPY --from=builder /backend/backend.so /nakama/data/modules
COPY --from=builder /backend/local.yml /nakama/data/
COPY --from=builder /backend/*.json /nakama/data/modules