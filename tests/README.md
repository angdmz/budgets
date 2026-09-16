# Integration Tests - Budget Management System

Selenium-based integration tests to verify the basic flow and functionality of the Budget Management System.

## Environment Variables

The integration tests read configuration from the root `.env` file (via `env_file: .env` in `docker-compose.yml`) and from Docker secrets. Settings are loaded in `tests/settings.py` via a Pydantic `BaseSettings` class with the `INTEGRATION_TESTS_` prefix.

### Environment Variables

| Variable | Default | Used In | Purpose |
|----------|---------|---------|---------|
| `INTEGRATION_TESTS_BASE_URL` | `http://localhost:8000` | `settings.py:12` → `Settings.base_url` | Base URL for the application under test. In Docker Compose, set to `http://nginx` (the internal Nginx service). For local runs, use `http://localhost:8000`. |
| `INTEGRATION_TESTS_SCREENSHOTS_DIR` | `/tests/screenshots` | `settings.py:13` → `Settings.screenshots_dir` | Directory inside the test container where screenshots are saved. |
| `INTEGRATION_TESTS_SCREENSHOTS_HOST_DIR` | `./tests/screenshots` | `docker-compose.yml:190` → `volumes` | Host directory mapped to the container's screenshots directory, so screenshots are accessible on the host after tests run. |
| `INTEGRATION_TESTS_SECRETS_PROVIDER` | `env` | `settings.py:14` → `Settings.secrets_provider` | Secrets provider for test credentials. Set to `docker` in Docker Compose to read from `/run/secrets/`. |
| `INTEGRATION_TEST` | `true` | `docker-compose.yml` | Flag indicating integration test mode. |
| `AUTH0_DOMAIN` | _(empty)_ | `settings.py:20` → `Settings.auth0_domain` | Auth0 tenant domain. Used to obtain Management API tokens for dynamic test user creation/deletion. |
| `AUTH0_MGMT_CLIENT_ID` | _(empty)_ | `settings.py:23` → `Settings.auth0_mgmt_client_id` | Auth0 Machine-to-Machine (M2M) client ID. Used to authenticate with the Auth0 Management API for creating and deleting test users. |
| `AUTH0_CLIENT_ID` | _(empty)_ | `settings.py:37` → `Settings.auth0_client_id` | Auth0 SPA client ID. Used for test user login flows. |
| `AUTH0_AUDIENCE` | `https://api.budget.local` | `settings.py:49` → `Settings.auth0_audience` | Auth0 API identifier. Used when requesting access tokens for test users. |
| `AUTH0_DB_CONNECTION` | `Username-Password-Authentication` | `settings.py:53` → `Settings.auth0_db_connection` | Auth0 database connection name. Used when creating test users via the Management API. |

### Secrets

The integration tests use three secrets, mounted via Docker secrets in `docker-compose.yml`:

| Secret Key | Retrieved In | Used For |
|------------|--------------|----------|
| `auth0_mgmt_client_secret` | `settings.py:28-33` → reads `/run/secrets/auth0_mgmt_client_secret` (docker) or `AUTH0_MGMT_CLIENT_SECRET` env var | Auth0 M2M client secret. Used with `AUTH0_MGMT_CLIENT_ID` to obtain a Management API token for dynamic test user creation and deletion. |
| `integration_tests_auth0_email` | `settings.py:57-60` → reads `/run/secrets/integration_tests_auth0_email` (docker) or `INTEGRATION_TESTS_AUTH0_EMAIL` env var | Email address for the test user account used in Auth0 login flow tests. |
| `integration_tests_auth0_password` | `settings.py:62-66` → reads `/run/secrets/integration_tests_auth0_password` (docker) or `INTEGRATION_TESTS_AUTH0_PASSWORD` env var | Password for the test user account used in Auth0 login flow tests. |

When `INTEGRATION_TESTS_SECRETS_PROVIDER=docker`, secrets are read from `/run/secrets/`. When set to `env`, they are read from the corresponding environment variables.

**Note:** Auth0 M2M credentials are optional. If not provided, authenticated tests (login flow, CRUD operations) are skipped — see `conftest.py:150-161` for the skip logic.

## Prerequisites

- Docker and Docker Compose
- System running on http://localhost:8000

## Quick Start (Recommended)

### Run Tests with Docker Compose

```bash
# From the project root
docker-compose --profile integration run --rm integration-tests
```

Or use the convenience script:

```bash
cd tests
./run_tests.sh
```

This will:
1. Check if services are running (start them if not)
2. Build the test container
3. Run all integration tests
4. Generate HTML report
5. Save screenshots

### View Results

```bash
# Open HTML report
open tests/report.html

# View screenshots
ls tests/screenshots/
```

## Running Tests Locally (Alternative)

### 1. Install Dependencies

```bash
cd tests
pip install -r requirements.txt
```

### 2. Configure Environment

```bash
cp .env.example .env
# Edit .env with your configuration
```

### 3. Ensure System is Running

```bash
cd ..
docker-compose up -d
```

### 4. Run Tests

```bash
pytest -v
```

### Run Specific Test Suite

```bash
# Landing page tests
pytest test_landing_page.py -v

# App routing tests
pytest test_app_routing.py -v

# Basic flow tests
pytest test_basic_flow.py -v
```

### Run with HTML Report

```bash
pytest --html=report.html --self-contained-html
```

### Run in Visible Browser (for debugging)

```bash
pytest -v -k "test_name" --capture=no
```

### Run with Verbose Output

```bash
pytest -v -s
```

## Test Suites

### 1. Landing Page Tests (`test_landing_page.py`)

Tests the landing page functionality:
- Page loads successfully
- CTA buttons are present
- Features section is visible
- Navigation links exist

### 2. App Routing Tests (`test_app_routing.py`)

Tests routing and asset loading:
- `/app` route is accessible
- **App assets (JS, CSS) load without 404 errors** 
  - Verifies no 404 errors in browser console
  - Uses Performance API to verify JS and CSS assets loaded
  - Asserts at least one JS and one CSS file loaded successfully
  - Reports asset load times for performance monitoring
- JavaScript executes correctly
- `/admin` route is accessible
- **Admin assets (JS, CSS) load without 404 errors** 
  - Same comprehensive checks as app assets
  - Ensures both React applications load correctly
- API health endpoint works
- Swagger UI is accessible

### 3. Basic Flow Tests (`test_basic_flow.py`)

Tests the complete navigation flow:
- Full navigation through all pages
- Network request monitoring
- Page load time measurements
- Console error detection

## Test Output

### Screenshots

Tests automatically save screenshots to `/tmp/`:
- `landing_page.png` - Landing page
- `app_route.png` - Main app page
- `admin_route.png` - Admin page
- `swagger_page.png` - Swagger UI
- `flow_*.png` - Navigation flow screenshots

### Console Logs

Tests capture and display:
- Browser console errors
- Network 404 errors
- JavaScript execution errors
- Performance metrics

## Debugging

### View Screenshots

```bash
open /tmp/landing_page.png
open /tmp/app_route.png
```

### Run Single Test with Visible Browser

Edit `conftest.py` to use `driver_visible` fixture or run:

```bash
pytest test_landing_page.py::TestLandingPage::test_landing_page_loads -v -s
```

### Check Console Logs

Tests print console logs to stdout. Run with `-s` flag:

```bash
pytest test_app_routing.py -v -s
```

## Common Issues

### ChromeDriver Not Found

The tests use `webdriver-manager` to automatically download ChromeDriver. If it fails:

```bash
pip install --upgrade webdriver-manager
```

### Connection Refused

Ensure the system is running:

```bash
docker-compose ps
curl http://localhost:8000/health
```

### 404 Errors on Assets

This indicates a routing issue. Check:
1. Nginx configuration
2. Vite build configuration (base path)
3. Asset paths in built files

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Start services
        run: docker-compose up -d
      
      - name: Wait for services
        run: sleep 30
      
      - name: Run tests
        run: |
          cd tests
          pip install -r requirements.txt
          pytest -v --html=report.html
      
      - name: Upload test report
        uses: actions/upload-artifact@v2
        with:
          name: test-report
          path: tests/report.html
```

## Test Coverage

Current test coverage:
- ✅ Landing page rendering
- ✅ App routing and navigation
- ✅ Asset loading verification
- ✅ API health checks
- ✅ Swagger UI accessibility
- ✅ Theme preference persistence (requires Auth0 M2M credentials for dynamic user creation)
- ⚠️ Auth0 login flow (requires Auth0 M2M credentials for dynamic user creation)
- ⚠️ CRUD operations (requires Auth0 M2M credentials for dynamic user creation)

## Adding New Tests

### Example Test

```python
def test_new_feature(driver, base_url):
    """Test description"""
    driver.get(f"{base_url}/new-page")
    
    # Wait for element
    element = WebDriverWait(driver, 10).until(
        EC.presence_of_element_located((By.ID, "element-id"))
    )
    
    # Assert
    assert element.text == "Expected Text"
    
    # Screenshot
    driver.save_screenshot("/tmp/new_feature.png")
```

## Troubleshooting

### Tests Fail with Timeout

Increase implicit wait in `conftest.py`:

```python
driver.implicitly_wait(20)  # Increase from 10 to 20 seconds
```

### Assets Still 404

Check the test output for specific URLs failing, then:

1. Verify Nginx configuration
2. Check Vite base path in `vite.config.ts`
3. Rebuild frontend without cache:
   ```bash
   docker-compose build --no-cache app admin
   ```

### Browser Crashes

Add more memory to Docker:
```bash
docker-compose down
# Increase Docker memory in Docker Desktop settings
docker-compose up -d
```

## Support

For issues or questions:
1. Check test output and screenshots
2. Review browser console logs
3. Verify system is running: `docker-compose ps`
4. Check Nginx logs: `docker-compose logs nginx`
