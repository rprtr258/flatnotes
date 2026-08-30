FROM golang:1.27 AS build
WORKDIR /build
RUN apt update && \
  apt install unzip
RUN curl -fsSL https://bun.sh/install | bash
ENV PATH="/root/.bun/bin:${PATH}"
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile
COPY flatnotes/src ./flatnotes/src
RUN bun run build
COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./
RUN go build -o /app ./cmd/main.go

FROM debian:12.2
ENV PUID=1000
ENV PGID=1000
ENV FLATNOTES_PATH=/data
RUN mkdir -p ${FLATNOTES_PATH}
RUN apt update && \
  rm -rf /var/lib/apt/lists/*
WORKDIR /app
# COPY flatnotes ./flatnotes
COPY --from=build /build/flatnotes/dist ./flatnotes/dist
COPY --from=build /app ./app
VOLUME /data
EXPOSE 8080/tcp
ENTRYPOINT [ "/app/app" ]
