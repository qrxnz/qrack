# qrack

[![Go Workflow](https://github.com/qrxnz/qrack/actions/workflows/go.yml/badge.svg)](https://github.com/qrxnz/qrack/actions/workflows/go.yml)

## ✒️ Description

> Simple bruteforcer for CrackMe binaries / CTF challegne solver

qrack is a simple bruteforcer for cracking simple binary executable files, commonly known as "CrackMe" challenges. It features a user-friendly terminal interface built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

-   📺 Demo

https://github.com/user-attachments/assets/f0c02036-e36c-4024-b16a-bacc6e02126f

## 📦 Installation

### Build from source

To build the project, you need to have Go installed.

```sh
go build .
```

Alternatively, if you have `just` installed, you can simply run:

```sh
just build
```

### Using Nix ❄️

-   Try it without installing:

```sh
nix run github:qrxnz/qrack
```

-   Installation:

Add input in your flake like:

```nix
{
 inputs = {
   nveem = {
     url = "github:qrxnz/qrack";
     inputs.nixpkgs.follows = "nixpkgs";
   };
 };
}
```

With the input added you can reference it directly:

```nix
{ inputs, system, ... }:
{
  # NixOS
  environment.systemPackages = [ inputs.qrack.packages.${pkgs.system}.default ];
  # home-manager
  home.packages = [ inputs.qrack.packages.${pkgs.system}.default ];
}
```

or

You can install this package imperatively with the following command:

```nix
nix profile install github:qrxnz/qrack
```

## 📖 Usage

Run the application with the following command, providing the necessary flags.

```sh
./qrack --dictionary <path> --binary <path> [flags]
```

### Flags

| Flag            | Description                                    | Default             | Required |
| --------------- | ---------------------------------------------- | ------------------- | -------- |
| `--dictionary`  | Path to the dictionary file (wordlist).        |                     | Yes      |
| `--binary`      | Path to the binary executable to crack.        |                     | Yes      |
| `--pattern`     | The success pattern to look for in the output. | "Password correct!" | No       |
| `--concurrency` | Number of concurrent workers to use.           | 4                   | No       |

### Example

```sh
./qrack \
  --dictionary /usr/share/wordlists/rockyou.txt \
  --binary ./example_crackme/test_crackme \
  --pattern "Password" \
  --concurrency 8
```

## 📜 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
