FROM golang:latest

WORKDIR /app

# Install Docker
RUN apt-get update && apt-get install -y apt-transport-https ca-certificates curl gnupg lsb-release
RUN curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
RUN echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
RUN apt-get update && apt-get install -y docker-ce-cli

# Mount Docker socket as a volume
VOLUME /var/run/docker.sock

RUN apt-get update && apt-get install -y curl
RUN curl -o /usr/local/bin/cog -L "https://github.com/replicate/cog/releases/latest/download/cog_$(uname -s)_$(uname -m)"
RUN chmod +x /usr/local/bin/cog


COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

COPY .env .env

RUN go build -o main .


CMD ["./streamlining-backend"]
