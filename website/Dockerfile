# Multi-stage Docker build for Nestor website

# Build stage
FROM hugomods/hugo:exts AS builder

# Set working directory
WORKDIR /src

# Copy Hugo site files
COPY . .

# Build the Hugo site
RUN hugo --minify --cleanDestinationDir

# Production stage
FROM nginx:alpine

# Copy built site from builder stage
COPY --from=builder /src/public /usr/share/nginx/html

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Expose port 80
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]