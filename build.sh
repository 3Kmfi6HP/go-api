#!/usr/bin/env bash

echo "buildAll Api..."
os_archs="darwin:amd64 darwin:arm64 freebsd:386 freebsd:amd64 linux:386 linux:amd64 linux:arm linux:arm64 windows:386 windows:amd64 linux:mips64 linux:mips64le linux:mips:softfloat linux:mipsle:softfloat"
LDFLAGS="-s -w"
for n in $os_archs
do
    os=$(echo $n | cut -d : -f 1)
    arch=$(echo $n | cut -d : -f 2)
    gomips=$(echo $n | cut -d : -f 3)
    target_suffix="${os}_${arch}"
    echo "Build ${os}_${arch} ...."
    env CGO_ENABLED=0 GOOS="${os}" GOARCH="${arch}" GOMIPS="${gomips}" go build -trimpath -ldflags "${LDFLAGS}" -o ./release/Api_"${target_suffix}" .
    echo "Build ${os}_${arch} done"
done

mv ./release/Api_windows_386 ./release/Api_windows_386.exe
mv ./release/Api_windows_amd64 ./release/Api_windows_amd64.exe