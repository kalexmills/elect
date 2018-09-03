FROM golang:alpine as builder

FROM scratch

COPY --from=builder /etc/passwd /etc/passwd
# Copy our static executable
COPY . $GOPATH/bin/elect /go/bin/
USER appuser

EXPOSE 12345
ENTRYPOINT ["/go/bin/elect"]