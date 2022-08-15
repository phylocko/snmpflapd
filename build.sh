flags="-X main.version=$(cat VERSION) -X 'main.build=$(date -R)'"
echo Building with flags $flags

go build -ldflags "-X main.version=$(cat VERSION) -X 'main.build=$(date -R)'"
