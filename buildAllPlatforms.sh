#!/bin/bash

OUTPUT="bin/"
PROGRAM="mnh"
LDFLAGS="-s -w"

ver="$(git describe --tags --match "v*" --dirty="" 2>/dev/null || git log -1 --pretty=format:"v0.0.0-%h" 2>/dev/null || echo "v0.0.0")"
[ -n "$(git status --porcelain |& grep -Ev '^\?\?')" ] && ver="$ver-$(date +"%Y%M%d-%H%m%S")"
LDFLAGS="$LDFLAGS -X main.version=$ver"

mkdir -p "$OUTPUT"
rm -f "$OUTPUT/${PROGRAM}-"*

platforms=(
	linux/386
	linux/amd64
	linux/arm
	linux/arm64
	linux/mips/softfloat
	linux/mips64
	linux/mips64le
	linux/mipsle/softfloat
	windows/386
	windows/amd64
	windows/arm
)
# platforms=($(go tool dist list))

for i in "${platforms[@]}"; do
	os="$(echo "$i" | awk -F/ '{print $1}')"
	arch="$(echo "$i" | awk -F/ '{print $2}')"
	mips="$(echo "$i" | awk -F/ '{print $3}')"
	if [ "$os" == "windows" ]; then
		suffix="${os}-${arch}.exe"
	else
		suffix="${os}-${arch}"
	fi
	filename="$OUTPUT/${PROGRAM}-$suffix"
	echo "build $filename for $i"
	CGO_ENABLED=0 GOOS=$os GOARCH=$arch GOMIPS=$mips go build -trimpath -ldflags "$LDFLAGS" -o "$filename"
done