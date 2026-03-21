import pytest
from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.chrome.service import Service
import os
from dotenv import load_dotenv

load_dotenv()

@pytest.fixture(scope="session")
def base_url():
    """Base URL for the application"""
    return os.getenv("TEST_BASE_URL", "http://localhost:8000")

@pytest.fixture(scope="session")
def auth0_test_user():
    """Auth0 test user credentials"""
    return {
        "email": os.getenv("AUTH0_TEST_EMAIL"),
        "password": os.getenv("AUTH0_TEST_PASSWORD")
    }

@pytest.fixture(scope="function")
def driver():
    """Create a Chrome WebDriver instance"""
    chrome_options = Options()
    chrome_options.add_argument("--headless")
    chrome_options.add_argument("--no-sandbox")
    chrome_options.add_argument("--disable-dev-shm-usage")
    chrome_options.add_argument("--disable-gpu")
    chrome_options.add_argument("--window-size=1920,1080")
    chrome_options.add_argument("--disable-extensions")
    chrome_options.add_argument("--disable-software-rasterizer")
    
    # Use Chromium binary (for Docker compatibility)
    chrome_options.binary_location = "/usr/bin/chromium"
    
    # Use chromedriver from PATH (installed in Docker)
    driver = webdriver.Chrome(options=chrome_options)
    driver.implicitly_wait(10)
    
    yield driver
    
    driver.quit()

@pytest.fixture(scope="function")
def driver_visible():
    """Create a visible Chrome WebDriver instance for debugging"""
    chrome_options = Options()
    chrome_options.add_argument("--window-size=1920,1080")
    chrome_options.add_argument("--no-sandbox")
    chrome_options.add_argument("--disable-dev-shm-usage")
    
    driver = webdriver.Chrome(options=chrome_options)
    driver.implicitly_wait(10)
    
    yield driver
    
    driver.quit()
