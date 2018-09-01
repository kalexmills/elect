# STEP 1 build executable binary
FROM golang:alpine as builder

# STEP 2 build a small image
# start from scratch
FROM scratch

COPY --from=builder /etc/passwd /etc/passwd
# Copy our static executable
COPY . $GOPATH/bin/elect /go/bin/
USER appuser

EXPOSE 12345
ENTRYPOINT ["/go/bin/elect"]