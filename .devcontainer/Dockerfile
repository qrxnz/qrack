FROM golang:1.26-bookworm

# Install build-essential for gcc and curl for downloading the wordlist
RUN apt-get update && apt-get install -y \
    build-essential \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go mod and sum files first to leverage Docker cache
# This also serves as a check for the correct build context
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the Go application (qrack)
RUN go build -o qrack .

# Build the example crackme
RUN gcc example_crackme/test_crackme.c -o example_crackme/test_crackme

# Use local rockyou.txt wordlist from the repository
RUN mkdir -p /usr/share/wordlists && \
    tar -xzf example_crackme/rockyou/rockyou.txt.tar.gz -C /usr/share/wordlists

# Ensure the crackme is executable
RUN chmod +x example_crackme/test_crackme

# Default command to run qrack with the dictionary and the binary
ENTRYPOINT ["./qrack"]
CMD ["--dictionary", "/usr/share/wordlists/rockyou.txt", "--binary", "./example_crackme/test_crackme", "--pattern", "Password correct!"]
