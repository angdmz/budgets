# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

COPY package*.json ./
# Install dependencies
RUN npm install

COPY . .
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
    location /admin/ { \
        try_files $uri $uri/ /admin/index.html; \
    } \
    location = /admin { \
        return 301 /admin/; \
    } \
}' > /etc/nginx/conf.d/default.conf

EXPOSE 3002

CMD ["nginx", "-g", "daemon off;"]
