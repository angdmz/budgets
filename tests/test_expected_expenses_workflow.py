import pytest
import time
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait, Select
from selenium.webdriver.support import expected_conditions as EC


class TestExpectedExpensesWorkflow:
    """End-to-end integration test for Expected Expenses CRUD:
    login → create group/budget → create expected expense → edit → delete.
    """

    TIMEOUT = 20

    def _wait(self, driver):
        return WebDriverWait(driver, self.TIMEOUT)

    def _login(self, driver, base_url, credentials, screenshots_dir):
        """Navigate to /app and complete Auth0 Universal Login if redirected."""
        driver.get(f"{base_url}/app")
        time.sleep(2)

        if "auth0.com" in driver.current_url or "auth0" in driver.page_source.lower():
            try:
                email_field = WebDriverWait(driver, 10).until(
                    EC.presence_of_element_located(
                        (By.CSS_SELECTOR, "input[name='username'], input[type='email'], input[id='username']")
                    )
                )
            except:
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
        driver.save_screenshot(f"{screenshots_dir}/ee_logged_in.png")

    def _nav(self, driver, link_text):
        """Click a top-nav link by its visible text."""
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

    def _submit_modal(self, driver, button_text="Create"):
        """Click the button inside the modal and wait for it to close."""
        driver.find_element(
            By.XPATH,
            f"//div[contains(@class,'fixed') and contains(@class,'inset-0')]//button[normalize-space()='{button_text}']",
        ).click()
        self._wait(driver).until(
            EC.invisibility_of_element_located((By.CSS_SELECTOR, ".fixed.inset-0"))
        )

    def _set_react_input(self, driver, element, value):
        """Set a React-controlled input by dispatching native events."""
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
        """Set a React-controlled date input."""
        driver.execute_script(
            "var setter = Object.getOwnPropertyDescriptor("
            "  window.HTMLInputElement.prototype, 'value').set;"
            "setter.call(arguments[0], arguments[1]);"
            "arguments[0].dispatchEvent(new Event('input', {bubbles: true}));"
            "arguments[0].dispatchEvent(new Event('change', {bubbles: true}));",
            element,
            date_str,
        )

    def _select_group(self, driver, group_name):
        """Pick a group by name from the first <select> on the page."""
        select_element = self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, "select"))
        )
        Select(select_element).select_by_visible_text(group_name)
        time.sleep(1)

    @pytest.mark.integration
    def test_expected_expenses_crud(
        self, driver, base_url, auth0_test_user, screenshots_dir
    ):
        """Full expected expenses CRUD workflow:
        1. Log in via Auth0
        2. Create a group and budget
        3. Create an expected expense
        4. Edit the expected expense
        5. Delete the expected expense
        """
        ts = str(int(time.time()))
        group_name = f"EE Test Group {ts}"
        budget_name = f"EE Test Budget {ts}"
        category_name = f"EE Category {ts}"
        expense_name = f"Expected Rent {ts}"
        expense_name_updated = f"Expected Rent {ts} (Edited)"

        # ── 1. Login ───────────────────────────────────────────────────────────
        self._login(driver, base_url, auth0_test_user, screenshots_dir)

        # ── 2. Create group ────────────────────────────────────────────────────
        self._nav(driver, "Groups")
        self._open_modal(driver, "Add Group")
        modal = driver.find_element(By.CSS_SELECTOR, ".fixed.inset-0")
        name_input = modal.find_element(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, name_input, group_name)
        self._submit_modal(driver)
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//td[normalize-space()='{group_name}']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/ee_01_group_created.png")

        # ── 3. Create budget ───────────────────────────────────────────────────
        self._nav(driver, "Budgets")
        self._select_group(driver, group_name)
        self._open_modal(driver, "Add Budget")
        modal = driver.find_element(By.CSS_SELECTOR, ".fixed.inset-0")
        name_input = modal.find_element(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, name_input, budget_name)
        date_inputs = modal.find_elements(By.CSS_SELECTOR, "input[type='date']")
        self._set_react_date(driver, date_inputs[0], "2025-01-01")
        self._set_react_date(driver, date_inputs[1], "2025-12-31")
        self._submit_modal(driver)
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//td[normalize-space()='{budget_name}']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/ee_02_budget_created.png")

        # ── 4. Create category ─────────────────────────────────────────────────
        self._nav(driver, "Categories")
        self._select_group(driver, group_name)
        self._open_modal(driver, "Add Category")
        modal = driver.find_element(By.CSS_SELECTOR, ".fixed.inset-0")
        name_input = modal.find_element(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, name_input, category_name)
        self._submit_modal(driver)
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//h3[normalize-space()='{category_name}']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/ee_03_category_created.png")

        # ── 5. Navigate to Expected Expenses and select group/budget ───────────
        self._nav(driver, "Expected")
        selects = self._wait(driver).until(
            lambda d: d.find_elements(By.CSS_SELECTOR, "select")
        )
        Select(selects[0]).select_by_visible_text(group_name)
        time.sleep(1)
        selects = driver.find_elements(By.CSS_SELECTOR, "select")
        Select(selects[1]).select_by_visible_text(budget_name)
        time.sleep(1)
        driver.save_screenshot(f"{screenshots_dir}/ee_04_selected_group_budget.png")

        # ── 6. Create expected expense with category ───────────────────────────
        self._open_modal(driver, "Add Expected Expense")
        modal = driver.find_element(By.CSS_SELECTOR, ".fixed.inset-0")
        
        text_inputs = modal.find_elements(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, text_inputs[0], expense_name)
        
        amount_input = modal.find_element(By.CSS_SELECTOR, "input[type='number']")
        self._set_react_input(driver, amount_input, "1500.00")
        
        category_button = self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//button[contains(., 'Select category')]")
            )
        )
        driver.execute_script("arguments[0].click()", category_button)
        
        category_search = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.CSS_SELECTOR, "input[placeholder*='Search categories']")
            )
        )
        self._set_react_input(driver, category_search, category_name)
        time.sleep(0.5)
        
        category_option = self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, f"//button[contains(., '{category_name}') and contains(@class,'text-left') and not(contains(.,'Select category'))]")
            )
        )
        category_option.click()
        
        self._wait(driver).until(
            EC.invisibility_of_element_located(
                (By.CSS_SELECTOR, "input[placeholder*='Search categories']")
            )
        )
        time.sleep(0.5)
        
        self._submit_modal(driver)
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//td[normalize-space()='{expense_name}']")
            )
        )
        
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//tr[.//td[normalize-space()='{expense_name}']]//td[contains(., '{category_name}')]")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/ee_05_expense_created_with_category.png")

        # ── 7. Edit expected expense and verify category ───────────────────────
        edit_button = self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, f"//td[normalize-space()='{expense_name}']/ancestor::tr//button[normalize-space()='Edit']")
            )
        )
        edit_button.click()
        
        modal = self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, ".fixed.inset-0"))
        )
        
        category_button = modal.find_element(By.XPATH, f"//button[contains(., '{category_name}')]")
        assert category_button is not None, "Category should be displayed in edit modal"
        
        text_inputs = modal.find_elements(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, text_inputs[0], expense_name_updated)
        
        self._submit_modal(driver, "Update")
        
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//td[normalize-space()='{expense_name_updated}']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/ee_06_expense_edited.png")

        # ── 8. Delete expected expense ─────────────────────────────────────────
        delete_button = self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, f"//td[normalize-space()='{expense_name_updated}']/ancestor::tr//button[normalize-space()='Delete']")
            )
        )
        delete_button.click()
        
        confirm_button = self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//div[contains(@class,'fixed')]//button[normalize-space()='Delete' and contains(@class,'bg-red')]")
            )
        )
        confirm_button.click()
        
        self._wait(driver).until(
            EC.invisibility_of_element_located(
                (By.XPATH, f"//td[normalize-space()='{expense_name_updated}']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/ee_07_expense_deleted.png")

        print("\n✅ Expected expenses CRUD workflow test passed!")
