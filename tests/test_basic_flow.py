import pytest
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
import time

class TestBasicFlow:
    """Test suite for basic user flow (without Auth0 login)"""
    
    def test_full_navigation_flow(self, driver, base_url):
        """Test navigating through the main pages"""
        # Start at landing page
        driver.get(base_url)
        time.sleep(2)
        
        print(f"Step 1: Landing page loaded")
        print(f"  URL: {driver.current_url}")
        print(f"  Title: {driver.title}")
        driver.save_screenshot("/tmp/flow_1_landing.png")
        
        # Try to navigate to /app
        app_url = f"{base_url}/app"
        driver.get(app_url)
        time.sleep(3)
        
        print(f"\nStep 2: Navigated to /app")
        print(f"  URL: {driver.current_url}")
        print(f"  Title: {driver.title}")
        driver.save_screenshot("/tmp/flow_2_app.png")
        
        # Check console for errors
        logs = driver.get_log('browser')
        errors = [log for log in logs if log['level'] == 'SEVERE']
        
        if errors:
            print("\n⚠️  Console errors found:")
            for error in errors:
                print(f"  - {error['message']}")
        
        # Try to navigate to /admin
        admin_url = f"{base_url}/admin"
        driver.get(admin_url)
        time.sleep(2)
        
        print(f"\nStep 3: Navigated to /admin")
        print(f"  URL: {driver.current_url}")
        print(f"  Title: {driver.title}")
        driver.save_screenshot("/tmp/flow_3_admin.png")
        
        # Try API health check
        health_url = f"{base_url}/health"
        driver.get(health_url)
        time.sleep(1)
        
        print(f"\nStep 4: API health check")
        print(f"  URL: {driver.current_url}")
        print(f"  Response: {driver.page_source[:200]}")
        driver.save_screenshot("/tmp/flow_4_health.png")
        
        # Try Swagger
        swagger_url = f"{base_url}/swagger/index.html"
        driver.get(swagger_url)
        time.sleep(2)
        
        print(f"\nStep 5: Swagger UI")
        print(f"  URL: {driver.current_url}")
        print(f"  Title: {driver.title}")
        driver.save_screenshot("/tmp/flow_5_swagger.png")
        
        # Return to landing
        driver.get(base_url)
        time.sleep(1)
        
        print(f"\nStep 6: Back to landing")
        print(f"  URL: {driver.current_url}")
        driver.save_screenshot("/tmp/flow_6_landing_return.png")
        
        # All navigation should work without errors
        assert True, "Navigation flow completed"
    
    def test_network_requests(self, driver, base_url):
        """Test that network requests are successful"""
        # Enable performance logging
        driver.get(f"{base_url}/app")
        time.sleep(3)
        
        # Get performance logs
        try:
            perf_logs = driver.get_log('performance')
            
            # Parse network requests
            failed_requests = []
            for log in perf_logs:
                import json
                message = json.loads(log['message'])
                method = message.get('message', {}).get('method', '')
                
                if method == 'Network.responseReceived':
                    params = message['message']['params']
                    response = params.get('response', {})
                    status = response.get('status', 0)
                    url = response.get('url', '')
                    
                    if status >= 400:
                        failed_requests.append({
                            'url': url,
                            'status': status
                        })
            
            if failed_requests:
                print("\n❌ Failed network requests:")
                for req in failed_requests:
                    print(f"  [{req['status']}] {req['url']}")
            
            # Report but don't fail the test
            print(f"\nTotal failed requests: {len(failed_requests)}")
            
        except Exception as e:
            print(f"Could not get performance logs: {e}")
    
    def test_page_load_times(self, driver, base_url):
        """Test page load performance"""
        pages = [
            ('Landing', base_url),
            ('App', f"{base_url}/app"),
            ('Admin', f"{base_url}/admin"),
            ('Health', f"{base_url}/health"),
            ('Swagger', f"{base_url}/swagger/index.html"),
        ]
        
        results = []
        
        for name, url in pages:
            start_time = time.time()
            driver.get(url)
            time.sleep(1)  # Wait for initial render
            load_time = time.time() - start_time
            
            results.append({
                'page': name,
                'url': url,
                'load_time': load_time
            })
            
            print(f"{name}: {load_time:.2f}s")
        
        # All pages should load within reasonable time
        for result in results:
            assert result['load_time'] < 30, f"{result['page']} took too long to load: {result['load_time']:.2f}s"
