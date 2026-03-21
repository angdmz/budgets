# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy source code
COPY . .

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
    location /app/ { \
        try_files $uri $uri/ /app/index.html; \
    } \
    location = /app { \
        return 301 /app/; \
    } \
}' > /etc/nginx/conf.d/default.conf

EXPOSE 3001

CMD ["nginx", "-g", "daemon off;"]
