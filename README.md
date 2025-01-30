# qrack

## ✒️ Description

> Simple bruteforcer for CrackMe binaries

## ⚒️ To build

```sh
go build -o ./qrack
```

Or you can use just

```sh
just
```

## 📖 Usage

```sh
./qrack --dictionary <dictionary path> --binary <binary path> --pattern <flag pattern>
```

example:

```sh
./qrack --dictionary /usr/share/wordlists/rockyou.txt --binary ./example_crackme/test_crackme --pattern "Password"
```

