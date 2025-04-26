FROM golang:latest as builder

WORKDIR /build

COPY . .

RUN go install tailscale.com/cmd/derper@latest

FROM debian:bookworm-slim

WORKDIR /app

ENV DERP_HOST=127.0.0.1
ENV CERT_DIR=/app/certs
ENV CONF_FILE=/app/san.conf
ENV DERP_PORT=8888
ENV STUN_PORT=8889
ENV HTTP_PORT=-1

RUN apt-get update && \
	apt-get install -y openssl curl libc6

COPY --from=builder /go/bin/derper /app/derper
COPY --from=builder /build/gen_cert.sh /app/gen_cert.sh

CMD /bin/bash /app/gen_cert.sh $DERP_HOST $CERT_DIR $CONF_FILE && \
	/app/derper -a :$DERP_PORT \
	-http-port $HTTP_PORT \
	-stun-port $STUN_PORT \
	-hostname $DERP_HOST \
	-certmode manual \
	-certdir /app/certs \
	-verify-clients
