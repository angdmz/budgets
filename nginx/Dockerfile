FROM nginx:alpine

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Expose port 80 (SSL is terminated by host Nginx)
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
