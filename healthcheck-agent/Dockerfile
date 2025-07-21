FROM alpine:3.20

# Install curl and jq
RUN apk add --no-cache bash curl jq

# Create app directory
WORKDIR /app

# Copy the startup script
COPY healthcheck-agent.sh /app/healthcheck-agent.sh

# Make sure the script is executable
RUN chmod +x /app/healthcheck-agent.sh

# Run the script
ENTRYPOINT ["/app/healthcheck-agent.sh"]
