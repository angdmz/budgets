import pytest
import time
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.common.exceptions import TimeoutException, NoSuchElementException


class TestOnboardingFlow:
    """End-to-end integration tests for the onboarding wizard.

    Covers:
    - New user redirected to /onboarding
    - Full happy-path walkthrough (all 9 steps)
    - Skip-all flow clears the gate
    - Progress indicator advances
    - Draft expenses persist across page reload
    - Completed user is not redirected back
    - Reset returns user to welcome step
    - Back button navigates to previous step
    - Skip actual expense shows empty state on compare
    """

    TIMEOUT = 20
    DRAFT_KEY = "onboarding.draftExpenses"

    # ── helpers ─────────────────────────────────────────────────────────────

    def _wait(self, driver):
        return WebDriverWait(driver, self.TIMEOUT)

    def _login(self, driver, base_url, credentials, screenshots_dir):
        """Login without dismissing onboarding — the onboarding gate is the
        thing we want to test."""
        driver.get(f"{base_url}/app")
        time.sleep(2)

        if "auth0.com" in driver.current_url or "auth0" in driver.page_source.lower():
            try:
                email_field = WebDriverWait(driver, 10).until(
                    EC.presence_of_element_located(
                        (By.CSS_SELECTOR,
                         "input[name='username'], input[type='email'], input[id='username']")
                    )
                )
            except TimeoutException:
                pytest.skip("Auth0 login form did not appear")

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

            self._wait(driver).until(lambda d: "auth0.com" not in d.current_url)
            time.sleep(3)

        # Wait for the app's nav (Groups link is always present inside Layout,
        # even when on /onboarding).
        WebDriverWait(driver, 30).until(
            EC.presence_of_element_located(
                (By.XPATH, "//nav//a[contains(normalize-space(),'Groups')]")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/ob_login_done.png")

    def _reset_onboarding(self, driver, base_url, credentials):
        """POST /onboarding/reset via the SPA's API client.

        We use the browser's fetch to call the backend directly, reusing
        the Auth0 token already in the SPA's cache.
        """
        driver.get(f"{base_url}/app")
        time.sleep(1)
        # The SPA stores the Auth0 token under @@auth0spajs@@ keys in localStorage.
        # We can call the API from the browser context to reuse the session.
        result = driver.execute_script("""
            return (async () => {
                // Find the auth0 token in localStorage
                const keys = Object.keys(window.localStorage);
                const auth0Key = keys.find(k => k.startsWith('@@auth0spajs@@'));
                if (!auth0Key) return {ok: false, error: 'no auth0 token'};

                const cache = JSON.parse(window.localStorage.getItem(auth0Key));
                const token = cache?.body?.access_token;
                if (!token) return {ok: false, error: 'no access_token in cache'};

                const resp = await fetch(
                    window.location.origin + '/api/v1/onboarding/reset',
                    {
                        method: 'POST',
                        headers: {
                            'Authorization': 'Bearer ' + token,
                            'Content-Type': 'application/json',
                        },
                        body: '{}',
                    }
                );
                return {ok: resp.ok, status: resp.status};
            })();
        """)
        if not result.get("ok"):
            print(f"  [reset] Failed: {result}")
        time.sleep(1)

    def _set_react_input(self, driver, element, value):
        driver.execute_script(
            "var setter = Object.getOwnPropertyDescriptor("
            "  window.HTMLInputElement.prototype, 'value').set;"
            "setter.call(arguments[0], arguments[1]);"
            "arguments[0].dispatchEvent(new Event('input', {bubbles: true}));"
            "arguments[0].dispatchEvent(new Event('change', {bubbles: true}));",
            element,
            value,
        )

    def _click_skip(self, driver):
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//button[normalize-space()='Skip']")
            )
        ).click()
        time.sleep(1)

    def _click_next(self, driver):
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//button[normalize-space()='Next']")
            )
        ).click()
        time.sleep(1)

    def _click_start(self, driver):
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//button[normalize-space()='Get Started']")
            )
        ).click()
        time.sleep(1)

    def _click_go_to_dashboard(self, driver):
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//button[normalize-space()='Go to dashboard']")
            )
        ).click()
        time.sleep(2)

    def _get_progress_text(self, driver):
        return driver.find_element(
            By.XPATH, "//span[contains(normalize-space(),'/')]"
        ).text

    def _click_back(self, driver):
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//button[contains(normalize-space(),'Back')]")
            )
        ).click()
        time.sleep(1)

    def _select_category_combobox(self, driver, container, category_name, create=False):
        """Interact with the CategoryCombobox component.

        container: the WebElement that contains the combobox button.
        category_name: the category to select or create.
        create: if True, create the category if it doesn't exist.
        """
        # The combobox is a button inside a relative div
        combo_btn = container.find_element(
            By.XPATH, ".//div[contains(@class,'relative')]//button[@type='button']"
        )
        combo_btn.click()
        time.sleep(0.5)

        # Type the category name in the search input
        search_input = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//input[contains(@placeholder,'Search') or contains(@placeholder,'Buscar')]")
            )
        )
        self._set_react_input(driver, search_input, category_name)
        time.sleep(0.5)

        if create:
            # Click the "Create Category: ..." button (case-insensitive match)
            create_btn = self._wait(driver).until(
                EC.element_to_be_clickable(
                    (By.XPATH, f"//button[contains(translate(normalize-space(), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz'), 'create category') and contains(normalize-space(),'{category_name}')]")
                )
            )
            create_btn.click()
        else:
            # Click the matching category in the dropdown (scope to dropdown div)
            cat_btn = self._wait(driver).until(
                EC.element_to_be_clickable(
                    (By.XPATH, f"//div[contains(@class,'absolute')]//button[@type='button' and contains(normalize-space(),'{category_name}')]")
                )
            )
            cat_btn.click()
        time.sleep(1)

    # ── tests ───────────────────────────────────────────────────────────────

    @pytest.mark.integration
    def test_new_user_is_redirected_to_onboarding(
        self, driver, base_url, auth0_test_user, screenshots_dir
    ):
        """A fresh user should be redirected to /onboarding after login."""
        self._login(driver, base_url, auth0_test_user, screenshots_dir)
        self._reset_onboarding(driver, base_url, auth0_test_user)
        driver.get(f"{base_url}/app")
        time.sleep(3)

        # Wait for redirect to /onboarding
        self._wait(driver).until(
            lambda d: "/onboarding" in d.current_url
        )
        driver.save_screenshot(f"{screenshots_dir}/ob_redirect.png")

        # Verify the onboarding title is visible
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h1[contains(normalize-space(),'Welcome to MiPlatita')]")
            )
        )
        assert "/onboarding" in driver.current_url

    @pytest.mark.integration
    def test_full_happy_path(
        self, driver, base_url, auth0_test_user, screenshots_dir
    ):
        """Complete the full onboarding walkthrough:
        welcome → choose_group → choose_cadence → add_expected_expenses →
        budget_summary → register_actual_expense → compare_expenses →
        dashboard_tour → complete → /dashboard
        """
        ts = str(int(time.time()))
        group_name = f"OB Group {ts}"
        category_name = f"OB Cat {ts}"
        expense_name = f"OB Rent {ts}"
        amount = "1000"

        self._login(driver, base_url, auth0_test_user, screenshots_dir)
        self._reset_onboarding(driver, base_url, auth0_test_user)
        driver.get(f"{base_url}/app")
        time.sleep(3)

        # Ensure we're on onboarding
        self._wait(driver).until(
            lambda d: "/onboarding" in d.current_url
        )

        # Step 1: Welcome
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h2[normalize-space()='Welcome!']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/ob_step1_welcome.png")
        self._click_start(driver)

        # Step 2: Choose Group — create a new group
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h2[normalize-space()='What do you want to organize?']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/ob_step2_choose_group.png")
        name_input = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//input[@placeholder='e.g. Personal expenses, Family, etc.']")
            )
        )
        self._set_react_input(driver, name_input, group_name)
        driver.find_element(
            By.XPATH, "//button[normalize-space()='Create']"
        ).click()
        time.sleep(2)
        self._click_next(driver)

        # Step 3: Choose Cadence
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h2[normalize-space()='Choose the cadence']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/ob_step3_cadence.png")
        driver.find_element(
            By.XPATH, "//button[normalize-space()='Monthly']"
        ).click()
        time.sleep(0.5)
        self._click_next(driver)

        # Step 4: Add Expected Expenses
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h2[normalize-space()='Define your expected expenses']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/ob_step4_expenses.png")

        # Add an expected expense row
        driver.find_element(
            By.XPATH, "//button[contains(normalize-space(),'Add expected expense')]"
        ).click()
        time.sleep(0.5)

        # Fill the expense row
        rows = driver.find_elements(By.CSS_SELECTOR, ".border.border-gray-200.p-3")
        assert len(rows) >= 1
        row = rows[0]

        text_inputs = row.find_elements(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, text_inputs[0], expense_name)

        amount_input = row.find_element(By.CSS_SELECTOR, "input[type='number']")
        self._set_react_input(driver, amount_input, amount)

        # Create and select the category via CategoryCombobox
        self._select_category_combobox(driver, row, category_name, create=True)
        time.sleep(0.5)

        self._click_next(driver)

        # Step 5: Budget Summary
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h2[normalize-space()='This is your budget']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/ob_step5_summary.png")
        # Verify total expected is visible
        total_text = driver.find_element(
            By.XPATH, "//p[contains(normalize-space(),'Total expected')]"
        ).text
        assert "Total expected" in total_text
        self._click_next(driver)

        # Step 6: Register Actual Expense
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h2[normalize-space()='Register a real expense']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/ob_step6_actual.png")

        actual_name = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//input[@placeholder='Name'] | //input[contains(@placeholder,'name')]")
            )
        )
        self._set_react_input(driver, actual_name, f"OB Actual {ts}")

        actual_amount = driver.find_element(
            By.CSS_SELECTOR, "input[type='number']"
        )
        self._set_react_input(driver, actual_amount, "500")

        # Select category via CategoryCombobox
        self._select_category_combobox(driver, driver, category_name, create=False)
        time.sleep(0.5)

        driver.find_element(
            By.XPATH, "//button[normalize-space()='Register expense']"
        ).click()
        time.sleep(2)

        self._click_next(driver)

        # Step 7: Compare Expenses
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h2[normalize-space()='Expected vs. Actual']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/ob_step7_compare.png")
        self._click_next(driver)

        # Step 8: Dashboard Tour
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h2[normalize-space()='Get to know the dashboard']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/ob_step8_tour.png")
        self._click_next(driver)

        # Step 9: Complete
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h2[normalize-space()='All done!']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/ob_step9_complete.png")
        self._click_go_to_dashboard(driver)

        # Verify we're on /dashboard
        self._wait(driver).until(
            lambda d: "/dashboard" in d.current_url
        )
        assert "/onboarding" not in driver.current_url
        driver.save_screenshot(f"{screenshots_dir}/ob_dashboard.png")

    @pytest.mark.integration
    def test_skip_all_steps_completes_onboarding(
        self, driver, base_url, auth0_test_user, screenshots_dir
    ):
        """Skipping every step should complete onboarding and clear the gate."""
        self._login(driver, base_url, auth0_test_user, screenshots_dir)
        self._reset_onboarding(driver, base_url, auth0_test_user)
        driver.get(f"{base_url}/app")
        time.sleep(3)

        # Ensure we're on onboarding
        self._wait(driver).until(
            lambda d: "/onboarding" in d.current_url
        )

        # Skip through all steps
        for i in range(15):
            if "/onboarding" not in driver.current_url:
                break
            try:
                skip_btn = driver.find_element(
                    By.XPATH, "//button[normalize-space()='Skip']"
                )
                skip_btn.click()
                time.sleep(1)
                continue
            except NoSuchElementException:
                pass
            try:
                go_btn = driver.find_element(
                    By.XPATH, "//button[normalize-space()='Go to dashboard']"
                )
                go_btn.click()
                time.sleep(2)
                continue
            except NoSuchElementException:
                pass
            time.sleep(1)

        driver.save_screenshot(f"{screenshots_dir}/ob_skip_all.png")
        assert "/onboarding" not in driver.current_url

        # Verify we can navigate to /groups without being redirected
        driver.get(f"{base_url}/app/groups")
        time.sleep(2)
        assert "/onboarding" not in driver.current_url, (
            "User was redirected back to onboarding after skip-all — gate not cleared"
        )

    @pytest.mark.integration
    def test_progress_indicator_advances(
        self, driver, base_url, auth0_test_user, screenshots_dir
    ):
        """The '1 / 9' progress indicator should advance as steps complete."""
        self._login(driver, base_url, auth0_test_user, screenshots_dir)
        self._reset_onboarding(driver, base_url, auth0_test_user)
        driver.get(f"{base_url}/app")
        time.sleep(3)

        self._wait(driver).until(
            lambda d: "/onboarding" in d.current_url
        )

        # Step 1: should show "1 / 9"
        progress = self._get_progress_text(driver)
        assert progress.startswith("1"), f"Expected '1 / 9', got '{progress}'"
        driver.save_screenshot(f"{screenshots_dir}/ob_progress_1.png")

        self._click_start(driver)

        # Step 2: should show "2 / 9"
        progress = self._get_progress_text(driver)
        assert progress.startswith("2"), f"Expected '2 / 9', got '{progress}'"
        driver.save_screenshot(f"{screenshots_dir}/ob_progress_2.png")

    @pytest.mark.integration
    def test_draft_expenses_persist_across_reload(
        self, driver, base_url, auth0_test_user, screenshots_dir
    ):
        """Draft expected expenses should survive a page reload (localStorage)."""
        self._login(driver, base_url, auth0_test_user, screenshots_dir)
        self._reset_onboarding(driver, base_url, auth0_test_user)
        driver.get(f"{base_url}/app")
        time.sleep(3)

        self._wait(driver).until(
            lambda d: "/onboarding" in d.current_url
        )

        # Advance to add_expected_expenses step
        self._click_start(driver)

        # Create a group to proceed
        ts = str(int(time.time()))
        group_name = f"OB Draft {ts}"
        name_input = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//input[@placeholder='e.g. Personal expenses, Family, etc.']")
            )
        )
        self._set_react_input(driver, name_input, group_name)
        driver.find_element(
            By.XPATH, "//button[normalize-space()='Create']"
        ).click()
        time.sleep(2)
        self._click_next(driver)

        # Choose cadence
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//button[normalize-space()='Monthly']")
            )
        ).click()
        time.sleep(0.5)
        self._click_next(driver)

        # Now on add_expected_expenses
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h2[normalize-space()='Define your expected expenses']")
            )
        )

        # Add a draft expense row
        driver.find_element(
            By.XPATH, "//button[contains(normalize-space(),'Add expected expense')]"
        ).click()
        time.sleep(0.5)

        rows = driver.find_elements(By.CSS_SELECTOR, ".border.border-gray-200.p-3")
        row = rows[0]
        text_inputs = row.find_elements(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, text_inputs[0], f"OB Draft Expense {ts}")

        # Verify localStorage has the draft
        draft = driver.execute_script(
            "return window.localStorage.getItem('onboarding.draftExpenses');"
        )
        assert draft is not None, "Draft not saved to localStorage"
        assert f"OB Draft Expense {ts}" in draft

        # Reload the page
        driver.refresh()
        time.sleep(3)

        # We should be back on onboarding (gate still active)
        self._wait(driver).until(
            lambda d: "/onboarding" in d.current_url
        )

        # After reload, the onboarding should resume at add_expected_expenses
        # (the step we were on before the reload).
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h2[normalize-space()='Define your expected expenses']")
            )
        )

        # Verify the draft was restored from localStorage
        rows = driver.find_elements(By.CSS_SELECTOR, ".border.border-gray-200.p-3")
        assert len(rows) >= 1, "Draft expense not restored after reload"
        text_inputs = rows[0].find_elements(By.CSS_SELECTOR, "input[type='text']")
        restored_name = text_inputs[0].get_attribute("value")
        assert f"OB Draft Expense {ts}" in restored_name, (
            f"Draft not restored: expected 'OB Draft Expense {ts}', got '{restored_name}'"
        )
        driver.save_screenshot(f"{screenshots_dir}/ob_draft_restored.png")

    @pytest.mark.integration
    def test_completed_user_is_not_redirected(
        self, driver, base_url, auth0_test_user, screenshots_dir
    ):
        """After completing onboarding, navigating to /onboarding should
        redirect to /dashboard (status === 'completed' branch)."""
        self._login(driver, base_url, auth0_test_user, screenshots_dir)
        self._reset_onboarding(driver, base_url, auth0_test_user)
        driver.get(f"{base_url}/app")
        time.sleep(3)

        # Skip through all onboarding
        self._wait(driver).until(
            lambda d: "/onboarding" in d.current_url
        )
        for i in range(15):
            if "/onboarding" not in driver.current_url:
                break
            try:
                skip_btn = driver.find_element(
                    By.XPATH, "//button[normalize-space()='Skip']"
                )
                skip_btn.click()
                time.sleep(1)
                continue
            except NoSuchElementException:
                pass
            try:
                go_btn = driver.find_element(
                    By.XPATH, "//button[normalize-space()='Go to dashboard']"
                )
                go_btn.click()
                time.sleep(2)
                continue
            except NoSuchElementException:
                pass
            time.sleep(1)

        # Now try to navigate to /onboarding directly
        driver.get(f"{base_url}/app/onboarding")
        time.sleep(3)

        # Should be redirected to /dashboard
        assert "/dashboard" in driver.current_url, (
            f"Expected redirect to /dashboard, got {driver.current_url}"
        )
        driver.save_screenshot(f"{screenshots_dir}/ob_completed_redirect.png")

    @pytest.mark.integration
    def test_reset_returns_user_to_welcome(
        self, driver, base_url, auth0_test_user, screenshots_dir
    ):
        """POST /onboarding/reset should return the user to the welcome step."""
        self._login(driver, base_url, auth0_test_user, screenshots_dir)
        self._reset_onboarding(driver, base_url, auth0_test_user)
        driver.get(f"{base_url}/app")
        time.sleep(3)

        # Skip through all onboarding
        self._wait(driver).until(
            lambda d: "/onboarding" in d.current_url
        )
        for i in range(15):
            if "/onboarding" not in driver.current_url:
                break
            try:
                skip_btn = driver.find_element(
                    By.XPATH, "//button[normalize-space()='Skip']"
                )
                skip_btn.click()
                time.sleep(1)
                continue
            except NoSuchElementException:
                pass
            try:
                go_btn = driver.find_element(
                    By.XPATH, "//button[normalize-space()='Go to dashboard']"
                )
                go_btn.click()
                time.sleep(2)
                continue
            except NoSuchElementException:
                pass
            time.sleep(1)

        # Now reset onboarding
        self._reset_onboarding(driver, base_url, auth0_test_user)

        # Navigate to /app — should redirect to /onboarding
        driver.get(f"{base_url}/app")
        time.sleep(3)

        self._wait(driver).until(
            lambda d: "/onboarding" in d.current_url
        )

        # Verify we're back on the welcome step
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h2[normalize-space()='Welcome!']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/ob_reset_welcome.png")
        assert "/onboarding" in driver.current_url

    @pytest.mark.integration
    def test_back_navigation(
        self, driver, base_url, auth0_test_user, screenshots_dir
    ):
        """Back button should navigate to the previous step."""
        self._login(driver, base_url, auth0_test_user, screenshots_dir)
        self._reset_onboarding(driver, base_url, auth0_test_user)
        driver.get(f"{base_url}/app")
        time.sleep(3)

        self._wait(driver).until(
            lambda d: "/onboarding" in d.current_url
        )

        # Step 1: Welcome → advance
        self._click_start(driver)
        assert self._get_progress_text(driver).startswith("2")

        # Step 2: Choose Group — create and advance
        ts = str(int(time.time()))
        group_name = f"OB Back {ts}"
        name_input = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//input[@placeholder='e.g. Personal expenses, Family, etc.']")
            )
        )
        self._set_react_input(driver, name_input, group_name)
        driver.find_element(
            By.XPATH, "//button[normalize-space()='Create']"
        ).click()
        time.sleep(2)
        self._click_next(driver)
        assert self._get_progress_text(driver).startswith("3")

        # Step 3: Choose Cadence — now go back
        self._click_back(driver)
        time.sleep(1)

        # Should be back on Choose Group (step 2)
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h2[normalize-space()='What do you want to organize?']")
            )
        )
        assert self._get_progress_text(driver).startswith("2")
        driver.save_screenshot(f"{screenshots_dir}/ob_back_to_step2.png")

        # Go back again — should be on Welcome (step 1)
        self._click_back(driver)
        time.sleep(1)
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h2[normalize-space()='Welcome!']")
            )
        )
        assert self._get_progress_text(driver).startswith("1")
        driver.save_screenshot(f"{screenshots_dir}/ob_back_to_step1.png")

    @pytest.mark.integration
    def test_skip_actual_expense(
        self, driver, base_url, auth0_test_user, screenshots_dir
    ):
        """Skipping the actual expense step should show empty state on compare."""
        self._login(driver, base_url, auth0_test_user, screenshots_dir)
        self._reset_onboarding(driver, base_url, auth0_test_user)
        driver.get(f"{base_url}/app")
        time.sleep(3)

        self._wait(driver).until(
            lambda d: "/onboarding" in d.current_url
        )

        # Advance through welcome → choose_group → choose_cadence → add_expected_expenses → budget_summary
        self._click_start(driver)

        ts = str(int(time.time()))
        group_name = f"OB Skip {ts}"
        category_name = f"OB SkipCat {ts}"
        expense_name = f"OB SkipExp {ts}"

        name_input = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//input[@placeholder='e.g. Personal expenses, Family, etc.']")
            )
        )
        self._set_react_input(driver, name_input, group_name)
        driver.find_element(
            By.XPATH, "//button[normalize-space()='Create']"
        ).click()
        time.sleep(2)
        self._click_next(driver)

        # Choose cadence
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//button[normalize-space()='Monthly']")
            )
        ).click()
        time.sleep(0.5)
        self._click_next(driver)

        # Add expected expenses
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h2[normalize-space()='Define your expected expenses']")
            )
        )
        driver.find_element(
            By.XPATH, "//button[contains(normalize-space(),'Add expected expense')]"
        ).click()
        time.sleep(0.5)
        rows = driver.find_elements(By.CSS_SELECTOR, ".border.border-gray-200.p-3")
        row = rows[0]
        text_inputs = row.find_elements(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, text_inputs[0], expense_name)
        amount_input = row.find_element(By.CSS_SELECTOR, "input[type='number']")
        self._set_react_input(driver, amount_input, "1000")
        self._select_category_combobox(driver, row, category_name, create=True)
        time.sleep(0.5)
        self._click_next(driver)

        # Budget summary
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h2[normalize-space()='This is your budget']")
            )
        )
        self._click_next(driver)

        # Register actual expense — skip it
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h2[normalize-space()='Register a real expense']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/ob_skip_actual_step.png")
        self._click_skip(driver)

        # Compare expenses — should show empty state
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h2[normalize-space()='Expected vs. Actual']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/ob_skip_actual_compare.png")

        # Verify the "no actual expenses" message is shown
        no_actual_text = driver.find_element(
            By.XPATH, "//p[contains(normalize-space(),'No worries')]"
        )
        assert no_actual_text is not None
        self._click_next(driver)
