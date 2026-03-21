import pytest
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC

class TestLandingPage:
    """Test suite for the landing page"""
    
    def test_landing_page_loads(self, driver, base_url):
        """Test that the landing page loads successfully"""
        driver.get(base_url)
        
        # Wait for page to load
        WebDriverWait(driver, 10).until(
            EC.presence_of_element_located((By.TAG_NAME, "body"))
        )
        
        # Check page title
        assert "Budget Manager" in driver.title or "Budget" in driver.page_source
        
        # Take screenshot for debugging
        driver.save_screenshot("/tmp/landing_page.png")
        print(f"Landing page URL: {driver.current_url}")
        print(f"Page source length: {len(driver.page_source)}")
    
    def test_landing_page_has_cta_buttons(self, driver, base_url):
        """Test that CTA buttons are present"""
        driver.get(base_url)
        
        # Look for Get Started or Sign In buttons
        page_source = driver.page_source.lower()
        assert "get started" in page_source or "sign in" in page_source
    
    def test_landing_page_has_features_section(self, driver, base_url):
        """Test that features section is present"""
        driver.get(base_url)
        
        page_source = driver.page_source.lower()
        # Check for key features
        assert any(keyword in page_source for keyword in [
            "multi-user", "group-based", "real-time", "secure", "encrypted"
        ])
    
    def test_navigation_links_exist(self, driver, base_url):
        """Test that navigation links are present"""
        driver.get(base_url)
        
        # Check for navigation elements
        try:
            # Try to find links
            links = driver.find_elements(By.TAG_NAME, "a")
            assert len(links) > 0, "No links found on landing page"
            
            # Print links for debugging
            for link in links[:5]:  # First 5 links
                print(f"Link found: {link.get_attribute('href')} - {link.text}")
        except Exception as e:
            print(f"Error finding links: {e}")
            driver.save_screenshot("/tmp/landing_page_error.png")
            raise
