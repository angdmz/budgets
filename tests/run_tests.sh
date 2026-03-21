#!/bin/bash

# Run integration tests with docker-compose

echo "🧪 Running Integration Tests..."
echo ""

# Ensure the system is running
echo "📋 Checking if services are running..."
cd ..
docker-compose ps | grep -q "Up" || {
    echo "⚠️  Services not running. Starting services..."
    docker-compose up -d
    echo "⏳ Waiting for services to be ready..."
    sleep 10
}

echo "✅ Services are running"
echo ""

# Run the integration tests
echo "🚀 Running tests..."
docker-compose --profile integration run --rm integration-tests

# Check exit code
if [ $? -eq 0 ]; then
    echo ""
    echo "✅ All tests passed!"
    echo "📊 Test report available at: tests/report.html"
    echo "📸 Screenshots available at: tests/screenshots/"
else
    echo ""
    echo "❌ Some tests failed"
    echo "📊 Check test report at: tests/report.html"
    echo "📸 Check screenshots at: tests/screenshots/"
    exit 1
fi
