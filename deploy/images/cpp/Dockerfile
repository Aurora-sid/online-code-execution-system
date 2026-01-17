FROM alpine:latest

# Install GCC and G++
RUN apk add --no-cache gcc g++ musl-dev

WORKDIR /app

# The container will stay alive waiting for exec commands
CMD ["sleep", "infinity"]
