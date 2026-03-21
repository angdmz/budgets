FROM python:3.11-slim

# Install Chromium and ChromeDriver
RUN apt-get update && apt-get install -y \
    chromium \
    chromium-driver \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /tests

# Copy requirements and install
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy test files
COPY . .

# Set environment variables
ENV PYTHONUNBUFFERED=1
ENV TEST_BASE_URL=http://nginx

# Run tests
CMD ["pytest", "-v", "-s"]
