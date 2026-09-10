# Builds the calculator API as a single static binary and ships it in an empty
# image. Build context is ./backend (see compose.yaml).

FROM golang:1.27-alpine AS build

WORKDIR /src

# No `go mod download` step: go.mod requires no third-party modules (the backend
# is standard library only), so there is nothing to cache separately.
COPY . .

# CGO_ENABLED=0 links the binary statically, so it needs no libc at runtime and
# the final stage can be empty. -s -w drop the symbol and DWARF tables.
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/server ./cmd/server

# scratch is the empty base image: the result contains the binary and nothing
# else -- no shell, no package manager, no CA bundle (none is needed; the server
# makes no outbound calls).
FROM scratch

COPY --from=build /out/server /server

# 65534 is `nobody`. Nothing in this image needs root.
USER 65534:65534

EXPOSE 8080

ENTRYPOINT ["/server"]
