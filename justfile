watch:
    watchexec -r just run

build:
    go build -o ./qrack

run: build
    ./qrack
