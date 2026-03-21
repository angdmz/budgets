# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

COPY package*.json ./
# Install dependencies
RUN npm install

COPY . .

# Accept build arguments for Auth0 configuration
ARG VITE_AUTH0_DOMAIN
ARG VITE_AUTH0_CLIENT_ID
ARG VITE_AUTH0_AUDIENCE

# Set environment variables for Vite build
ENV VITE_AUTH0_DOMAIN=${VITE_AUTH0_DOMAIN}
ENV VITE_AUTH0_CLIENT_ID=${VITE_AUTH0_CLIENT_ID}
ENV VITE_AUTH0_AUDIENCE=${VITE_AUTH0_AUDIENCE}

RUN npm run build

# Production stage
FROM nginx:alpine AS production

COPY --from=builder /app/dist /usr/share/nginx/html/admin

# Copy nginx configuration for SPA
RUN echo 'server { \
    listen 3002; \
    server_name localhost; \
    root /usr/share/nginx/html; \
    index index.html; \
    location ~ ^/admin(/.*)?$ { \
        try_files $uri /admin/index.html; \
    } \
}' > /etc/nginx/conf.d/default.conf

EXPOSE 3002

CMD ["nginx", "-g", "daemon off;"]
