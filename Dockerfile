FROM golang:latest as builder
WORKDIR /app/auto-reply
RUN set -ex \
  && apt-get update -y \
  && apt-get upgrade -y \
  && apt-get install make
COPY go.* /app/auto-reply/
RUN go mod download
COPY . .
RUN make -j10 cibuild

FROM scratch
COPY --from=builder bin/ /bin/.
