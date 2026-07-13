import pytest
import time
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait, Select
from selenium.webdriver.support import expected_conditions as EC


class TestEditDeleteWorkflow:
    """End-to-end integration test covering edit and delete operations:
    login → create entities → edit each → verify updates → delete each → verify removal.
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
        driver.save_screenshot(f"{screenshots_dir}/edit_delete_logged_in.png")

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
    def test_edit_and_delete_workflow(
        self, driver, base_url, auth0_test_user, screenshots_dir
    ):
        """Full edit/delete workflow:
        1. Log in via Auth0
        2. Create a group, budget, category, and expense
        3. Edit each entity and verify the updated name
        4. Delete each entity and verify removal from UI
        """
        ts = str(int(time.time()))
        group_name = f"Test Group {ts}"
        budget_name = f"Test Budget {ts}"
        category_name = f"Test Category {ts}"
        expense_name = f"Test Expense {ts}"

        group_name_updated = f"{group_name} (Edited)"
        budget_name_updated = f"{budget_name} (Edited)"
        category_name_updated = f"{category_name} (Edited)"
        expense_name_updated = f"{expense_name} (Edited)"

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
        driver.save_screenshot(f"{screenshots_dir}/edit_delete_01_group_created.png")

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
        driver.save_screenshot(f"{screenshots_dir}/edit_delete_02_budget_created.png")

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
        driver.save_screenshot(f"{screenshots_dir}/edit_delete_03_category_created.png")

        # ── 5. Create expense ──────────────────────────────────────────────────
        self._nav(driver, "Expenses")
        selects = self._wait(driver).until(
            lambda d: d.find_elements(By.CSS_SELECTOR, "select")
        )
        Select(selects[0]).select_by_visible_text(group_name)
        time.sleep(1)
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
        self._submit_modal(driver)
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//td[normalize-space()='{expense_name}']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/edit_delete_04_expense_created.png")

        # ── 6. Edit category ───────────────────────────────────────────────────
        self._nav(driver, "Categories")
        self._select_group(driver, group_name)
        time.sleep(1)
        driver.save_screenshot(f"{screenshots_dir}/edit_delete_05_before_category_edit.png")
        
        # Click Edit button next to the category
        edit_button = self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, f"//h3[normalize-space()='{category_name}']/../../..//button[normalize-space()='Edit']")
            )
        )
        edit_button.click()
        
        # Wait for edit modal and update name
        modal = self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, ".fixed.inset-0"))
        )
        name_input = modal.find_element(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, name_input, category_name_updated)
        self._submit_modal(driver, "Update")
        
        # Verify updated name appears
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//h3[normalize-space()='{category_name_updated}']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/edit_delete_05_category_edited.png")

        # ── 7. Edit budget ─────────────────────────────────────────────────────
        self._nav(driver, "Budgets")
        self._select_group(driver, group_name)
        time.sleep(1)
        
        # Click Edit button in the budget row
        edit_button = self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, f"//td[normalize-space()='{budget_name}']/ancestor::tr//button[normalize-space()='Edit']")
            )
        )
        edit_button.click()
        
        # Wait for edit modal and update name
        modal = self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, ".fixed.inset-0"))
        )
        name_input = modal.find_element(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, name_input, budget_name_updated)
        self._submit_modal(driver, "Update")
        
        # Verify updated name appears
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//td[normalize-space()='{budget_name_updated}']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/edit_delete_06_budget_edited.png")

        # ── 8. Edit expense ────────────────────────────────────────────────────
        self._nav(driver, "Expenses")
        selects = driver.find_elements(By.CSS_SELECTOR, "select")
        Select(selects[0]).select_by_visible_text(group_name)
        time.sleep(1)
        selects = driver.find_elements(By.CSS_SELECTOR, "select")
        Select(selects[1]).select_by_visible_text(budget_name_updated)
        time.sleep(1)
        
        # Click Edit button in the expense row
        edit_button = self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, f"//td[normalize-space()='{expense_name}']/ancestor::tr//button[normalize-space()='Edit']")
            )
        )
        edit_button.click()
        
        # Wait for edit modal and update name
        modal = self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, ".fixed.inset-0"))
        )
        name_input = modal.find_element(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, name_input, expense_name_updated)
        self._submit_modal(driver, "Update")
        
        # Verify updated name appears
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//td[normalize-space()='{expense_name_updated}']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/edit_delete_07_expense_edited.png")

        # ── 9. Delete expense ──────────────────────────────────────────────────
        delete_button = self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, f"//td[normalize-space()='{expense_name_updated}']/ancestor::tr//button[normalize-space()='Delete']")
            )
        )
        delete_button.click()
        
        # Confirm deletion in modal
        confirm_button = self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//div[contains(@class,'fixed')]//button[normalize-space()='Delete' and contains(@class,'bg-red')]")
            )
        )
        confirm_button.click()
        
        # Verify expense is removed
        self._wait(driver).until(
            EC.invisibility_of_element_located(
                (By.XPATH, f"//td[normalize-space()='{expense_name_updated}']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/edit_delete_08_expense_deleted.png")

        # ── 10. Delete category ────────────────────────────────────────────────
        self._nav(driver, "Categories")
        self._select_group(driver, group_name)
        time.sleep(1)
        
        delete_button = self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, f"//h3[normalize-space()='{category_name_updated}']/../../..//button[normalize-space()='Delete']")
            )
        )
        delete_button.click()
        
        # Confirm deletion
        confirm_button = self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//div[contains(@class,'fixed')]//button[normalize-space()='Delete' and contains(@class,'bg-red')]")
            )
        )
        confirm_button.click()
        
        # Verify category is removed
        self._wait(driver).until(
            EC.invisibility_of_element_located(
                (By.XPATH, f"//h3[normalize-space()='{category_name_updated}']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/edit_delete_09_category_deleted.png")

        # ── 11. Delete budget ──────────────────────────────────────────────────
        self._nav(driver, "Budgets")
        self._select_group(driver, group_name)
        time.sleep(1)
        
        delete_button = self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, f"//td[normalize-space()='{budget_name_updated}']/ancestor::tr//button[normalize-space()='Delete']")
            )
        )
        delete_button.click()
        
        # Confirm deletion
        confirm_button = self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//div[contains(@class,'fixed')]//button[normalize-space()='Delete' and contains(@class,'bg-red')]")
            )
        )
        confirm_button.click()
        
        # Verify budget is removed
        self._wait(driver).until(
            EC.invisibility_of_element_located(
                (By.XPATH, f"//td[normalize-space()='{budget_name_updated}']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/edit_delete_10_budget_deleted.png")

        print("\n✅ Full edit/delete workflow test passed!")
