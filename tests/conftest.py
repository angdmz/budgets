import uuid
import pytest
import requests
from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.chrome.service import Service
import os
from dotenv import load_dotenv
from settings import settings

load_dotenv()


def _get_mgmt_token(domain: str, client_id: str, client_secret: str) -> str:
    """Obtain an Auth0 Management API access token via client credentials grant."""
    url = f"https://{domain}/oauth/token"
    payload = {
        "grant_type": "client_credentials",
        "client_id": client_id,
        "client_secret": client_secret,
        "audience": f"https://{domain}/api/v2/",
    }
    print(f"\n  [auth0] Requesting Management API token from: {url}")
    print(f"  [auth0]   client_id: {client_id}")
    print(f"  [auth0]   audience:  https://{domain}/api/v2/")

    try:
        resp = requests.post(url, json=payload, timeout=15)
    except requests.exceptions.ConnectionError as exc:
        pytest.fail(
            f"Cannot connect to Auth0 at {url}.\n"
            f"  Check AUTH0_DOMAIN is correct (current: '{domain}').\n"
            f"  Error: {exc}"
        )

    if resp.status_code != 200:
        body = resp.text[:500]
        pytest.fail(
            f"Failed to obtain Management API token (HTTP {resp.status_code}).\n"
            f"  URL: {url}\n"
            f"  Response: {body}\n\n"
            f"  Troubleshooting:\n"
            f"    - 401/403: AUTH0_MGMT_CLIENT_ID or secret is wrong, or the M2M app "
            f"is not authorized for audience https://{domain}/api/v2/\n"
            f"    - 404: AUTH0_DOMAIN is incorrect (current: '{domain}')\n"
        )

    token = resp.json().get("access_token")
    if not token:
        pytest.fail(
            f"Auth0 token response missing 'access_token'.\n"
            f"  Response body: {resp.text[:500]}"
        )

    print(f"  [auth0] ✓ Management API token obtained")
    return token


def _create_auth0_user(domain: str, token: str, connection: str, email: str, password: str) -> str:
    """Create an Auth0 user and return its user_id."""
    url = f"https://{domain}/api/v2/users"
    payload = {
        "connection": connection,
        "email": email,
        "password": password,
        "email_verified": True,
    }
    print(f"  [auth0] Creating user: {email} (connection: {connection})")

    resp = requests.post(
        url,
        headers={"Authorization": f"Bearer {token}"},
        json=payload,
        timeout=15,
    )

    if resp.status_code == 201:
        user_id = resp.json()["user_id"]
        print(f"  [auth0] ✓ User created: {user_id}")
        return user_id

    body = resp.text[:500]
    error_detail = ""
    if resp.status_code == 403:
        error_detail = (
            "  Troubleshooting:\n"
            "    - The M2M app lacks the 'create:users' scope on the Management API.\n"
            "    - Go to Auth0 Dashboard → Applications → APIs → Auth0 Management API → "
            "Machine to Machine Applications → authorize your M2M app with 'create:users'.\n"
        )
    elif resp.status_code == 400:
        error_detail = (
            "  Troubleshooting:\n"
            f"    - Check AUTH0_DB_CONNECTION (current: '{connection}') exists in your tenant.\n"
            "    - Go to Auth0 Dashboard → Authentication → Database to verify connection name.\n"
            "    - Password may not meet the connection's password policy.\n"
        )
    elif resp.status_code == 409:
        error_detail = (
            "  Troubleshooting:\n"
            f"    - A user with email '{email}' already exists (unlikely with random emails).\n"
        )

    pytest.fail(
        f"Failed to create Auth0 user (HTTP {resp.status_code}).\n"
        f"  URL: {url}\n"
        f"  Payload: connection={connection}, email={email}\n"
        f"  Response: {body}\n\n"
        f"{error_detail}"
    )


def _delete_auth0_user(domain: str, token: str, user_id: str) -> None:
    """Delete an Auth0 user by user_id."""
    url = f"https://{domain}/api/v2/users/{user_id}"
    print(f"  [auth0] Deleting user: {user_id}")

    resp = requests.delete(
        url,
        headers={"Authorization": f"Bearer {token}"},
        timeout=15,
    )

    if resp.status_code == 204:
        print(f"  [auth0] ✓ User deleted")
        return

    body = resp.text[:500]
    hint = ""
    if resp.status_code == 403:
        hint = " (M2M app may lack 'delete:users' scope)"
    raise RuntimeError(
        f"Failed to delete Auth0 user (HTTP {resp.status_code}){hint}.\n"
        f"  URL: {url}\n"
        f"  Response: {body}"
    )


@pytest.fixture(scope="session")
def base_url():
    """Base URL for the application"""
    return os.getenv("TEST_BASE_URL", "http://localhost:8000")


@pytest.fixture(scope="session")
def auth0_test_user():
    """Create a temporary Auth0 user for the test session and delete it on teardown."""
    domain = settings.auth0_domain
    client_id = settings.auth0_mgmt_client_id
    client_secret = settings.auth0_mgmt_client_secret
    connection = settings.auth0_db_connection

    print("\n  ─── Auth0 Integration Test Configuration ───")
    print(f"  AUTH0_DOMAIN:             {'✓ ' + domain if domain else '✗ MISSING'}")
    print(f"  AUTH0_MGMT_CLIENT_ID:     {'✓ ' + client_id if client_id else '✗ MISSING'}")
    print(f"  AUTH0_MGMT_CLIENT_SECRET: {'✓ (set, length=' + str(len(client_secret)) + ')' if client_secret else '✗ MISSING'}")
    print(f"  AUTH0_DB_CONNECTION:      {connection}")
    print(f"  SECRETS_PROVIDER:         {settings.secrets_provider}")
    print("  ─────────────────────────────────────────────")

    if not domain or not client_id or not client_secret:
        missing = []
        if not domain:
            missing.append("AUTH0_DOMAIN")
        if not client_id:
            missing.append("AUTH0_MGMT_CLIENT_ID")
        if not client_secret:
            missing.append("AUTH0_MGMT_CLIENT_SECRET (check secrets/auth0_mgmt_client_secret.txt or env var)")
        pytest.skip(
            f"Missing Auth0 config — skipping authenticated tests.\n"
            f"  Missing: {', '.join(missing)}"
        )

    token = _get_mgmt_token(domain, client_id, client_secret)

    email = f"test-{uuid.uuid4().hex[:8]}@integration-tests.local"
    password = f"T3st!{uuid.uuid4().hex}"

    user_id = _create_auth0_user(domain, token, connection, email, password)
    print(f"\n  ✓ Created Auth0 test user: {email} ({user_id})")

    yield {"email": email, "password": password}

    try:
        _delete_auth0_user(domain, token, user_id)
        print(f"\n  ✓ Deleted Auth0 test user: {email} ({user_id})")
    except Exception as exc:
        print(f"\n  ✗ Failed to delete Auth0 test user {user_id}: {exc}")

@pytest.fixture(scope="session")
def credentials(auth0_test_user):
    """Alias for auth0_test_user, providing login credentials to tests."""
    return auth0_test_user


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
    base = os.getenv("TEST_BASE_URL", "http://localhost:8000")
    chrome_options.add_argument(f"--unsafely-treat-insecure-origin-as-secure={base}")
    
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
