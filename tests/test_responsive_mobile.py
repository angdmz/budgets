import pytest
import time
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from conftest import dismiss_onboarding, assert_no_horizontal_overflow


class TestResponsiveMobile:
    """E2E responsive layout tests using mobile viewport fixtures.

    Verifies that the mobile layout shell (bottom nav, More drawer, mobile header),
    key pages, and modal dialogs render correctly without horizontal overflow
    at 390x844 and 320x568 viewports.
    """

    TIMEOUT = 20

    def _wait(self, driver):
        return WebDriverWait(driver, self.TIMEOUT)

    def _login(self, driver, base_url, credentials, screenshots_dir):
        driver.get(f"{base_url}/app")
        time.sleep(2)

        if "auth0.com" in driver.current_url or "auth0" in driver.page_source.lower():
            try:
                email_field = WebDriverWait(driver, 10).until(
                    EC.presence_of_element_located(
                        (By.CSS_SELECTOR, "input[name='username'], input[type='email'], input[id='username']")
                    )
                )
            except Exception:
                pytest.skip("Auth0 login form did not appear")

            email_field.clear()
            email_field.send_keys(credentials["email"])
            driver.find_element(By.CSS_SELECTOR, "button[type='submit']").click()

            pwd_field = self._wait(driver).until(
                EC.visibility_of_element_located((By.CSS_SELECTOR, "input[type='password']"))
            )
            pwd_field.clear()
            pwd_field.send_keys(credentials["password"])
            driver.find_element(By.CSS_SELECTOR, "button[type='submit']").click()

            self._wait(driver).until(lambda d: "auth0.com" not in d.current_url)
            time.sleep(3)

        WebDriverWait(driver, 30).until(
            EC.presence_of_element_located(
                (By.XPATH, "//nav//a[contains(normalize-space(),'Dashboard')]")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/mobile_logged_in.png")
        dismiss_onboarding(driver, base_url)

    # ── Layout shell tests (no login required) ──

    def test_mobile_no_horizontal_overflow_landing(self, driver_mobile, base_url, screenshots_dir):
        """Landing page must not overflow horizontally at 390px."""
        driver_mobile.get(base_url)
        time.sleep(2)
        driver_mobile.save_screenshot(f"{screenshots_dir}/mobile_landing.png")
        assert_no_horizontal_overflow(driver_mobile, "landing")

    def test_mobile_no_horizontal_overflow_app(self, driver_mobile, base_url, screenshots_dir):
        """/app route must not overflow horizontally at 390px."""
        driver_mobile.get(f"{base_url}/app")
        time.sleep(3)
        driver_mobile.save_screenshot(f"{screenshots_dir}/mobile_app.png")
        assert_no_horizontal_overflow(driver_mobile, "app")

    def test_narrow_no_horizontal_overflow_landing(self, driver_mobile_narrow, base_url, screenshots_dir):
        """Landing page must not overflow at 320px (smallest common viewport)."""
        driver_mobile_narrow.get(base_url)
        time.sleep(2)
        driver_mobile_narrow.save_screenshot(f"{screenshots_dir}/narrow_landing.png")
        assert_no_horizontal_overflow(driver_mobile_narrow, "landing-320")

    def test_narrow_no_horizontal_overflow_app(self, driver_mobile_narrow, base_url, screenshots_dir):
        """/app route must not overflow at 320px."""
        driver_mobile_narrow.get(f"{base_url}/app")
        time.sleep(3)
        driver_mobile_narrow.save_screenshot(f"{screenshots_dir}/narrow_app.png")
        assert_no_horizontal_overflow(driver_mobile_narrow, "app-320")

    # ── Mobile layout shell tests (require login) ──

    def test_mobile_bottom_nav_visible(self, driver_mobile, base_url, credentials, screenshots_dir):
        """Mobile bottom nav bar should be visible after login."""
        self._login(driver_mobile, base_url, credentials, screenshots_dir)

        bottom_nav = self._wait(driver_mobile).until(
            EC.presence_of_element_located(
                (By.CSS_SELECTOR, "nav.fixed.bottom-0")
            )
        )
        assert bottom_nav.is_displayed(), "Mobile bottom nav should be visible"
        driver_mobile.save_screenshot(f"{screenshots_dir}/mobile_bottom_nav.png")

    def test_mobile_desktop_nav_hidden(self, driver_mobile, base_url, credentials, screenshots_dir):
        """Desktop top nav should be hidden on mobile viewport."""
        self._login(driver_mobile, base_url, credentials, screenshots_dir)

        desktop_navs = driver_mobile.find_elements(
            By.CSS_SELECTOR, "nav.hidden.md\\:block"
        )
        for nav in desktop_navs:
            assert not nav.is_displayed(), "Desktop nav should be hidden on mobile"

    def test_mobile_header_visible(self, driver_mobile, base_url, credentials, screenshots_dir):
        """Mobile header bar should be visible."""
        self._login(driver_mobile, base_url, credentials, screenshots_dir)

        header = self._wait(driver_mobile).until(
            EC.presence_of_element_located(
                (By.CSS_SELECTOR, "header.md\\:hidden.sticky")
            )
        )
        assert header.is_displayed(), "Mobile header should be visible"
        driver_mobile.save_screenshot(f"{screenshots_dir}/mobile_header.png")

    def test_mobile_more_drawer_opens(self, driver_mobile, base_url, credentials, screenshots_dir):
        """Tapping 'More' in bottom nav opens the More drawer dialog."""
        self._login(driver_mobile, base_url, credentials, screenshots_dir)

        more_btn = self._wait(driver_mobile).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//button[normalize-space()='More' or .//span[normalize-space()='More']]")
            )
        )
        more_btn.click()
        time.sleep(1)

        dialog = self._wait(driver_mobile).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, ".fixed.inset-0"))
        )
        assert dialog.is_displayed(), "More drawer should open"
        driver_mobile.save_screenshot(f"{screenshots_dir}/mobile_more_drawer.png")

    def test_mobile_more_drawer_navigates(self, driver_mobile, base_url, credentials, screenshots_dir):
        """Tapping a nav item inside the More drawer navigates to that page."""
        self._login(driver_mobile, base_url, credentials, screenshots_dir)

        more_btn = self._wait(driver_mobile).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//button[normalize-space()='More' or .//span[normalize-space()='More']]")
            )
        )
        more_btn.click()
        time.sleep(1)

        groups_link = self._wait(driver_mobile).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//div[contains(@class,'fixed')]//a[contains(normalize-space(),'Groups')]")
            )
        )
        groups_link.click()
        time.sleep(2)

        assert "/groups" in driver_mobile.current_url, "Should navigate to /groups"
        assert_no_horizontal_overflow(driver_mobile, "groups")
        driver_mobile.save_screenshot(f"{screenshots_dir}/mobile_more_nav_groups.png")

    # ── Page-level overflow tests (require login) ──

    def test_mobile_dashboard_no_overflow(self, driver_mobile, base_url, credentials, screenshots_dir):
        """Dashboard page must not overflow at 390px."""
        self._login(driver_mobile, base_url, credentials, screenshots_dir)
        driver_mobile.get(f"{base_url}/app/dashboard")
        time.sleep(2)
        assert_no_horizontal_overflow(driver_mobile, "dashboard")
        driver_mobile.save_screenshot(f"{screenshots_dir}/mobile_dashboard.png")

    def test_mobile_groups_no_overflow(self, driver_mobile, base_url, credentials, screenshots_dir):
        """Groups page must not overflow at 390px."""
        self._login(driver_mobile, base_url, credentials, screenshots_dir)
        driver_mobile.get(f"{base_url}/app/groups")
        time.sleep(2)
        assert_no_horizontal_overflow(driver_mobile, "groups")
        driver_mobile.save_screenshot(f"{screenshots_dir}/mobile_groups.png")

    def test_mobile_budgets_no_overflow(self, driver_mobile, base_url, credentials, screenshots_dir):
        """Budgets page must not overflow at 390px."""
        self._login(driver_mobile, base_url, credentials, screenshots_dir)
        driver_mobile.get(f"{base_url}/app/budgets")
        time.sleep(2)
        assert_no_horizontal_overflow(driver_mobile, "budgets")
        driver_mobile.save_screenshot(f"{screenshots_dir}/mobile_budgets.png")

    def test_mobile_categories_no_overflow(self, driver_mobile, base_url, credentials, screenshots_dir):
        """Categories page must not overflow at 390px."""
        self._login(driver_mobile, base_url, credentials, screenshots_dir)
        driver_mobile.get(f"{base_url}/app/categories")
        time.sleep(2)
        assert_no_horizontal_overflow(driver_mobile, "categories")
        driver_mobile.save_screenshot(f"{screenshots_dir}/mobile_categories.png")

    def test_mobile_expenses_no_overflow(self, driver_mobile, base_url, credentials, screenshots_dir):
        """Expenses page must not overflow at 390px."""
        self._login(driver_mobile, base_url, credentials, screenshots_dir)
        driver_mobile.get(f"{base_url}/app/expenses")
        time.sleep(2)
        assert_no_horizontal_overflow(driver_mobile, "expenses")
        driver_mobile.save_screenshot(f"{screenshots_dir}/mobile_expenses.png")

    def test_mobile_expected_expenses_no_overflow(self, driver_mobile, base_url, credentials, screenshots_dir):
        """Expected expenses page must not overflow at 390px."""
        self._login(driver_mobile, base_url, credentials, screenshots_dir)
        driver_mobile.get(f"{base_url}/app/expected-expenses")
        time.sleep(2)
        assert_no_horizontal_overflow(driver_mobile, "expected-expenses")
        driver_mobile.save_screenshot(f"{screenshots_dir}/mobile_expected_expenses.png")

    # ── Narrow viewport (320px) page tests ──

    def test_narrow_dashboard_no_overflow(self, driver_mobile_narrow, base_url, credentials, screenshots_dir):
        """Dashboard must not overflow at 320px."""
        self._login(driver_mobile_narrow, base_url, credentials, screenshots_dir)
        driver_mobile_narrow.get(f"{base_url}/app/dashboard")
        time.sleep(2)
        assert_no_horizontal_overflow(driver_mobile_narrow, "dashboard-320")
        driver_mobile_narrow.save_screenshot(f"{screenshots_dir}/narrow_dashboard.png")

    def test_narrow_budgets_no_overflow(self, driver_mobile_narrow, base_url, credentials, screenshots_dir):
        """Budgets must not overflow at 320px."""
        self._login(driver_mobile_narrow, base_url, credentials, screenshots_dir)
        driver_mobile_narrow.get(f"{base_url}/app/budgets")
        time.sleep(2)
        assert_no_horizontal_overflow(driver_mobile_narrow, "budgets-320")
        driver_mobile_narrow.save_screenshot(f"{screenshots_dir}/narrow_budgets.png")

    def test_narrow_expenses_no_overflow(self, driver_mobile_narrow, base_url, credentials, screenshots_dir):
        """Expenses must not overflow at 320px."""
        self._login(driver_mobile_narrow, base_url, credentials, screenshots_dir)
        driver_mobile_narrow.get(f"{base_url}/app/expenses")
        time.sleep(2)
        assert_no_horizontal_overflow(driver_mobile_narrow, "expenses-320")
        driver_mobile_narrow.save_screenshot(f"{screenshots_dir}/narrow_expenses.png")

    # ── Modal dialog tests on mobile ──

    def test_mobile_modal_no_overflow(self, driver_mobile, base_url, credentials, screenshots_dir):
        """Modal dialogs must not overflow the 390px viewport."""
        self._login(driver_mobile, base_url, credentials, screenshots_dir)

        driver_mobile.get(f"{base_url}/app/groups")
        time.sleep(2)

        add_btn = self._wait(driver_mobile).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//button[normalize-space()='Add Group']")
            )
        )
        add_btn.click()
        time.sleep(1)

        dialog = self._wait(driver_mobile).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, ".fixed.inset-0"))
        )
        assert dialog.is_displayed(), "Add Group modal should appear"

        assert_no_horizontal_overflow(driver_mobile, "modal-add-group")
        driver_mobile.save_screenshot(f"{screenshots_dir}/mobile_modal.png")

    def test_mobile_modal_buttons_stacked(self, driver_mobile, base_url, credentials, screenshots_dir):
        """Modal action buttons should stack vertically on mobile (flex-col-reverse)."""
        self._login(driver_mobile, base_url, credentials, screenshots_dir)

        driver_mobile.get(f"{base_url}/app/groups")
        time.sleep(2)

        add_btn = self._wait(driver_mobile).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//button[normalize-space()='Add Group']")
            )
        )
        add_btn.click()
        time.sleep(1)

        self._wait(driver_mobile).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, ".fixed.inset-0"))
        )

        button_container = driver_mobile.find_element(
            By.CSS_SELECTOR, ".flex.flex-col-reverse"
        )
        assert button_container.is_displayed(), "Buttons should use flex-col-reverse on mobile"
        driver_mobile.save_screenshot(f"{screenshots_dir}/mobile_modal_buttons.png")

    def test_mobile_modal_touch_target_size(self, driver_mobile, base_url, credentials, screenshots_dir):
        """Modal buttons should meet 44px minimum touch target height."""
        self._login(driver_mobile, base_url, credentials, screenshots_dir)

        driver_mobile.get(f"{base_url}/app/groups")
        time.sleep(2)

        add_btn = self._wait(driver_mobile).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//button[normalize-space()='Add Group']")
            )
        )
        add_btn.click()
        time.sleep(1)

        self._wait(driver_mobile).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, ".fixed.inset-0"))
        )

        buttons = driver_mobile.find_elements(
            By.CSS_SELECTOR, ".fixed.inset-0 button.min-h-\\[44px\\]"
        )
        assert len(buttons) > 0, "Should find buttons with min-h-[44px] touch target class"
        for btn in buttons:
            height = btn.size.get("height", 0)
            assert height >= 44, f"Button '{btn.text}' height {height}px < 44px minimum"
        driver_mobile.save_screenshot(f"{screenshots_dir}/mobile_touch_targets.png")

    # ── Table horizontal scroll on mobile ──

    def test_mobile_expense_table_scrollable(self, driver_mobile, base_url, credentials, screenshots_dir):
        """Expense tables on mobile should not cause page-level overflow.

        Tables may scroll horizontally within their container, but the page
        itself must not overflow.
        """
        self._login(driver_mobile, base_url, credentials, screenshots_dir)

        driver_mobile.get(f"{base_url}/app/expenses")
        time.sleep(2)
        assert_no_horizontal_overflow(driver_mobile, "expenses-table")
        driver_mobile.save_screenshot(f"{screenshots_dir}/mobile_expense_table.png")
