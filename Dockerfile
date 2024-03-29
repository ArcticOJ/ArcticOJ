FROM --platform=${BUILDPLATFORM} node:20-alpine AS avalanche-builder
WORKDIR /usr/src/avalanche

COPY avalanche .

RUN yarn install --frozen-lockfile --immutable

RUN yarn build

FROM --platform=${BUILDPLATFORM} golang:alpine AS arctic-builder
WORKDIR /usr/src/app

COPY . .

COPY --from=avalanche-builder /usr/src/avalanche/out avalanche/out

RUN go mod download

RUN make release OUT=./out/arctic

FROM --platform=${BUILDPLATFORM} alpine as arctic
WORKDIR /arctic

COPY --from=arctic-builder /usr/src/app/out/arctic ./

ENTRYPOINT ["/arctic/arctic"]

