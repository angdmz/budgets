import pytest
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
import time

class TestAppRouting:
    """Test suite for app routing and asset loading"""
    
    def test_app_route_accessible(self, driver, base_url):
        """Test that /app route is accessible"""
        app_url = f"{base_url}/app"
        driver.get(app_url)
        
        # Wait for page to load
        time.sleep(2)
        
        # Check current URL
        current_url = driver.current_url
        print(f"Navigated to: {current_url}")
        print(f"Expected: {app_url}")
        
        # Take screenshot
        driver.save_screenshot("/tmp/app_route.png")
        
        # Check if we're on Auth0 login or the app
        page_source = driver.page_source.lower()
        
        # Should either be on Auth0 login or see the app
        is_auth0 = "auth0" in current_url or "login" in page_source
        is_app_loaded = "budget" in page_source or "dashboard" in page_source
        
        assert is_auth0 or is_app_loaded, f"App route not loading correctly. URL: {current_url}"
    
    def test_app_assets_load(self, driver, base_url):
        """Test that app assets (JS, CSS) load without 404 errors"""
        app_url = f"{base_url}/app"
        driver.get(app_url)
        
        time.sleep(3)  # Wait for assets to load
        
        # Get browser console logs
        logs = driver.get_log('browser')
        
        # Check for 404 errors
        errors_404 = [log for log in logs if '404' in log.get('message', '')]
        
        if errors_404:
            print("404 Errors found:")
            for error in errors_404:
                print(f"  - {error['message']}")
            
            # Take screenshot
            driver.save_screenshot("/tmp/app_404_errors.png")
        
        # Get network performance entries to verify assets loaded successfully
        performance_entries = driver.execute_script("""
            return performance.getEntriesByType('resource').map(entry => ({
                name: entry.name,
                type: entry.initiatorType,
                status: entry.responseStatus || 'unknown',
                duration: entry.duration
            }));
        """)
        
        # Filter for JS and CSS assets
        js_assets = [e for e in performance_entries if e['name'].endswith('.js')]
        css_assets = [e for e in performance_entries if e['name'].endswith('.css')]
        
        print(f"\nLoaded JS assets: {len(js_assets)}")
        for asset in js_assets:
            print(f"  - {asset['name'].split('/')[-1]}: {asset['duration']:.2f}ms")
        
        print(f"\nLoaded CSS assets: {len(css_assets)}")
        for asset in css_assets:
            print(f"  - {asset['name'].split('/')[-1]}: {asset['duration']:.2f}ms")
        
        # Print all console logs for debugging
        print("\nAll browser console logs:")
        for log in logs:
            print(f"  [{log['level']}] {log['message']}")
        
        # Assert no 404 errors
        assert len(errors_404) == 0, f"Found {len(errors_404)} 404 errors in app"
        
        # Assert that at least one JS and one CSS file loaded
        assert len(js_assets) > 0, "No JavaScript assets loaded"
        assert len(css_assets) > 0, "No CSS assets loaded"
    
    def test_app_javascript_loads(self, driver, base_url):
        """Test that JavaScript executes correctly"""
        app_url = f"{base_url}/app"
        driver.get(app_url)
        
        time.sleep(2)
        
        # Try to execute JavaScript
        try:
            result = driver.execute_script("return document.readyState")
            assert result == "complete", f"Document not fully loaded: {result}"
            
            # Check if React is loaded
            react_loaded = driver.execute_script(
                "return typeof React !== 'undefined' || document.getElementById('root') !== null"
            )
            print(f"React/Root element present: {react_loaded}")
            
        except Exception as e:
            print(f"JavaScript execution error: {e}")
            driver.save_screenshot("/tmp/app_js_error.png")
            raise
    
    def test_admin_route_accessible(self, driver, base_url):
        """Test that /admin route is accessible"""
        admin_url = f"{base_url}/admin"
        driver.get(admin_url)
        
        time.sleep(2)
        
        current_url = driver.current_url
        print(f"Admin URL: {current_url}")
        
        driver.save_screenshot("/tmp/admin_route.png")
        
        # Should redirect to Auth0 or show admin page
        page_source = driver.page_source.lower()
        is_auth0 = "auth0" in current_url or "login" in page_source
        is_admin_loaded = "admin" in page_source or "budget" in page_source
        
        assert is_auth0 or is_admin_loaded, f"Admin route not loading correctly. URL: {current_url}"
    
    def test_admin_assets_load(self, driver, base_url):
        """Test that admin assets (JS, CSS) load without 404 errors"""
        admin_url = f"{base_url}/admin"
        driver.get(admin_url)
        
        time.sleep(3)  # Wait for assets to load
        
        # Get browser console logs
        logs = driver.get_log('browser')
        
        # Check for 404 errors
        errors_404 = [log for log in logs if '404' in log.get('message', '')]
        
        if errors_404:
            print("404 Errors found in admin:")
            for error in errors_404:
                print(f"  - {error['message']}")
            driver.save_screenshot("/tmp/admin_404_errors.png")
        
        # Get network performance entries
        performance_entries = driver.execute_script("""
            return performance.getEntriesByType('resource').map(entry => ({
                name: entry.name,
                type: entry.initiatorType,
                duration: entry.duration
            }));
        """)
        
        # Filter for JS and CSS assets
        js_assets = [e for e in performance_entries if e['name'].endswith('.js')]
        css_assets = [e for e in performance_entries if e['name'].endswith('.css')]
        
        print(f"\nAdmin loaded JS assets: {len(js_assets)}")
        for asset in js_assets:
            print(f"  - {asset['name'].split('/')[-1]}: {asset['duration']:.2f}ms")
        
        print(f"\nAdmin loaded CSS assets: {len(css_assets)}")
        for asset in css_assets:
            print(f"  - {asset['name'].split('/')[-1]}: {asset['duration']:.2f}ms")
        
        # Assert no 404 errors
        assert len(errors_404) == 0, f"Found {len(errors_404)} 404 errors in admin"
        
        # Assert that at least one JS and one CSS file loaded
        assert len(js_assets) > 0, "No JavaScript assets loaded in admin"
        assert len(css_assets) > 0, "No CSS assets loaded in admin"
    
    def test_api_health_endpoint(self, driver, base_url):
        """Test that API health endpoint is accessible"""
        health_url = f"{base_url}/health"
        driver.get(health_url)
        
        time.sleep(1)
        
        page_source = driver.page_source
        print(f"Health endpoint response: {page_source}")
        
        # Should return JSON with status
        assert "ok" in page_source.lower() or "status" in page_source.lower()
    
    def test_swagger_accessible(self, driver, base_url):
        """Test that Swagger UI is accessible"""
        swagger_url = f"{base_url}/swagger/index.html"
        driver.get(swagger_url)
        
        time.sleep(2)
        
        page_source = driver.page_source.lower()
        print(f"Swagger page loaded, length: {len(page_source)}")
        
        driver.save_screenshot("/tmp/swagger_page.png")
        
        # Check for Swagger UI elements
        assert "swagger" in page_source or "api" in page_source
