source develop.env

function cleanup() {
    rm -f webhook-receiver
}
trap cleanup EXIT


# Compile Go
GO111MODULE=on GOGC=off go build -mod=vendor -v -o webhook-receiver ./cmd/api/
./webhook-receiver
