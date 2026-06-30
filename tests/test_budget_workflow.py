import pytest
import time
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait, Select
from selenium.webdriver.support import expected_conditions as EC
from selenium.common.exceptions import NoSuchElementException, TimeoutException


class TestBudgetWorkflow:
    """End-to-end integration test covering the full budget workflow:
    login → create group → create budget → create category → add expense → view budget.
    """

    TIMEOUT = 20

    # ── internal helpers ───────────────────────────────────────────────────────

    def _wait(self, driver):
        return WebDriverWait(driver, self.TIMEOUT)

    def _login(self, driver, base_url, credentials, screenshots_dir):
        """Navigate to /app and complete Auth0 Universal Login if redirected."""
        driver.get(f"{base_url}/app")
        time.sleep(2)
        driver.save_screenshot(f"{screenshots_dir}/wf_01_pre_login.png")

        if "auth0.com" in driver.current_url or "auth0" in driver.page_source.lower():
            # Auth0 returns 403 when redirect_uri is not registered in the tenant.
            # Detect this by waiting briefly and checking whether a login input
            # ever appears; if not, skip the test with an actionable message.
            try:
                email_field = WebDriverWait(driver, 10).until(
                    EC.presence_of_element_located(
                        (By.CSS_SELECTOR,
                         "input[name='username'], input[type='email'], input[id='username']")
                    )
                )
            except TimeoutException:
                driver.save_screenshot(f"{screenshots_dir}/wf_auth0_error.png")
                pytest.skip(
                    "Auth0 login form did not appear — the redirect_uri "
                    f"'{driver.current_url.split('redirect_uri=')[0]}' may not be "
                    "registered in the Auth0 tenant (add the test base URL as an "
                    "Allowed Callback URL in the Auth0 dashboard)."
                )
            email_field.clear()
            email_field.send_keys(credentials["email"])

            # Auth0 Universal Login may use identifier-first flow where the
            # password field only becomes visible/interactive after submitting
            # the email.  Always click Continue first, then wait for the
            # password field to be *visible* (not just present in the DOM).
            driver.find_element(By.CSS_SELECTOR, "button[type='submit']").click()

            pwd_field = self._wait(driver).until(
                EC.visibility_of_element_located(
                    (By.CSS_SELECTOR, "input[type='password']")
                )
            )

            pwd_field.clear()
            pwd_field.send_keys(credentials["password"])
            driver.find_element(By.CSS_SELECTOR, "button[type='submit']").click()

            # Wait until the browser leaves the Auth0 domain
            try:
                self._wait(driver).until(lambda d: "auth0.com" not in d.current_url)
            except TimeoutException:
                driver.save_screenshot(f"{screenshots_dir}/wf_login_stuck.png")
                page_text = driver.find_element(By.TAG_NAME, "body").text[:400]
                pytest.fail(
                    f"Browser did not leave auth0.com after submitting credentials.\n"
                    f"Current URL: {driver.current_url}\n"
                    f"Page text snippet: {page_text}"
                )
            time.sleep(3)

        # After the Auth0 callback, the SPA exchanges the authorization code
        # for tokens (network round-trip to Auth0) and then re-renders the
        # authenticated layout.  This can take several seconds, so use a
        # generous timeout.
        try:
            WebDriverWait(driver, 30).until(
                EC.presence_of_element_located(
                    (By.XPATH, "//nav//a[contains(normalize-space(),'Groups')]")
                )
            )
        except TimeoutException:
            driver.save_screenshot(f"{screenshots_dir}/wf_post_login_timeout.png")
            page_text = driver.find_element(By.TAG_NAME, "body").text[:500]
            pytest.fail(
                f"Authenticated app layout did not appear after login.\n"
                f"Current URL: {driver.current_url}\n"
                f"Page title: {driver.title}\n"
                f"Page text snippet: {page_text}"
            )
        driver.save_screenshot(f"{screenshots_dir}/wf_01_logged_in.png")
        print("\n  ✓ Logged in")

    def _nav(self, driver, link_text):
        """Click a top-nav link by its visible text and wait for the page to settle."""
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, f"//nav//a[contains(normalize-space(),'{link_text}')]")
            )
        ).click()
        time.sleep(1)

    def _open_modal(self, driver, button_text):
        """Click a button and wait for the modal overlay to appear."""
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, f"//button[normalize-space()='{button_text}']")
            )
        ).click()
        self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, ".fixed.inset-0"))
        )

    def _submit_modal(self, driver):
        """Click the Create button inside the modal and wait for it to close."""
        driver.find_element(
            By.XPATH,
            "//div[contains(@class,'fixed') and contains(@class,'inset-0')]"
            "//button[normalize-space()='Create']",
        ).click()
        self._wait(driver).until(
            EC.invisibility_of_element_located((By.CSS_SELECTOR, ".fixed.inset-0"))
        )

    def _set_react_input(self, driver, element, value):
        """Set a React-controlled text/number input by dispatching native events.

        Mirrors _set_react_date: uses the native value setter so that React's
        internal fiber state is updated, then dispatches synthetic-compatible
        events to trigger the onChange handler.
        """
        driver.execute_script(
            "var setter = Object.getOwnPropertyDescriptor("
            "  window.HTMLInputElement.prototype, 'value').set;"
            "setter.call(arguments[0], arguments[1]);"
            "arguments[0].dispatchEvent(new Event('input', {bubbles: true}));"
            "arguments[0].dispatchEvent(new Event('change', {bubbles: true}));",
            element,
            value,
        )

    def _set_react_date(self, driver, element, date_str):
        """Set a React-controlled date input by dispatching native events.

        Plain send_keys on date inputs is unreliable with React because React
        intercepts the native value setter. Using the native value descriptor
        and dispatching synthetic events ensures the controlled component state
        is updated correctly.
        """
        driver.execute_script(
            "var setter = Object.getOwnPropertyDescriptor("
            "  window.HTMLInputElement.prototype, 'value').set;"
            "setter.call(arguments[0], arguments[1]);"
            "arguments[0].dispatchEvent(new Event('input', {bubbles: true}));"
            "arguments[0].dispatchEvent(new Event('change', {bubbles: true}));",
            element,
            date_str,
        )

    def _select_group(self, driver, group_name, select_element=None):
        """Pick a group by name from the first <select> on the page (or the given element)."""
        if select_element is None:
            select_element = self._wait(driver).until(
                EC.presence_of_element_located((By.CSS_SELECTOR, "select"))
            )
        Select(select_element).select_by_visible_text(group_name)
        time.sleep(1)

    # ── main test ──────────────────────────────────────────────────────────────

    @pytest.mark.integration
    def test_create_group_budget_category_expense_and_view(
        self, driver, base_url, auth0_test_user, screenshots_dir
    ):
        """Full workflow:
        1. Log in via Auth0
        2. Create a group
        3. Create a budget inside the group
        4. Create a category inside the group
        5. Add an expense to the budget
        6. Verify the budget is visible in the Groups tab budget viewer
        """
        ts = str(int(time.time()))
        group_name    = f"Test Group {ts}"
        budget_name   = f"Test Budget {ts}"
        category_name = f"Test Category {ts}"
        expense_name  = f"Test Expense {ts}"

        # ── 1. Login ───────────────────────────────────────────────────────────
        self._login(driver, base_url, auth0_test_user, screenshots_dir)

        # ── 2. Create group ────────────────────────────────────────────────────
        self._nav(driver, "Groups")
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h1[normalize-space()='Groups']")
            )
        )

        self._open_modal(driver, "Add Group")
        modal = driver.find_element(By.CSS_SELECTOR, ".fixed.inset-0")
        name_input = modal.find_element(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, name_input, group_name)
        driver.save_screenshot(f"{screenshots_dir}/wf_02_group_modal.png")
        self._submit_modal(driver)

        # Group name must appear in the groups table
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//td[normalize-space()='{group_name}']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/wf_02_group_created.png")
        print(f"  ✓ Group created: {group_name}")

        # ── 3. Create budget ───────────────────────────────────────────────────
        self._nav(driver, "Budgets")
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h1[normalize-space()='Budgets']")
            )
        )

        self._select_group(driver, group_name)

        self._open_modal(driver, "Add Budget")
        modal = driver.find_element(By.CSS_SELECTOR, ".fixed.inset-0")

        name_input = modal.find_element(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, name_input, budget_name)

        date_inputs = modal.find_elements(By.CSS_SELECTOR, "input[type='date']")
        self._set_react_date(driver, date_inputs[0], "2025-01-01")
        self._set_react_date(driver, date_inputs[1], "2025-12-31")

        driver.save_screenshot(f"{screenshots_dir}/wf_03_budget_modal.png")
        self._submit_modal(driver)

        # Budget name must appear in the budgets table
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//td[normalize-space()='{budget_name}']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/wf_03_budget_created.png")
        print(f"  ✓ Budget created: {budget_name}")

        # ── 4. Create category ─────────────────────────────────────────────────
        self._nav(driver, "Categories")
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h1[normalize-space()='Categories']")
            )
        )

        self._select_group(driver, group_name)

        self._open_modal(driver, "Add Category")
        modal = driver.find_element(By.CSS_SELECTOR, ".fixed.inset-0")
        name_input = modal.find_element(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, name_input, category_name)
        driver.save_screenshot(f"{screenshots_dir}/wf_04_category_modal.png")
        self._submit_modal(driver)

        # Category name must appear as a card heading
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//h3[normalize-space()='{category_name}']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/wf_04_category_created.png")
        print(f"  ✓ Category created: {category_name}")

        # ── 5. Add expense ─────────────────────────────────────────────────────
        self._nav(driver, "Expenses")
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h1[normalize-space()='Expenses']")
            )
        )

        # Expenses page has two selects side-by-side: group then budget.
        # Wait until both are rendered.
        self._wait(driver).until(
            lambda d: len(d.find_elements(By.CSS_SELECTOR, "select")) >= 2
        )
        selects = driver.find_elements(By.CSS_SELECTOR, "select")

        Select(selects[0]).select_by_visible_text(group_name)
        time.sleep(1)  # budget select populates after group is chosen

        # Re-fetch selects after group selection triggers re-render
        selects = driver.find_elements(By.CSS_SELECTOR, "select")
        Select(selects[1]).select_by_visible_text(budget_name)
        time.sleep(1)

        self._open_modal(driver, "Add Expense")
        modal = driver.find_element(By.CSS_SELECTOR, ".fixed.inset-0")

        name_input = modal.find_element(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, name_input, expense_name)

        amount_input = modal.find_element(By.CSS_SELECTOR, "input[type='number']")
        self._set_react_input(driver, amount_input, "42.50")

        date_input = modal.find_element(By.CSS_SELECTOR, "input[type='date']")
        self._set_react_date(driver, date_input, "2025-06-15")

        driver.save_screenshot(f"{screenshots_dir}/wf_05_expense_modal.png")
        self._submit_modal(driver)

        # Expense name must appear in the expenses table
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//td[normalize-space()='{expense_name}']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/wf_05_expense_created.png")
        print(f"  ✓ Expense created: {expense_name}")

        # ── 6. View budget in Groups tab ───────────────────────────────────────
        self._nav(driver, "Groups")
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h1[normalize-space()='Groups']")
            )
        )

        # The Groups page has a dedicated "Budgets" section below the groups
        # table with its own group selector. Target it via the h2 heading.
        budget_section_select = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h2[normalize-space()='Budgets']/following::select[1]")
            )
        )
        Select(budget_section_select).select_by_visible_text(group_name)
        time.sleep(1)

        # Budget name must appear in the budget list for the selected group
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//td[normalize-space()='{budget_name}']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/wf_06_budget_visible_in_groups.png")
        print(f"  ✓ Budget visible in Groups tab: {budget_name}")
        print("\n✅ Full workflow test passed!")
