FROM python:3.12-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    chromium \
    chromium-driver \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /tests

COPY requirements.txt .
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install --no-cache-dir -r requirements.txt

COPY . .

ENV PYTHONUNBUFFERED=1
ENV TEST_BASE_URL=http://nginx
ENV INTEGRATION_TESTS_SCREENSHOTS_DIR=/tests/screenshots

CMD ["pytest", "-v", "-s"]
