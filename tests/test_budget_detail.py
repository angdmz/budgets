import pytest
import time
import uuid
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait, Select
from selenium.webdriver.support import expected_conditions as EC


class TestBudgetDetail:
    """End-to-end integration tests for the Budget Detail page.

    Covers navigation from Budgets list → Budget Detail, adding/editing/deleting
    expected and actual expenses.
    """

    TIMEOUT = 20

    # ── helpers ──

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
        driver.save_screenshot(f"{screenshots_dir}/bd_logged_in.png")

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

    def _create_budget(self, driver, group_name, budget_name, screenshots_dir, tag):
        """Create a simple budget via the Add Budget modal on the Budgets page."""
        self._nav(driver, "Budgets")
        self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, "//h1[normalize-space()='Budgets']"))
        )
        self._select_group(driver, group_name)
        time.sleep(1)

        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//button[normalize-space()='Add Budget']")
            )
        ).click()
        self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, ".fixed.inset-0"))
        )
        modal = driver.find_element(By.CSS_SELECTOR, ".fixed.inset-0")
        name_input = modal.find_element(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, name_input, budget_name)

        # Set start/end dates
        date_inputs = modal.find_elements(By.CSS_SELECTOR, "input[type='date']")
        if len(date_inputs) >= 2:
            self._set_react_input(driver, date_inputs[0], "2025-01-01")
            self._set_react_input(driver, date_inputs[1], "2025-01-31")

        modal.find_element(
            By.XPATH, ".//button[normalize-space()='Create']"
        ).click()
        time.sleep(2)
        driver.save_screenshot(f"{screenshots_dir}/{tag}_budget_created.png")

    def _click_budget_name(self, driver, budget_name):
        """Click a budget name in the Budgets table to navigate to its detail page."""
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, f"//td//button[contains(normalize-space(),'{budget_name}')]")
            )
        ).click()
        time.sleep(2)

    def _add_expected_expense(self, driver, name, amount, screenshots_dir, tag):
        """Add an expected expense via the Budget Detail page."""
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//h2[normalize-space()='Expected Expenses']"
                "//following-sibling::button[normalize-space()='Add Expected Expense']"
                )
            )
        ).click()
        self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, ".fixed.inset-0"))
        )
        modal = driver.find_element(By.CSS_SELECTOR, ".fixed.inset-0")
        name_input = modal.find_element(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, name_input, name)

        amount_input = modal.find_element(By.CSS_SELECTOR, "input[type='number']")
        self._set_react_input(driver, amount_input, amount)

        # Select first available category in the combobox
        self._select_first_category_in_modal(driver, modal)

        modal.find_element(
            By.XPATH, ".//button[normalize-space()='Create']"
        ).click()
        time.sleep(2)
        driver.save_screenshot(f"{screenshots_dir}/{tag}_ee_added.png")

    def _add_actual_expense(self, driver, name, amount, screenshots_dir, tag):
        """Add an actual expense via the Budget Detail page."""
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//h2[normalize-space()='Actual Expenses']"
                "//following-sibling::button[normalize-space()='Add Expense']"
                )
            )
        ).click()
        self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, ".fixed.inset-0"))
        )
        modal = driver.find_element(By.CSS_SELECTOR, ".fixed.inset-0")
        name_input = modal.find_element(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, name_input, name)

        amount_input = modal.find_element(By.CSS_SELECTOR, "input[type='number']")
        self._set_react_input(driver, amount_input, amount)

        # Set expense date
        date_input = modal.find_element(By.CSS_SELECTOR, "input[type='date']")
        self._set_react_input(driver, date_input, "2025-01-15")

        # Select first available category
        self._select_first_category_in_modal(driver, modal)

        modal.find_element(
            By.XPATH, ".//button[normalize-space()='Create']"
        ).click()
        time.sleep(2)
        driver.save_screenshot(f"{screenshots_dir}/{tag}_ae_added.png")

    def _select_first_category_in_modal(self, driver, modal):
        """Click the category combobox and select the first option."""
        # The CategoryCombobox renders a button with text "Select category" or the selected name
        combobox_btn = modal.find_element(
            By.XPATH, ".//button[contains(normalize-space(),'Select category') or contains(normalize-space(),'Seleccionar categoría')]"
        )
        combobox_btn.click()
        time.sleep(1)
        # The dropdown options are buttons inside a div with max-h-60
        first_option = self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//div[contains(@class,'max-h-60')]//button[contains(@class,'hover')]")
            )
        )
        first_option.click()
        time.sleep(0.5)

    def _delete_expense_by_name(self, driver, expense_name, section_title):
        """Click the Delete button in the row containing expense_name within a section."""
        row = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//td[normalize-space()='{expense_name}']/ancestor::tr")
            )
        )
        delete_btn = row.find_element(
            By.XPATH, ".//button[normalize-space()='Delete']"
        )
        delete_btn.click()
        time.sleep(1)

        # Confirm in the delete modal
        confirm_modal = self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, ".fixed.inset-0"))
        )
        confirm_modal.find_element(
            By.XPATH, ".//button[normalize-space()='Delete']"
        ).click()
        time.sleep(2)

    # ── tests ──

    @pytest.mark.order(1)
    def test_budget_detail_add_and_delete_expenses(
        self, driver, base_url, credentials, screenshots_dir
    ):
        """Navigate to budget detail, add expected + actual expenses, verify totals, then delete."""
        tag = "bd"
        unique = uuid.uuid4().hex[:10]
        group_name = f"DetailGroup {unique}"
        category_name = f"DetailCat {unique}"
        budget_name = f"DetailBudget {unique}"

        self._login(driver, base_url, credentials, screenshots_dir)
        self._create_group(driver, group_name, screenshots_dir, tag)
        self._create_category(driver, group_name, category_name, screenshots_dir, tag)
        self._create_budget(driver, group_name, budget_name, screenshots_dir, tag)

        # Navigate to Budgets page and click the budget name
        self._nav(driver, "Budgets")
        self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, "//h1[normalize-space()='Budgets']"))
        )
        self._select_group(driver, group_name)
        time.sleep(2)
        driver.save_screenshot(f"{screenshots_dir}/{tag}_budgets_list.png")

        self._click_budget_name(driver, budget_name)

        # Verify we're on the detail page
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//h1[normalize-space()='{budget_name}']")
            )
        )
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//button[contains(normalize-space(),'Back to Budgets')]")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/{tag}_detail_page.png")

        # Add expected expense
        ee_name = f"Expected {unique}"
        self._add_expected_expense(driver, ee_name, "500.00", screenshots_dir, tag)

        # Verify expected expense appears in the table
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//td[normalize-space()='{ee_name}']")
            )
        )

        # Add actual expense
        ae_name = f"Actual {unique}"
        self._add_actual_expense(driver, ae_name, "300.00", screenshots_dir, tag)

        # Verify actual expense appears in the table
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//td[normalize-space()='{ae_name}']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/{tag}_both_expenses.png")

        # Delete the expected expense
        self._delete_expense_by_name(driver, ee_name, "Expected Expenses")
        # Verify it's gone
        remaining_ee = driver.find_elements(
            By.XPATH, f"//td[normalize-space()='{ee_name}']"
        )
        assert len(remaining_ee) == 0, f"Expected expense '{ee_name}' should have been deleted"

        # Delete the actual expense
        self._delete_expense_by_name(driver, ae_name, "Actual Expenses")
        remaining_ae = driver.find_elements(
            By.XPATH, f"//td[normalize-space()='{ae_name}']"
        )
        assert len(remaining_ae) == 0, f"Actual expense '{ae_name}' should have been deleted"

        driver.save_screenshot(f"{screenshots_dir}/{tag}_after_delete.png")

    @pytest.mark.order(2)
    def test_budget_detail_back_navigation(
        self, driver, base_url, credentials, screenshots_dir
    ):
        """Verify the Back to Budgets button returns to the Budgets list."""
        tag = "bd_back"
        unique = uuid.uuid4().hex[:10]
        group_name = f"BackGroup {unique}"
        budget_name = f"BackBudget {unique}"

        self._login(driver, base_url, credentials, screenshots_dir)
        self._create_group(driver, group_name, screenshots_dir, tag)
        self._create_budget(driver, group_name, budget_name, screenshots_dir, tag)

        # Navigate to budget detail
        self._nav(driver, "Budgets")
        self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, "//h1[normalize-space()='Budgets']"))
        )
        self._select_group(driver, group_name)
        time.sleep(2)
        self._click_budget_name(driver, budget_name)

        # Verify detail page
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//h1[normalize-space()='{budget_name}']")
            )
        )

        # Click Back to Budgets
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//button[contains(normalize-space(),'Back to Budgets')]")
            )
        ).click()
        time.sleep(2)

        # Verify we're back on the Budgets list
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//h1[normalize-space()='Budgets']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/{tag}_back_to_list.png")
