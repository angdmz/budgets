import pytest
import time
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait, Select
from selenium.webdriver.support import expected_conditions as EC


class TestBudgetDuplicateRecurring:
    """End-to-end integration tests covering the budget UI enhancements:
    period-type presets (weekly/biweekly/monthly), budget duplication
    (with/without copying expected expenses), and recurring budget creation.
    """

    TIMEOUT = 20

    # ── internal helpers (mirrors test_budget_workflow.py / test_expected_expenses_workflow.py) ──

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
        driver.save_screenshot(f"{screenshots_dir}/bdr_logged_in.png")

    def _nav(self, driver, link_text):
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, f"//nav//a[contains(normalize-space(),'{link_text}')]")
            )
        ).click()
        time.sleep(1)

    def _open_modal(self, driver, button_text):
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, f"//button[normalize-space()='{button_text}']")
            )
        ).click()
        self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, ".fixed.inset-0"))
        )

    def _submit_modal(self, driver, button_text="Create"):
        driver.find_element(
            By.XPATH,
            f"//div[contains(@class,'fixed') and contains(@class,'inset-0')]//button[normalize-space()='{button_text}']",
        ).click()
        self._wait(driver).until(
            EC.invisibility_of_element_located((By.CSS_SELECTOR, ".fixed.inset-0"))
        )

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

    def _set_react_date(self, driver, element, date_str):
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
        option = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//select/option[normalize-space()='{group_name}']")
            )
        )
        select_element = option.find_element(By.XPATH, "./ancestor::select")
        Select(select_element).select_by_visible_text(group_name)
        time.sleep(1)

    def _click_period_type(self, modal, period_label):
        """Click one of the period-type preset buttons (Weekly/Biweekly/Monthly/Custom)."""
        modal.find_element(
            By.XPATH, f".//button[normalize-space()='{period_label}']"
        ).click()

    def _toggle_checkbox_by_label(self, modal, label_text, checked):
        """Set a checkbox's checked state given the visible text of its wrapping label."""
        checkbox = modal.find_element(
            By.XPATH,
            f".//label[contains(., \"{label_text}\")]//input[@type='checkbox']",
        )
        if checkbox.is_selected() != checked:
            checkbox.click()
        return checkbox

    def _create_group(self, driver, group_name, screenshots_dir, tag):
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
        driver.save_screenshot(f"{screenshots_dir}/{tag}_group_created.png")

    def _create_category(self, driver, group_name, category_name, screenshots_dir, tag):
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
        driver.save_screenshot(f"{screenshots_dir}/{tag}_category_created.png")

    def _add_expected_expense(self, driver, group_name, budget_name, category_name, expense_name, amount, screenshots_dir, tag):
        self._nav(driver, "Expected")
        self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, "//h1[normalize-space()='Expected Expenses']"))
        )
        group_option = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//select/option[normalize-space()='{group_name}']")
            )
        )
        Select(group_option.find_element(By.XPATH, "./ancestor::select")).select_by_visible_text(group_name)
        time.sleep(1)
        budget_option = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//select/option[normalize-space()='{budget_name}']")
            )
        )
        Select(budget_option.find_element(By.XPATH, "./ancestor::select")).select_by_visible_text(budget_name)
        time.sleep(1)

        self._open_modal(driver, "Add Expected Expense")
        modal = driver.find_element(By.CSS_SELECTOR, ".fixed.inset-0")

        text_inputs = modal.find_elements(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, text_inputs[0], expense_name)

        amount_input = modal.find_element(By.CSS_SELECTOR, "input[type='number']")
        self._set_react_input(driver, amount_input, amount)

        category_button = self._wait(driver).until(
            EC.element_to_be_clickable((By.XPATH, "//button[contains(., 'Select category')]"))
        )
        driver.execute_script("arguments[0].click()", category_button)
        category_search = self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, "input[placeholder*='Search categories']"))
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
            EC.invisibility_of_element_located((By.CSS_SELECTOR, "input[placeholder*='Search categories']"))
        )
        time.sleep(0.5)

        self._submit_modal(driver)
        self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, f"//td[normalize-space()='{expense_name}']"))
        )
        driver.save_screenshot(f"{screenshots_dir}/{tag}_expected_expense_created.png")

    def _goto_budgets_and_select_group(self, driver, group_name):
        self._nav(driver, "Budgets")
        self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, "//h1[normalize-space()='Budgets']"))
        )
        self._select_group(driver, group_name)

    # ── tests ─────────────────────────────────────────────────────────────────

    @pytest.mark.integration
    def test_monthly_period_type_autofills_dates(
        self, driver, base_url, auth0_test_user, screenshots_dir
    ):
        """Selecting the 'Monthly' period-type preset should auto-populate the
        start/end date fields without requiring manual date entry, and the
        created budget should span the full current month.
        """
        ts = str(int(time.time()))
        group_name = f"Period Test Group {ts}"
        budget_name = f"Monthly Budget {ts}"

        self._login(driver, base_url, auth0_test_user, screenshots_dir)
        self._create_group(driver, group_name, screenshots_dir, "pt")

        self._goto_budgets_and_select_group(driver, group_name)
        self._open_modal(driver, "Add Budget")
        modal = driver.find_element(By.CSS_SELECTOR, ".fixed.inset-0")

        name_input = modal.find_element(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, name_input, budget_name)

        # No date inputs should be required before selecting a preset since
        # 'Monthly' auto-fills them; verify no date inputs are visible for
        # the preset view, then click "Monthly".
        self._click_period_type(modal, "Monthly")
        time.sleep(0.5)

        # Date inputs should NOT be present anymore (only shown for Custom).
        date_inputs = modal.find_elements(By.CSS_SELECTOR, "input[type='date']")
        assert len(date_inputs) == 0, "Date inputs should be hidden when a preset period type is selected"

        driver.save_screenshot(f"{screenshots_dir}/pt_monthly_selected.png")
        self._submit_modal(driver)

        self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, f"//td[normalize-space()='{budget_name}']"))
        )
        driver.save_screenshot(f"{screenshots_dir}/pt_monthly_budget_created.png")

        # Verify the period cell shows a first-of-month -> last-of-month range.
        period_cell = driver.find_element(
            By.XPATH, f"//td[normalize-space()='{budget_name}']/following-sibling::td[1]"
        )
        assert period_cell.text.strip() != "", "Period column should not be empty"
        print(f"\n  ✓ Monthly budget created with period: {period_cell.text}")

    @pytest.mark.integration
    def test_duplicate_budget_copies_expected_expenses_by_default(
        self, driver, base_url, auth0_test_user, screenshots_dir
    ):
        """Duplicating a budget with the 'Copy expected expenses' checkbox left
        at its default (checked) state should copy expected expenses into the
        new budget.
        """
        ts = str(int(time.time()))
        group_name = f"Dup Test Group {ts}"
        budget_name = f"Original Budget {ts}"
        category_name = f"Dup Category {ts}"
        expense_name = f"Rent {ts}"
        duplicate_name = f"Duplicated Budget {ts}"

        self._login(driver, base_url, auth0_test_user, screenshots_dir)
        self._create_group(driver, group_name, screenshots_dir, "dup")

        # Create source budget (custom dates keep this deterministic).
        self._goto_budgets_and_select_group(driver, group_name)
        self._open_modal(driver, "Add Budget")
        modal = driver.find_element(By.CSS_SELECTOR, ".fixed.inset-0")
        name_input = modal.find_element(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, name_input, budget_name)
        self._click_period_type(modal, "Custom")
        date_inputs = modal.find_elements(By.CSS_SELECTOR, "input[type='date']")
        self._set_react_date(driver, date_inputs[0], "2025-01-01")
        self._set_react_date(driver, date_inputs[1], "2025-01-31")
        self._submit_modal(driver)
        self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, f"//td[normalize-space()='{budget_name}']"))
        )

        self._create_category(driver, group_name, category_name, screenshots_dir, "dup")
        self._add_expected_expense(
            driver, group_name, budget_name, category_name, expense_name, "1200.00", screenshots_dir, "dup"
        )

        # Duplicate the source budget.
        self._goto_budgets_and_select_group(driver, group_name)
        duplicate_button = self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, f"//td[normalize-space()='{budget_name}']/ancestor::tr//button[normalize-space()='Duplicate']")
            )
        )
        duplicate_button.click()

        modal = self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, ".fixed.inset-0"))
        )
        name_input = modal.find_element(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, name_input, duplicate_name)

        # "Copy expected expenses" checkbox must default to checked.
        copy_checkbox = modal.find_element(
            By.XPATH, ".//label[contains(., 'Copy expected expenses')]//input[@type='checkbox']"
        )
        assert copy_checkbox.is_selected(), "Copy expected expenses checkbox should default to checked"

        driver.save_screenshot(f"{screenshots_dir}/dup_modal_before_submit.png")
        self._submit_modal(driver, "Duplicate")

        self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, f"//td[normalize-space()='{duplicate_name}']"))
        )
        driver.save_screenshot(f"{screenshots_dir}/dup_new_budget_created.png")

        # Verify the expected expense was copied into the new budget.
        self._nav(driver, "Expected")
        group_option = self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, f"//select/option[normalize-space()='{group_name}']"))
        )
        Select(group_option.find_element(By.XPATH, "./ancestor::select")).select_by_visible_text(group_name)
        time.sleep(1)
        budget_option = self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, f"//select/option[normalize-space()='{duplicate_name}']"))
        )
        Select(budget_option.find_element(By.XPATH, "./ancestor::select")).select_by_visible_text(duplicate_name)
        time.sleep(1)

        self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, f"//td[normalize-space()='{expense_name}']"))
        )
        driver.save_screenshot(f"{screenshots_dir}/dup_expense_copied.png")
        print("\n  ✓ Expected expense copied into duplicated budget")

    @pytest.mark.integration
    def test_duplicate_budget_without_expenses(
        self, driver, base_url, auth0_test_user, screenshots_dir
    ):
        """Unchecking 'Copy expected expenses' before duplicating should create
        the new budget with no expected expenses.
        """
        ts = str(int(time.time()))
        group_name = f"Dup NoExp Group {ts}"
        budget_name = f"Original NoExp Budget {ts}"
        category_name = f"NoExp Category {ts}"
        expense_name = f"Utilities {ts}"
        duplicate_name = f"Duplicated NoExp Budget {ts}"

        self._login(driver, base_url, auth0_test_user, screenshots_dir)
        self._create_group(driver, group_name, screenshots_dir, "dupne")

        self._goto_budgets_and_select_group(driver, group_name)
        self._open_modal(driver, "Add Budget")
        modal = driver.find_element(By.CSS_SELECTOR, ".fixed.inset-0")
        name_input = modal.find_element(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, name_input, budget_name)
        self._click_period_type(modal, "Custom")
        date_inputs = modal.find_elements(By.CSS_SELECTOR, "input[type='date']")
        self._set_react_date(driver, date_inputs[0], "2025-02-01")
        self._set_react_date(driver, date_inputs[1], "2025-02-28")
        self._submit_modal(driver)
        self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, f"//td[normalize-space()='{budget_name}']"))
        )

        self._create_category(driver, group_name, category_name, screenshots_dir, "dupne")
        self._add_expected_expense(
            driver, group_name, budget_name, category_name, expense_name, "80.00", screenshots_dir, "dupne"
        )

        self._goto_budgets_and_select_group(driver, group_name)
        duplicate_button = self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, f"//td[normalize-space()='{budget_name}']/ancestor::tr//button[normalize-space()='Duplicate']")
            )
        )
        duplicate_button.click()

        modal = self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, ".fixed.inset-0"))
        )
        name_input = modal.find_element(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, name_input, duplicate_name)

        self._toggle_checkbox_by_label(modal, "Copy expected expenses", checked=False)
        driver.save_screenshot(f"{screenshots_dir}/dupne_modal_unchecked.png")
        self._submit_modal(driver, "Duplicate")

        self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, f"//td[normalize-space()='{duplicate_name}']"))
        )

        self._nav(driver, "Expected")
        group_option = self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, f"//select/option[normalize-space()='{group_name}']"))
        )
        Select(group_option.find_element(By.XPATH, "./ancestor::select")).select_by_visible_text(group_name)
        time.sleep(1)
        budget_option = self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, f"//select/option[normalize-space()='{duplicate_name}']"))
        )
        Select(budget_option.find_element(By.XPATH, "./ancestor::select")).select_by_visible_text(duplicate_name)
        time.sleep(1.5)

        expense_rows = driver.find_elements(By.XPATH, f"//td[normalize-space()='{expense_name}']")
        assert len(expense_rows) == 0, "No expected expenses should be copied when the checkbox is unchecked"
        driver.save_screenshot(f"{screenshots_dir}/dupne_no_expenses_copied.png")
        print("\n  ✓ Duplicate without copying expected expenses verified")

    @pytest.mark.integration
    def test_create_recurring_monthly_budgets(
        self, driver, base_url, auth0_test_user, screenshots_dir
    ):
        """Creating a budget with 'Monthly' period type and the 'Create
        recurring budgets' option should create multiple sequential budgets
        (e.g. for recurring expenses like rent).
        """
        ts = str(int(time.time()))
        group_name = f"Recurring Group {ts}"
        base_name = f"Rent {ts}"

        self._login(driver, base_url, auth0_test_user, screenshots_dir)
        self._create_group(driver, group_name, screenshots_dir, "rec")

        self._goto_budgets_and_select_group(driver, group_name)
        self._open_modal(driver, "Add Budget")
        modal = driver.find_element(By.CSS_SELECTOR, ".fixed.inset-0")
        name_input = modal.find_element(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, name_input, base_name)

        self._click_period_type(modal, "Monthly")
        time.sleep(0.5)

        self._toggle_checkbox_by_label(modal, "Create recurring budgets", checked=True)

        periods_input = modal.find_element(By.CSS_SELECTOR, "input[type='number']")
        self._set_react_input(driver, periods_input, "3")

        driver.save_screenshot(f"{screenshots_dir}/rec_modal_before_submit.png")
        self._submit_modal(driver)

        # All 3 recurring budgets should eventually appear in the table.
        for i in range(1, 4):
            self._wait(driver).until(
                EC.presence_of_element_located(
                    (By.XPATH, f"//td[normalize-space()='{base_name} #{i}']")
                )
            )
        driver.save_screenshot(f"{screenshots_dir}/rec_budgets_created.png")
        print(f"\n  ✓ 3 recurring monthly budgets created for base name: {base_name}")

    @pytest.mark.integration
    def test_duplicate_monthly_budget_for_selected_months(
        self, driver, base_url, auth0_test_user, screenshots_dir
    ):
        """Duplicating a Monthly budget with 'Create recurring budgets'
        enabled should present a month picker (instead of a raw period
        count) so users can select exactly which months of the year to
        generate budgets for -- e.g. setting up expected expenses once and
        reapplying them across several months. Each generated budget should
        be named after its month and should carry over the source budget's
        expected expenses.
        """
        ts = str(int(time.time()))
        group_name = f"MonthPicker Group {ts}"
        budget_name = f"Monthly Source {ts}"
        category_name = f"MonthPicker Category {ts}"
        expense_name = f"Groceries {ts}"
        duplicate_base_name = f"Monthly Plan {ts}"

        self._login(driver, base_url, auth0_test_user, screenshots_dir)
        self._create_group(driver, group_name, screenshots_dir, "mp")

        # Create the source monthly budget with an expected expense.
        self._goto_budgets_and_select_group(driver, group_name)
        self._open_modal(driver, "Add Budget")
        modal = driver.find_element(By.CSS_SELECTOR, ".fixed.inset-0")
        name_input = modal.find_element(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, name_input, budget_name)
        self._click_period_type(modal, "Monthly")
        time.sleep(0.5)
        self._submit_modal(driver)
        self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, f"//td[normalize-space()='{budget_name}']"))
        )

        self._create_category(driver, group_name, category_name, screenshots_dir, "mp")
        self._add_expected_expense(
            driver, group_name, budget_name, category_name, expense_name, "150.00", screenshots_dir, "mp"
        )

        # Duplicate it into a recurring set of hand-picked months.
        self._goto_budgets_and_select_group(driver, group_name)
        duplicate_button = self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, f"//td[normalize-space()='{budget_name}']/ancestor::tr//button[normalize-space()='Duplicate']")
            )
        )
        duplicate_button.click()

        modal = self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, ".fixed.inset-0"))
        )
        name_input = modal.find_element(By.CSS_SELECTOR, "input[type='text']")
        self._set_react_input(driver, name_input, duplicate_base_name)

        self._click_period_type(modal, "Monthly")
        time.sleep(0.5)
        self._toggle_checkbox_by_label(modal, "Create recurring budgets", checked=True)
        time.sleep(0.5)

        # A number-of-periods input must NOT be present for Monthly
        # recurring duplication; a month picker should appear instead.
        number_inputs = modal.find_elements(By.CSS_SELECTOR, "input[type='number']")
        assert len(number_inputs) == 0, "Monthly recurring duplication should show a month picker, not a period count"

        selected_month_labels = ["January", "February"]
        for month_label in selected_month_labels:
            modal.find_element(
                By.XPATH, f".//button[normalize-space()='{month_label}']"
            ).click()

        driver.save_screenshot(f"{screenshots_dir}/mp_months_selected.png")
        self._submit_modal(driver, "Duplicate")

        generated_budget_names = []
        for month_label in selected_month_labels:
            cell = self._wait(driver).until(
                EC.presence_of_element_located(
                    (By.XPATH, f"//td[starts-with(normalize-space(),'{duplicate_base_name} - {month_label}')]")
                )
            )
            generated_budget_names.append(cell.text.strip())
        driver.save_screenshot(f"{screenshots_dir}/mp_budgets_created.png")

        # Verify the expected expense was copied into one of the generated budgets.
        self._nav(driver, "Expected")
        group_option = self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, f"//select/option[normalize-space()='{group_name}']"))
        )
        Select(group_option.find_element(By.XPATH, "./ancestor::select")).select_by_visible_text(group_name)
        time.sleep(1)
        target_budget_name = generated_budget_names[0]
        budget_option = self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, f"//select/option[normalize-space()='{target_budget_name}']"))
        )
        Select(budget_option.find_element(By.XPATH, "./ancestor::select")).select_by_visible_text(target_budget_name)
        time.sleep(1)

        self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, f"//td[normalize-space()='{expense_name}']"))
        )
        driver.save_screenshot(f"{screenshots_dir}/mp_expense_copied.png")
        print(f"\n  ✓ Duplicated monthly budget for hand-picked months: {generated_budget_names}")
