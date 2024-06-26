FROM golang:1.22.4-alpine3.20

# Go to a temporary directory until we install all the dependencies
RUN mkdir -p /app/temp
WORKDIR /app/temp

# Install the necessary packages
RUN apk update && apk add --no-cache \
    bash=5.2.26-r0 \
    wget=1.24.5-r0 \
    git=2.45.2-r0 \
    openssh=9.7_p1-r3 \
    rclone=1.66.0-r3

# Install task
RUN wget https://github.com/go-task/task/releases/download/v3.34.1/task_linux_amd64.tar.gz && \
    tar -xzf task_linux_amd64.tar.gz && \
    mv ./task /usr/local/bin/task && \
    chmod 777 /usr/local/bin/task

# Delete the temporary directory and go to the app directory
RUN rm -rf /app/temp
WORKDIR /app

# Copy and install go dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the rest of the files
COPY . .

# Build the app
RUN task build

# Expose the port 22 and run the app
EXPOSE 22
CMD ["task", "serve"]
