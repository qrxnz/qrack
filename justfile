watch:
    watchexec -r just run

build:
    go build .

run: build
    ./qrack
