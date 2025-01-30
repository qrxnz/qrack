# qrack

## ✒️ Description

> Simple bruteforcer for CrackMe binaries

https://github.com/user-attachments/assets/8ad86144-5839-4a16-ad60-f6797c90dd6b

## ⚒️ To build

```sh
go build -o ./qrack
```

Or you can use just:

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

