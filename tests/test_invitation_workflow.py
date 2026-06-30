import os
import pytest
import time
from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait, Select
from selenium.webdriver.support import expected_conditions as EC
from selenium.common.exceptions import TimeoutException


class TestInvitationWorkflow:
    """End-to-end tests for group invitation system.

    Tests the flow of:
    - Owner creating an invitation
    - Second user accepting the invitation
    - Access granted to group and budgets
    - Revocation scenarios

    Each test uses a fresh WebDriver instance for the second user so that
    Auth0 SSO cookies from user 1 never bleed into user 2's session.
    """

    TIMEOUT = 20

    @staticmethod
    def _make_driver():
        """Create a fresh Chrome WebDriver with its own cookie jar (no shared SSO)."""
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
        chrome_options.binary_location = "/usr/bin/chromium"
        drv = webdriver.Chrome(options=chrome_options)
        drv.set_page_load_timeout(30)
        drv.implicitly_wait(10)
        return drv

    def _wait(self, driver):
        return WebDriverWait(driver, self.TIMEOUT)

    def _login(self, driver, base_url, credentials):
        """Navigate to /app and complete Auth0 Universal Login if redirected."""
        driver.get(f"{base_url}/app")
        # Wait up to 15 s for the React app to boot and either redirect to auth0.com
        # or render its root element.  The second Chrome instance has a cold start and
        # may need significantly longer than the fixed 2-second sleep used before.
        try:
            WebDriverWait(driver, 15).until(
                lambda d: "auth0.com" in d.current_url
                    or bool(d.execute_script(
                        "return document.getElementById('root') && "
                        "document.getElementById('root').children.length > 0"
                    ))
            )
        except TimeoutException:
            pass  # handled below by checking current_url / page_source

        if "auth0.com" in driver.current_url or "auth0" in driver.page_source.lower():
            try:
                email_field = WebDriverWait(driver, 10).until(
                    EC.presence_of_element_located(
                        (By.CSS_SELECTOR,
                         "input[name='username'], input[type='email'], input[id='username']")
                    )
                )
            except TimeoutException:
                pytest.skip(
                    "Auth0 login form did not appear — the redirect_uri may not be "
                    "registered in the Auth0 tenant."
                )
            email_field.clear()
            email_field.send_keys(credentials["email"])

            driver.find_element(By.CSS_SELECTOR, "button[type='submit']").click()

            pwd_field = self._wait(driver).until(
                EC.visibility_of_element_located(
                    (By.CSS_SELECTOR, "input[type='password']")
                )
            )

            pwd_field.clear()
            pwd_field.send_keys(credentials["password"])
            driver.find_element(By.CSS_SELECTOR, "button[type='submit']").click()

            try:
                self._wait(driver).until(lambda d: "auth0.com" not in d.current_url)
            except TimeoutException:
                pytest.fail("Browser did not leave auth0.com after submitting credentials.")
            time.sleep(3)

        try:
            WebDriverWait(driver, 30).until(
                EC.presence_of_element_located(
                    (By.XPATH, "//nav//a[contains(normalize-space(),'Groups')]")
                )
            )
        except TimeoutException:
            pytest.fail("Authenticated app layout did not appear after login.")

    def _go_to_groups(self, driver, base_url):
        """Navigate to Groups via the nav link."""
        groups_link = WebDriverWait(driver, 15).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//nav//a[contains(normalize-space(),'Groups')]")
            )
        )
        groups_link.click()
        time.sleep(2)

    def _go_to_budgets(self, driver):
        """Navigate to Budgets via the nav link."""
        budgets_link = WebDriverWait(driver, 15).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//nav//a[contains(normalize-space(),'Budgets')]")
            )
        )
        budgets_link.click()
        time.sleep(2)

    def test_owner_invites_and_user_accepts(self, driver, base_url, credentials, second_auth0_test_user):
        """Test complete invitation flow: create group → invite → accept → access granted."""
        # ── User 1: create group and generate invite ──────────────────────────────
        self._login(driver, base_url, credentials)
        self._go_to_groups(driver, base_url)

        self._wait(driver).until(
            EC.element_to_be_clickable((By.XPATH, "//button[contains(text(),'Add Group')]"))
        ).click()
        time.sleep(1)

        group_name_input = self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, "input[type='text'][required]"))
        )
        group_name_input.clear()
        group_name_input.send_keys("Invitation Test Group")
        driver.find_element(By.CSS_SELECTOR, "button[type='submit']").click()
        time.sleep(2)

        invite_button = self._wait(driver).until(
            EC.element_to_be_clickable((By.XPATH, "//button[contains(text(),'Invite')]"))
        )
        invite_button.click()
        time.sleep(1)

        generate_button = self._wait(driver).until(
            EC.element_to_be_clickable((By.XPATH, "//button[contains(text(),'Generate Invite Link')]"))
        )
        generate_button.click()
        time.sleep(1)

        invitation_link = self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, "input[readonly], textarea[readonly]"))
        )
        invite_url = invitation_link.get_attribute("value")

        # ── User 2: accept invite in a fresh browser (own cookie jar, no SSO bleed) ──
        second_driver = self._make_driver()
        try:
            self._login(second_driver, base_url, second_auth0_test_user)
            second_driver.get(invite_url)
            time.sleep(2)

            accept_button = self._wait(second_driver).until(
                EC.element_to_be_clickable((By.XPATH, "//button[contains(text(),'Join Group') or contains(text(),'Accept')]"))
            )
            accept_button.click()

            # AcceptInvitation navigates to /groups after a 2-second timeout on success
            try:
                WebDriverWait(second_driver, 15).until(
                    lambda d: "/invite/" not in d.current_url
                )
            except TimeoutException:
                pytest.fail(
                    f"Accept invitation did not navigate away from invite page. "
                    f"URL: {second_driver.current_url}"
                )
            time.sleep(1)

            self._go_to_groups(second_driver, base_url)

            group_visible = self._wait(second_driver).until(
                EC.presence_of_element_located((By.XPATH, "//*[contains(text(),'Invitation Test Group')]"))
            )
            assert group_visible is not None
            print("\n  ✓ User accepted invitation and can see group")
        finally:
            second_driver.quit()

    def test_revoke_before_accept(self, driver, base_url, credentials, second_auth0_test_user):
        """Test that revoked invitations cannot be accepted."""
        # ── User 1: create group, generate invite, revoke ─────────────────────────
        self._login(driver, base_url, credentials)
        self._go_to_groups(driver, base_url)

        self._wait(driver).until(
            EC.element_to_be_clickable((By.XPATH, "//button[contains(text(),'Add Group')]"))
        ).click()
        time.sleep(1)

        group_name_input = self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, "input[type='text'][required]"))
        )
        group_name_input.send_keys("Revocation Test Group")
        driver.find_element(By.CSS_SELECTOR, "button[type='submit']").click()
        time.sleep(2)

        invite_button = self._wait(driver).until(
            EC.element_to_be_clickable((By.XPATH, "//button[contains(text(),'Invite')]"))
        )
        invite_button.click()
        time.sleep(1)

        generate_button = self._wait(driver).until(
            EC.element_to_be_clickable((By.XPATH, "//button[contains(text(),'Generate Invite Link')]"))
        )
        generate_button.click()
        time.sleep(1)

        invitation_link = self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, "input[readonly]"))
        )
        invite_url = invitation_link.get_attribute("value")

        revoke_button = self._wait(driver).until(
            EC.element_to_be_clickable((By.XPATH, "//button[contains(text(),'Revoke')]"))
        )
        revoke_button.click()
        time.sleep(1)

        # ── User 2: attempt to accept revoked invite in a fresh browser ───────────
        second_driver = self._make_driver()
        try:
            self._login(second_driver, base_url, second_auth0_test_user)
            second_driver.get(invite_url)
            time.sleep(2)

            error_msg = self._wait(second_driver).until(
                EC.presence_of_element_located((By.XPATH, "//*[contains(text(),'expired') or contains(text(),'revoked') or contains(text(),'no longer valid')]"))
            )
            assert error_msg is not None
            print("\n  ✓ Revoked invitation shows error message")
        finally:
            second_driver.quit()

    def test_accepted_user_sees_budgets(self, driver, base_url, credentials, second_auth0_test_user):
        """Test that users who accept invitations can see group budgets."""
        # ── User 1: create group, generate invite, add a budget ───────────────────
        self._login(driver, base_url, credentials)
        self._go_to_groups(driver, base_url)

        self._wait(driver).until(
            EC.element_to_be_clickable((By.XPATH, "//button[contains(text(),'Add Group')]"))
        ).click()
        time.sleep(1)

        group_name_input = self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, "input[type='text'][required]"))
        )
        group_name_input.send_keys("Budget Access Test")
        driver.find_element(By.CSS_SELECTOR, "button[type='submit']").click()
        time.sleep(2)

        invite_button = self._wait(driver).until(
            EC.element_to_be_clickable((By.XPATH, "//button[contains(text(),'Invite')]"))
        )
        invite_button.click()
        time.sleep(1)

        generate_button = self._wait(driver).until(
            EC.element_to_be_clickable((By.XPATH, "//button[contains(text(),'Generate Invite Link')]"))
        )
        generate_button.click()
        time.sleep(1)

        invitation_link = self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, "input[readonly]"))
        )
        invite_url = invitation_link.get_attribute("value")

        self._wait(driver).until(
            EC.element_to_be_clickable((By.XPATH, "//button[normalize-space(text())='Close']"))
        ).click()
        time.sleep(1)

        self._go_to_budgets(driver)

        group_select = self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, "select"))
        )
        Select(group_select).select_by_visible_text("Budget Access Test")
        time.sleep(1)

        add_budget_btn = self._wait(driver).until(
            EC.element_to_be_clickable((By.XPATH, "//button[contains(text(),'Add Budget')]"))
        )
        add_budget_btn.click()
        time.sleep(1)

        budget_name_input = self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, "input[type='text']"))
        )
        budget_name_input.send_keys("Test Budget")

        date_inputs = driver.find_elements(By.CSS_SELECTOR, "input[type='date']")
        for inp, val in zip(date_inputs[:2], ["2025-01-01", "2025-12-31"]):
            driver.execute_script("""
                var setter = Object.getOwnPropertyDescriptor(
                    window.HTMLInputElement.prototype, 'value').set;
                setter.call(arguments[0], arguments[1]);
                arguments[0].dispatchEvent(new Event('input', { bubbles: true }));
                arguments[0].dispatchEvent(new Event('change', { bubbles: true }));
            """, inp, val)

        driver.find_element(By.CSS_SELECTOR, "button[type='submit']").click()
        self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, "//*[contains(text(),'Test Budget')]"))
        )

        # ── User 2: accept invite and verify budget access in a fresh browser ──────
        second_driver = self._make_driver()
        try:
            self._login(second_driver, base_url, second_auth0_test_user)
            second_driver.get(invite_url)
            time.sleep(2)

            accept_button = self._wait(second_driver).until(
                EC.element_to_be_clickable((By.XPATH, "//button[contains(text(),'Join Group') or contains(text(),'Accept')]"))
            )
            accept_button.click()

            # AcceptInvitation navigates to /groups after a 2-second timeout on success
            try:
                WebDriverWait(second_driver, 15).until(
                    lambda d: "/invite/" not in d.current_url
                )
            except TimeoutException:
                pytest.fail(
                    f"Accept invitation did not navigate away from invite page. "
                    f"URL: {second_driver.current_url}"
                )
            time.sleep(1)

            self._go_to_budgets(second_driver)

            # Wait for the group option to appear in the select (groups query may still be loading)
            self._wait(second_driver).until(
                EC.presence_of_element_located(
                    (By.XPATH, "//select//option[contains(text(),'Budget Access Test')]")
                )
            )
            group_select = second_driver.find_element(By.CSS_SELECTOR, "select")
            opt = second_driver.find_element(
                By.XPATH, "//select//option[contains(text(),'Budget Access Test')]"
            )
            grp_id = opt.get_attribute("value")
            # Use native setter + change event so React's onChange fires reliably
            second_driver.execute_script("""
                var setter = Object.getOwnPropertyDescriptor(
                    window.HTMLSelectElement.prototype, 'value').set;
                setter.call(arguments[0], arguments[1]);
                arguments[0].dispatchEvent(new Event('change', { bubbles: true }));
            """, group_select, grp_id)
            time.sleep(2)

            budget_visible = self._wait(second_driver).until(
                EC.presence_of_element_located((By.XPATH, "//*[contains(text(),'Test Budget')]"))
            )
            assert budget_visible is not None
            print("\n  ✓ Invited user can see budgets after accepting invitation")
        finally:
            second_driver.quit()
