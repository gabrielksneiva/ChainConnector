# Build stage
FROM node:20-alpine as build

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies from the lockfile for reproducible builds
RUN npm ci && \
    npm cache clean --force

# Copy source code
COPY . .

# Build application
RUN npm run build

# Production stage
FROM nginx:alpine3.20

# Install curl for healthchecks and create necessary directories
RUN apk add --no-cache curl && rm -rf /var/cache/apk/* && \
    mkdir -p /var/run/nginx && \
    mkdir -p /var/cache/nginx && \
    chown -R nginx:nginx /var/run/nginx /var/cache/nginx /var/log/nginx

# Copy nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built application from build stage
COPY --from=build /app/dist /usr/share/nginx/html

# Fix permissions for nginx user and run as non-root
RUN mkdir -p /var/run/nginx /var/cache/nginx && \
    chown -R nginx:nginx /var/run/nginx /var/cache/nginx /var/log/nginx /usr/share/nginx/html

# Replace default nginx main config to use a rootless PID path
COPY nginx.main.conf /etc/nginx/nginx.conf

USER nginx

EXPOSE 80

HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost --silent --show-error || exit 1

CMD ["nginx", "-g", "daemon off;"]
