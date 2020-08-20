FROM go:1.15.0 as BUILD

WORKDIR /build

COPY . ./

RUN go build

FROM go:1.15.0

WORKDIR /app

COPY --from=BUILD /build/walletdb app

CMD ["./app"]
