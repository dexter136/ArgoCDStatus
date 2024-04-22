FROM golang:1.22.2-bullseye as base

COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /statuspage

FROM scratch
COPY --from=base /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=base /etc/passwd /etc/passwd
COPY --from=base /etc/group /etc/group
COPY --from=base /statuspage .
COPY status.tmpl ./status.tmpl

LABEL org.opencontainers.image.source https://github.com/dexter136/ArgoCDStatus

CMD [ "./statuspage" ]
