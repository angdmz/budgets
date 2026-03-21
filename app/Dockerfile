# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy source code
COPY . .

# Accept build arguments for Auth0 configuration
ARG VITE_AUTH0_DOMAIN
ARG VITE_AUTH0_CLIENT_ID
ARG VITE_AUTH0_AUDIENCE

# Set environment variables for Vite build
ENV VITE_AUTH0_DOMAIN=${VITE_AUTH0_DOMAIN}
ENV VITE_AUTH0_CLIENT_ID=${VITE_AUTH0_CLIENT_ID}
ENV VITE_AUTH0_AUDIENCE=${VITE_AUTH0_AUDIENCE}

# Build the application
RUN npm run build

# Production stage
FROM nginx:alpine AS production

# Copy built files to nginx
COPY --from=builder /app/dist /usr/share/nginx/html/app

# Copy nginx configuration for SPA
RUN echo 'server { \
    listen 3001; \
    server_name localhost; \
    root /usr/share/nginx/html; \
    index index.html; \
    location ~ ^/app(/.*)?$ { \
        try_files $uri /app/index.html; \
    } \
}' > /etc/nginx/conf.d/default.conf

EXPOSE 3001

CMD ["nginx", "-g", "daemon off;"]
