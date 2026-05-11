# Build stage
FROM node:20-alpine as build

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies with ci for reproducibility
RUN npm ci --only=production && \
    npm ci --only=dev && \
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

# Fix permissions for nginx user
RUN mkdir -p /var/run/nginx /var/cache/nginx && \
    chown -R nginx:nginx /var/run/nginx /var/cache/nginx /var/log/nginx /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost --silent --show-error || exit 1

CMD ["nginx", "-g", "daemon off;"]