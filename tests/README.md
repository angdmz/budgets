# Integration Tests - Budget Management System

Selenium-based integration tests to verify the basic flow and functionality of the Budget Management System.

## Prerequisites

- Docker and Docker Compose
- System running on http://localhost:8000
- **For authenticated tests:** Auth0 tenant with Machine-to-Machine (M2M) application configured

## Auth0 Configuration (Required for Invitation & Budget Tests)

The invitation and budget workflow tests dynamically create and delete Auth0 users via the Management API. This requires:

### 1. Create a Machine-to-Machine Application

1. Go to Auth0 Dashboard → Applications → Applications
2. Click "Create Application"
3. Name: "Integration Tests M2M"
4. Type: **Machine to Machine Applications**
5. API: **Auth0 Management API**

### 2. Authorize Required Scopes

After creating the M2M app:
1. Go to APIs → Auth0 Management API → Machine to Machine Applications
2. Find your M2M app and expand it
3. Enable these scopes:
   - `create:users` - Create test users
   - `delete:users` - Clean up test users after tests

### 3. Configure Environment Variables

Create or update `secrets/auth0_mgmt_client_secret.txt`:
```bash
echo "your-m2m-client-secret" > secrets/auth0_mgmt_client_secret.txt
```

Set environment variables in `.env` or docker-compose:
```bash
AUTH0_DOMAIN=your-tenant.auth0.com
AUTH0_AUDIENCE=https://api.your-domain.com  # Your API identifier
AUTH0_CLIENT_ID=your-frontend-client-id     # Regular application client
AUTH0_MGMT_CLIENT_ID=your-m2m-client-id     # M2M app client ID
AUTH0_DB_CONNECTION=Username-Password-Authentication  # Database connection name
```

### 4. Verify Database Connection

Ensure you have a Database connection configured:
1. Go to Auth0 Dashboard → Authentication → Database
2. Verify connection name matches `AUTH0_DB_CONNECTION` (default: `Username-Password-Authentication`)
3. Check password policy allows generated passwords (tests use format: `T3st!<random>`)

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

# Budget workflow tests (requires Auth0)
pytest test_budget_workflow.py -v

# Invitation workflow tests (requires Auth0)
pytest test_invitation_workflow.py -v
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

### 4. Budget Workflow Tests (`test_budget_workflow.py`)

**Requires Auth0 configuration**

Tests budget creation and management:
- Group creation
- Budget creation within groups
- Category management
- Expense tracking
- Multi-user budget access

### 5. Invitation Workflow Tests (`test_invitation_workflow.py`)

**Requires Auth0 configuration**

Tests the complete group invitation system:
- **Owner creates invitation** - Group owner generates invite link
- **Second user accepts invitation** - Different user accepts and joins group
- **Access control verification** - Invited user can access group budgets
- **Revocation flow** - Revoked invitations are rejected properly

**Important:** These tests dynamically create TWO Auth0 test users per session to properly test the invitation flow.

## Test Output

### Screenshots

Tests automatically save screenshots to a configurable directory (default: `/tests/screenshots` inside the container, mapped to `./tests/screenshots` on the host via the Docker volume defined in `docker-compose.yml`):
- `landing_page.png` - Landing page
- `app_route.png` - Main app page
- `admin_route.png` - Admin page
- `swagger_page.png` - Swagger UI
- `flow_*.png` - Navigation flow screenshots

The directory is controlled by `INTEGRATION_TESTS_SCREENSHOTS_DIR` (set in `.env`).

### Console Logs

Tests capture and display:
- Browser console errors
- Network 404 errors
- JavaScript execution errors
- Performance metrics

## Debugging

### View Screenshots

```bash
open tests/screenshots/landing_page.png
open tests/screenshots/app_route.png
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
- ✅ Auth0 login flow (with M2M credentials)
- ✅ Budget CRUD operations (with M2M credentials)
- ✅ **Group invitation flow** (with M2M credentials)
  - Invitation creation
  - Multi-user acceptance
  - Access control verification
  - Revocation handling

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
    driver.save_screenshot(f"{screenshots_dir}/new_feature.png")
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
