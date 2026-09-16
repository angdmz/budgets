FROM nginx:1.27-alpine

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Expose ports 80 (HTTP redirect) and 443 (TLS termination)
EXPOSE 80 443

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
