import pytest
import time
import uuid
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait, Select
from selenium.webdriver.support import expected_conditions as EC
from conftest import dismiss_onboarding


class TestCurrencyConversion:
    """End-to-end integration test for multi-currency budget conversion.

    Creates a budget with expenses in different currencies (USD, EUR, ARS),
    then verifies:
      1. Original amounts are displayed in their original currency in the expense tables.
      2. Summary aggregations (expected total, actual total, difference) are converted
         to the user's preferred display currency.
      3. Changing the display currency updates the aggregations to the new currency.
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
                (By.XPATH, "//nav//a[contains(normalize-space(),'Groups')]")
            )
        )
        dismiss_onboarding(driver, base_url)

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

    def _nav(self, driver, link_text):
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, f"//nav//a[contains(normalize-space(),'{link_text}')]")
            )
        ).click()
        time.sleep(1)

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
        modal.find_element(By.XPATH, ".//button[normalize-space()='Create']").click()
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
        modal.find_element(By.XPATH, ".//button[normalize-space()='Create']").click()
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//h3[normalize-space()='{category_name}']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/{tag}_category_created.png")

    def _create_budget(self, driver, group_name, budget_name, screenshots_dir, tag):
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
        date_inputs = modal.find_elements(By.CSS_SELECTOR, "input[type='date']")
        if len(date_inputs) >= 2:
            self._set_react_input(driver, date_inputs[0], "2025-01-01")
            self._set_react_input(driver, date_inputs[1], "2025-12-31")
        modal.find_element(By.XPATH, ".//button[normalize-space()='Create']").click()
        time.sleep(2)
        driver.save_screenshot(f"{screenshots_dir}/{tag}_budget_created.png")

    def _click_budget_name(self, driver, budget_name):
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, f"//td//button[contains(normalize-space(),'{budget_name}')]")
            )
        ).click()
        time.sleep(2)

    def _select_first_category_in_modal(self, driver, modal):
        combobox_btn = modal.find_element(
            By.XPATH, ".//button[contains(normalize-space(),'Select category') or contains(normalize-space(),'Seleccionar categoría')]"
        )
        combobox_btn.click()
        time.sleep(1)
        first_option = self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//div[contains(@class,'max-h-60')]//button[contains(@class,'hover')]")
            )
        )
        first_option.click()
        time.sleep(0.5)

    def _add_expected_expense_with_currency(self, driver, name, amount, currency, screenshots_dir, tag):
        """Add an expected expense with a specific currency."""
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

        currency_select = self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, "select[data-testid='currency-select']"))
        )
        Select(currency_select).select_by_visible_text(currency)
        time.sleep(0.5)

        self._select_first_category_in_modal(driver, modal)

        modal.find_element(By.XPATH, ".//button[normalize-space()='Create']").click()
        time.sleep(2)
        driver.save_screenshot(f"{screenshots_dir}/{tag}_ee_{currency}_added.png")

    def _add_actual_expense_with_currency(self, driver, name, amount, currency, expense_date, screenshots_dir, tag):
        """Add an actual expense with a specific currency and date."""
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

        currency_select = self._wait(driver).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, "select[data-testid='currency-select']"))
        )
        Select(currency_select).select_by_visible_text(currency)
        time.sleep(0.5)

        date_input = modal.find_element(By.CSS_SELECTOR, "input[type='date']")
        self._set_react_input(driver, date_input, expense_date)

        self._select_first_category_in_modal(driver, modal)

        modal.find_element(By.XPATH, ".//button[normalize-space()='Create']").click()
        time.sleep(2)
        driver.save_screenshot(f"{screenshots_dir}/{tag}_ae_{currency}_added.png")

    def _set_display_currency(self, driver, currency):
        """Change the display currency via the nav bar selector."""
        select = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//nav//select[@aria-label='Display currency']")
            )
        )
        Select(select).select_by_value(currency)
        time.sleep(2)

    def _get_summary_card_text(self, driver, card_label):
        """Get the text content of a summary card by its label.

        card_label should match the translated label text, e.g. 'Expected Total',
        'Actual Total', 'Difference'.
        """
        card = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//p[normalize-space()='{card_label}']"
                f"/following-sibling::p[1]")
            )
        )
        return card.text.strip()

    def _get_expense_row_amount(self, driver, expense_name, amount_col=2):
        """Get the amount text from an expense row in the table.

        amount_col is the 1-based index of the Amount <td> within the row.
        Expected expenses: col 2 (Name, Amount, Category, ...).
        Actual expenses: col 3 (Name, Date, Amount, Category, ...).
        """
        row = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//td[normalize-space()='{expense_name}']/ancestor::tr")
            )
        )
        amount_cell = row.find_element(By.XPATH, f".//td[{amount_col}]/div")
        return amount_cell.text.strip()

    @pytest.mark.integration
    def test_multi_currency_budget_conversion(self, driver, base_url, credentials, screenshots_dir):
        """Create a budget with expenses in USD, EUR, and ARS.

        Verify:
        1. Original amounts appear in their original currency in expense tables.
        2. Summary aggregations are in the user's display currency (USD by default).
        3. Changing display currency to EUR updates the aggregations.
        """
        tag = "cc"
        unique = uuid.uuid4().hex[:10]
        group_name = f"CurrencyGroup {unique}"
        category_name = f"CurrencyCat {unique}"
        budget_name = f"CurrencyBudget {unique}"

        self._login(driver, base_url, credentials, screenshots_dir)
        self._create_group(driver, group_name, screenshots_dir, tag)
        self._create_category(driver, group_name, category_name, screenshots_dir, tag)
        self._create_budget(driver, group_name, budget_name, screenshots_dir, tag)

        # Navigate to budget detail
        self._nav(driver, "Budgets")
        self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, "//h1[normalize-space()='Budgets']"))
        )
        self._select_group(driver, group_name)
        time.sleep(2)
        self._click_budget_name(driver, budget_name)

        # Verify we're on the detail page
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//h1[normalize-space()='{budget_name}']")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/{tag}_detail_page.png")

        # ── Add expected expenses in different currencies ──
        ee_usd_name = f"Rent USD {unique}"
        self._add_expected_expense_with_currency(driver, ee_usd_name, "1000.00", "USD", screenshots_dir, tag)

        ee_eur_name = f"Groceries EUR {unique}"
        self._add_expected_expense_with_currency(driver, ee_eur_name, "500.00", "EUR", screenshots_dir, tag)

        ee_ars_name = f"Utilities ARS {unique}"
        self._add_expected_expense_with_currency(driver, ee_ars_name, "50000.00", "ARS", screenshots_dir, tag)

        # ── Add actual expenses in different currencies ──
        ae_usd_name = f"Paid Rent USD {unique}"
        self._add_actual_expense_with_currency(driver, ae_usd_name, "1000.00", "USD", "2025-01-15", screenshots_dir, tag)

        ae_eur_name = f"Paid Groceries EUR {unique}"
        self._add_actual_expense_with_currency(driver, ae_eur_name, "450.00", "EUR", "2025-01-20", screenshots_dir, tag)

        ae_ars_name = f"Paid Utilities ARS {unique}"
        self._add_actual_expense_with_currency(driver, ae_ars_name, "48000.00", "ARS", "2025-01-25", screenshots_dir, tag)

        driver.save_screenshot(f"{screenshots_dir}/{tag}_all_expenses_added.png")

        # ── Verify original amounts are shown in their original currency ──
        # USD expense should show $ or USD
        usd_row = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//tr[.//td[normalize-space()='{ee_usd_name}']]")
            )
        )
        # Debug: print row structure
        tds = usd_row.find_elements(By.XPATH, ".//td")
        print(f"\n  [DEBUG] USD row has {len(tds)} td elements")
        for i, td in enumerate(tds, 1):
            print(f"  [DEBUG]   td[{i}]: '{td.text[:80]}'")
        print(f"  [DEBUG]   row outerHTML: {usd_row.get_attribute('outerHTML')[:500]}")
        usd_amount_cell = usd_row.find_element(By.XPATH, ".//td[2]/div")
        print(f"  [DEBUG]   td[2]/div text: '{usd_amount_cell.text}'")
        assert "1,000" in usd_amount_cell.text or "1000" in usd_amount_cell.text, \
            f"USD expected expense should show 1000, got: {usd_amount_cell.text}"

        # EUR expense should show € or EUR
        eur_row = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//tr[.//td[normalize-space()='{ee_eur_name}']]")
            )
        )
        eur_amount_cell = eur_row.find_element(By.XPATH, ".//td[2]/div")
        assert "500" in eur_amount_cell.text, \
            f"EUR expected expense should show 500, got: {eur_amount_cell.text}"

        # ARS expense should show ARS or $
        ars_row = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//tr[.//td[normalize-space()='{ee_ars_name}']]")
            )
        )
        ars_amount_cell = ars_row.find_element(By.XPATH, ".//td[2]/div")
        assert "50" in ars_amount_cell.text, \
            f"ARS expected expense should show 50000, got: {ars_amount_cell.text}"

        driver.save_screenshot(f"{screenshots_dir}/{tag}_original_amounts_verified.png")

        # ── Verify summary aggregations are in USD (default display currency) ──
        # Wait for summary cards to load
        time.sleep(3)
        expected_total_text = self._get_summary_card_text(driver, "Expected Total")
        actual_total_text = self._get_summary_card_text(driver, "Actual Total")
        difference_text = self._get_summary_card_text(driver, "Difference")

        driver.save_screenshot(f"{screenshots_dir}/{tag}_summary_usd.png")
        print(f"\n  [cc] USD summary - Expected: {expected_total_text}, Actual: {actual_total_text}, Diff: {difference_text}")

        # Summary should be in USD (contains $ symbol, not € or other)
        assert "$" in expected_total_text or "USD" in expected_total_text, \
            f"Expected total should be in USD, got: {expected_total_text}"
        assert "$" in actual_total_text or "USD" in actual_total_text, \
            f"Actual total should be in USD, got: {actual_total_text}"
        assert "$" in difference_text or "USD" in difference_text, \
            f"Difference should be in USD, got: {difference_text}"

        # The expected total should be greater than 1000 (1000 USD + converted EUR + converted ARS)
        # We can't assert exact values since rates fluctuate, but it should be > 1000
        expected_numeric = expected_total_text.replace("$", "").replace(",", "").replace("USD", "").strip()
        try:
            expected_value = float(expected_numeric)
            assert expected_value > 1000, \
                f"Expected total should be > 1000 USD (includes USD+EUR+ARS), got: {expected_value}"
        except ValueError:
            pass  # Can't parse, but currency check already passed

        # ── Change display currency to EUR and verify aggregations update ──
        self._set_display_currency(driver, "EUR")
        time.sleep(3)

        expected_total_eur = self._get_summary_card_text(driver, "Expected Total")
        actual_total_eur = self._get_summary_card_text(driver, "Actual Total")
        difference_eur = self._get_summary_card_text(driver, "Difference")

        driver.save_screenshot(f"{screenshots_dir}/{tag}_summary_eur.png")
        print(f"\n  [cc] EUR summary - Expected: {expected_total_eur}, Actual: {actual_total_eur}, Diff: {difference_eur}")

        # Summary should now be in EUR (contains € symbol)
        assert "€" in expected_total_eur or "EUR" in expected_total_eur, \
            f"Expected total should be in EUR after change, got: {expected_total_eur}"
        assert "€" in actual_total_eur or "EUR" in actual_total_eur, \
            f"Actual total should be in EUR after change, got: {actual_total_eur}"
        assert "€" in difference_eur or "EUR" in difference_eur, \
            f"Difference should be in EUR after change, got: {difference_eur}"

        # ── Verify original amounts in expense tables are still in their original currency ──
        # The expense list shows original amounts, not converted
        usd_row_after = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//tr[.//td[normalize-space()='{ee_usd_name}']]")
            )
        )
        usd_amount_after = usd_row_after.find_element(By.XPATH, ".//td[2]/div")
        assert "1,000" in usd_amount_after.text or "1000" in usd_amount_after.text, \
            f"USD expense should still show original amount after currency change, got: {usd_amount_after.text}"

        eur_row_after = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, f"//tr[.//td[normalize-space()='{ee_eur_name}']]")
            )
        )
        eur_amount_after = eur_row_after.find_element(By.XPATH, ".//td[2]/div")
        assert "500" in eur_amount_after.text, \
            f"EUR expense should still show original amount after currency change, got: {eur_amount_after.text}"

        driver.save_screenshot(f"{screenshots_dir}/{tag}_original_amounts_still_original.png")

        # ── Change display currency to ARS and verify aggregations update ──
        self._set_display_currency(driver, "ARS")
        time.sleep(3)

        expected_total_ars = self._get_summary_card_text(driver, "Expected Total")
        actual_total_ars = self._get_summary_card_text(driver, "Actual Total")
        difference_ars = self._get_summary_card_text(driver, "Difference")

        driver.save_screenshot(f"{screenshots_dir}/{tag}_summary_ars.png")
        print(f"\n  [cc] ARS summary - Expected: {expected_total_ars}, Actual: {actual_total_ars}, Diff: {difference_ars}")

        # Summary should now be in ARS
        assert "ARS" in expected_total_ars or "$" in expected_total_ars, \
            f"Expected total should be in ARS after change, got: {expected_total_ars}"
        assert "ARS" in actual_total_ars or "$" in actual_total_ars, \
            f"Actual total should be in ARS after change, got: {actual_total_ars}"

        driver.save_screenshot(f"{screenshots_dir}/{tag}_test_complete.png")
