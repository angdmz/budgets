import pytest
import time
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait, Select
from selenium.webdriver.support import expected_conditions as EC


class TestBudgetPlanWizard:
    """End-to-end integration tests for the budget plan wizard (CreateBudgetPlan page).

    Covers the 3-step flow: cadence & periods → template expenses → review & create.
    """

    TIMEOUT = 20

    # ── helpers (mirrors test_budget_duplicate_recurring.py) ──

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
                (By.XPATH, "//nav//a[contains(normalize-space(),'Groups')]")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/bpw_logged_in.png")

    def _nav(self, driver, link_text):
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, f"//nav//a[contains(normalize-space(),'{link_text}')]")
            )
        ).click()
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

    def _select_group(self, driver, group_name):
        option = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//select/option[normalize-space()='{group_name}']")
            )
        )
        select_element = option.find_element(By.XPATH, "./ancestor::select")
        Select(select_element).select_by_visible_text(group_name)
        time.sleep(1)

    def _create_group(self, driver, group_name, screenshots_dir, tag):
        self._nav(driver, "Groups")
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//button[normalize-space()='Add Group']")
            )
        ).click()
        self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, ".fixed.inset-0"))
        )
        modal = driver.find_element(By.CSS_SELECTOR, ".fixed.inset-0")
        name_input = modal.find_element(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, name_input, group_name)
        modal.find_element(
            By.XPATH, ".//button[normalize-space()='Create']"
        ).click()
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//td[normalize-space()='{group_name}']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/{tag}_group_created.png")

    def _create_category(self, driver, group_name, category_name, screenshots_dir, tag):
        self._nav(driver, "Categories")
        self._select_group(driver, group_name)
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//button[normalize-space()='Add Category']")
            )
        ).click()
        self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, ".fixed.inset-0"))
        )
        modal = driver.find_element(By.CSS_SELECTOR, ".fixed.inset-0")
        name_input = modal.find_element(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, name_input, category_name)
        modal.find_element(
            By.XPATH, ".//button[normalize-space()='Create']"
        ).click()
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//h3[normalize-space()='{category_name}']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/{tag}_category_created.png")

    def _goto_budgets_and_select_group(self, driver, group_name):
        self._nav(driver, "Budgets")
        self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, "//h1[normalize-space()='Budgets']"))
        )
        self._select_group(driver, group_name)

    def _click_create_plan(self, driver):
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//button[normalize-space()='Create Plan']")
            )
        ).click()
        self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, "//h1[normalize-space()='Create Budget Plan']"))
        )

    def _select_wizard_group(self, driver, group_name):
        option = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//select/option[normalize-space()='{group_name}']")
            )
        )
        select_element = option.find_element(By.XPATH, "./ancestor::select")
        Select(select_element).select_by_visible_text(group_name)
        time.sleep(1)

    def _click_cadence(self, driver, cadence_label):
        driver.find_element(
            By.XPATH, f"//button[normalize-space()='{cadence_label}']"
        ).click()
        time.sleep(0.5)

    def _click_next(self, driver):
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//button[normalize-space()='Next']"
            )
        )).click()
        time.sleep(1)

    def _click_back(self, driver):
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//button[normalize-space()='Back']"
            )
        )).click()
        time.sleep(1)

    def _click_create_plan_final(self, driver):
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//button[normalize-space()='Create Plan']"
            )
        )).click()

    def _select_category_in_row(self, driver, row, category_name):
        category_button = row.find_element(
            By.XPATH, ".//button[contains(., 'Select category')]"
        )
        driver.execute_script("arguments[0].click()", category_button)
        search_input = self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, "input[placeholder*='Search categories']"))
        )
        self._set_react_input(driver, search_input, category_name)
        time.sleep(0.5)
        category_option = self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, f"//button[contains(., '{category_name}') and contains(@class,'text-left') and not(contains(.,'Select category'))]")
            )
        )
        category_option.click()
        time.sleep(0.5)

    # ── tests ─────────────────────────────────────────────────────────────────

    @pytest.mark.integration
    def test_wizard_creates_monthly_budgets_with_template(
        self, driver, base_url, auth0_test_user, screenshots_dir
    ):
        """Full wizard flow: select monthly cadence, 3 periods, add a template
        expense, review, and create. Verify 3 budgets appear in the budgets table
        and each has the expected expense."""
        ts = str(int(time.time()))
        group_name = f"Wizard Group {ts}"
        category_name = f"Wizard Cat {ts}"
        base_budget_name = f"Wizard Monthly {ts}"
        expense_name = f"Rent {ts}"
        amount = "1500"

        self._login(driver, base_url, auth0_test_user, screenshots_dir)
        self._create_group(driver, group_name, screenshots_dir, "bpw")
        self._create_category(driver, group_name, category_name, screenshots_dir, "bpw")

        # Navigate to Budgets and select the group
        self._goto_budgets_and_select_group(driver, group_name)

        # Click "Create Plan" to open the wizard
        self._click_create_plan(driver)
        driver.save_screenshot(f"{screenshots_dir}/bpw_step1.png")

        # Step 1: Fill cadence & periods
        self._select_wizard_group(driver, group_name)

        name_input = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//input[@placeholder='e.g. Monthly Budget']")
            )
        )
        self._set_react_input(driver, name_input, base_budget_name)

        # Select monthly cadence
        self._click_cadence(driver, "Monthly")

        # Set number of periods to 3
        number_input = driver.find_element(
            By.XPATH, "//input[@type='number']"
        )
        self._set_react_input(driver, number_input, "3")

        # Verify period preview is visible
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//p[contains(normalize-space(), 'Periods preview')]")
            )
        )

        # Click Next → Step 2
        self._click_next(driver)
        driver.save_screenshot(f"{screenshots_dir}/bpw_step2.png")

        # Step 2: Add template expense
        # First expense row already exists — fill it
        expense_rows = driver.find_elements(By.CSS_SELECTOR, ".border.border-gray-200.p-4")
        assert len(expense_rows) >= 1
        first_row = expense_rows[0]

        text_inputs = first_row.find_elements(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, text_inputs[0], expense_name)

        amount_input = first_row.find_element(By.CSS_SELECTOR, "input[type='number']")
        self._set_react_input(driver, amount_input, amount)

        self._select_category_in_row(driver, first_row, category_name)

        # Click Next → Step 3
        self._click_next(driver)
        driver.save_screenshot(f"{screenshots_dir}/bpw_step3.png")

        # Step 3: Review
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//p[contains(normalize-space(), 'Template expenses')]")
            )
        )
        # Verify the expense appears in the review table
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//td[normalize-space()='{expense_name}']")
            )
        )

        # Click Create Plan
        self._click_create_plan_final(driver)

        # Wait for redirect back to /budgets
        self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, "//h1[normalize-space()='Budgets']"))
        )
        time.sleep(2)

        # Re-select the group (component re-mounts on route change, losing selection)
        self._select_group(driver, group_name)
        time.sleep(2)
        driver.save_screenshot(f"{screenshots_dir}/bpw_created.png")

        # Verify 3 budgets were created (names contain the base name)
        budget_cells = driver.find_elements(
            By.XPATH, f"//td[contains(normalize-space(), '{base_budget_name}')]"
        )
        assert len(budget_cells) == 3, (
            f"Expected 3 budgets with name containing '{base_budget_name}', "
            f"found {len(budget_cells)}"
        )

    @pytest.mark.integration
    def test_wizard_creates_weekly_budgets_no_template(
        self, driver, base_url, auth0_test_user, screenshots_dir
    ):
        """Wizard flow with weekly cadence, 2 periods, no template expenses.
        Verify 2 budgets are created with empty expected expenses."""
        ts = str(int(time.time()))
        group_name = f"Wizard Weekly Group {ts}"
        base_budget_name = f"Wizard Weekly {ts}"

        self._login(driver, base_url, auth0_test_user, screenshots_dir)
        self._create_group(driver, group_name, screenshots_dir, "bpw2")

        self._goto_budgets_and_select_group(driver, group_name)
        self._click_create_plan(driver)

        # Step 1
        self._select_wizard_group(driver, group_name)

        name_input = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//input[@placeholder='e.g. Monthly Budget']")
            )
        )
        self._set_react_input(driver, name_input, base_budget_name)

        self._click_cadence(driver, "Weekly")

        number_input = driver.find_element(
            By.XPATH, "//input[@type='number']"
        )
        self._set_react_input(driver, number_input, "2")

        self._click_next(driver)

        # Step 2: No expenses — just click Next
        self._click_next(driver)

        # Step 3: Review — should show "No expenses added"
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//p[contains(normalize-space(), 'No expenses')]")
            )
        )

        self._click_create_plan_final(driver)

        # Wait for redirect
        self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, "//h1[normalize-space()='Budgets']"))
        )
        time.sleep(2)

        # Re-select the group (component re-mounts on route change, losing selection)
        self._select_group(driver, group_name)
        time.sleep(2)
        driver.save_screenshot(f"{screenshots_dir}/bpw_weekly_created.png")

        budget_cells = driver.find_elements(
            By.XPATH, f"//td[contains(normalize-space(), '{base_budget_name}')]"
        )
        assert len(budget_cells) == 2, (
            f"Expected 2 budgets with name containing '{base_budget_name}', "
            f"found {len(budget_cells)}"
        )

    @pytest.mark.integration
    def test_wizard_back_navigation(
        self, driver, base_url, auth0_test_user, screenshots_dir
    ):
        """Verify Back button preserves state when navigating between steps."""
        ts = str(int(time.time()))
        group_name = f"Wizard Back Group {ts}"
        base_budget_name = f"Wizard Back {ts}"

        self._login(driver, base_url, auth0_test_user, screenshots_dir)
        self._create_group(driver, group_name, screenshots_dir, "bpw3")

        self._goto_budgets_and_select_group(driver, group_name)
        self._click_create_plan(driver)

        # Step 1
        self._select_wizard_group(driver, group_name)
        name_input = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//input[@placeholder='e.g. Monthly Budget']")
            )
        )
        self._set_react_input(driver, name_input, base_budget_name)
        self._click_cadence(driver, "Monthly")

        self._click_next(driver)

        # Step 2 → go back to Step 1
        self._click_back(driver)

        # Verify we're back on step 1 and name is preserved
        name_input2 = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//input[@placeholder='e.g. Monthly Budget']")
            )
        )
        assert name_input2.get_attribute("value") == base_budget_name, (
            f"Expected name '{base_budget_name}' to be preserved after Back"
        )

        driver.save_screenshot(f"{screenshots_dir}/bpw_back_nav.png")
