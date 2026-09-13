import pytest
import time
import uuid
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait, Select
from selenium.webdriver.support import expected_conditions as EC
from conftest import dismiss_onboarding


class TestBudgetPlanDraft:
    """End-to-end integration tests for localStorage draft persistence and async creation.

    Covers:
    - Draft saved to localStorage and resumed after page reload
    - Draft discarded
    - Async creation navigates immediately and budgets appear
    """

    TIMEOUT = 20

    # ── helpers (mirrors test_budget_plan_wizard.py) ──

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
        driver.save_screenshot(f"{screenshots_dir}/draft_logged_in.png")
        dismiss_onboarding(driver, base_url)

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

    def _fill_step1(self, driver, group_name, base_budget_name, cadence_label="Monthly"):
        """Fill step 1 of the wizard without clicking Next."""
        self._select_wizard_group(driver, group_name)

        name_input = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//input[@placeholder and contains(@placeholder, 'Monthly Budget')]")
            )
        )
        self._set_react_input(driver, name_input, base_budget_name)

        self._click_cadence(driver, cadence_label)

    # ── tests ──

    @pytest.mark.order(1)
    def test_draft_resume_after_reload(
        self, driver, base_url, credentials, screenshots_dir
    ):
        """Fill step 1, reload page, verify draft resume prompt appears, resume and create."""
        tag = "draft_resume"
        unique = uuid.uuid4().hex[:10]
        group_name = f"DraftGroup {unique}"
        base_budget_name = f"DraftBudget {unique}"

        self._login(driver, base_url, credentials, screenshots_dir)
        self._create_group(driver, group_name, screenshots_dir, tag)

        # Go to wizard and fill step 1
        self._goto_budgets_and_select_group(driver, group_name)
        self._click_create_plan(driver)
        self._fill_step1(driver, group_name, base_budget_name)
        time.sleep(1)
        driver.save_screenshot(f"{screenshots_dir}/{tag}_step1_filled.png")

        # Reload the page
        driver.get(driver.current_url)
        time.sleep(3)

        # Verify resume prompt appears
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//p[contains(normalize-space(),'saved draft was found')]")
            )
        )
        driver.save_screenshot(f"{screenshots_dir}/{tag}_resume_prompt.png")

        # Click "Resume draft"
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//button[normalize-space()='Resume draft']")
            )
        ).click()
        time.sleep(1)

        # Verify the name field is restored
        name_input = driver.find_element(
            By.XPATH, "//input[@placeholder and contains(@placeholder, 'Monthly Budget')]"
        )
        assert name_input.get_attribute("value") == base_budget_name, (
            f"Expected name field to be '{base_budget_name}' after resume, "
            f"got '{name_input.get_attribute('value')}'"
        )
        driver.save_screenshot(f"{screenshots_dir}/{tag}_resumed.png")

        # Proceed to step 3 and create
        self._click_next(driver)  # step 1 → 2
        self._click_next(driver)  # step 2 → 3

        # Click Create Plan
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//button[normalize-space()='Create Plan']")
            )
        ).click()

        # Verify we navigate to /budgets (async — should be fast)
        self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, "//h1[normalize-space()='Budgets']"))
        )
        time.sleep(5)

        # Re-select group and verify budget was created
        self._select_group(driver, group_name)
        time.sleep(3)
        driver.save_screenshot(f"{screenshots_dir}/{tag}_created.png")

        budget_cells = driver.find_elements(
            By.XPATH, f"//td//button[contains(normalize-space(), '{base_budget_name}')]"
        )
        assert len(budget_cells) >= 1, (
            f"Expected at least 1 budget with name containing '{base_budget_name}', "
            f"found {len(budget_cells)}"
        )

    @pytest.mark.order(2)
    def test_draft_discard(
        self, driver, base_url, credentials, screenshots_dir
    ):
        """Fill step 1, reload, discard draft, verify fields are empty."""
        tag = "draft_discard"
        unique = uuid.uuid4().hex[:10]
        group_name = f"DiscardGroup {unique}"
        base_budget_name = f"DiscardBudget {unique}"

        self._login(driver, base_url, credentials, screenshots_dir)
        self._create_group(driver, group_name, screenshots_dir, tag)

        # Go to wizard and fill step 1
        self._goto_budgets_and_select_group(driver, group_name)
        self._click_create_plan(driver)
        self._fill_step1(driver, group_name, base_budget_name)
        time.sleep(1)

        # Reload
        driver.get(driver.current_url)
        time.sleep(3)

        # Verify resume prompt
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//p[contains(normalize-space(),'saved draft was found')]")
            )
        )

        # Click "Discard draft"
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//button[normalize-space()='Discard draft']")
            )
        ).click()
        time.sleep(1)

        # Verify the prompt is gone
        prompts = driver.find_elements(
            By.XPATH, "//p[contains(normalize-space(),'saved draft was found')]"
        )
        assert len(prompts) == 0, "Draft prompt should be gone after discarding"

        # Verify name field is empty
        name_input = driver.find_element(
            By.XPATH, "//input[@placeholder and contains(@placeholder, 'Monthly Budget')]"
        )
        assert name_input.get_attribute("value") == "", (
            f"Expected empty name field after discard, got '{name_input.get_attribute('value')}'"
        )
        driver.save_screenshot(f"{screenshots_dir}/{tag}_discarded.png")

    @pytest.mark.order(3)
    def test_async_creation_navigates_immediately(
        self, driver, base_url, credentials, screenshots_dir
    ):
        """Create a plan and verify navigation to /budgets happens immediately (async)."""
        tag = "async"
        unique = uuid.uuid4().hex[:10]
        group_name = f"AsyncGroup {unique}"
        base_budget_name = f"AsyncBudget {unique}"

        self._login(driver, base_url, credentials, screenshots_dir)
        self._create_group(driver, group_name, screenshots_dir, tag)

        # Go to wizard
        self._goto_budgets_and_select_group(driver, group_name)
        self._click_create_plan(driver)
        self._fill_step1(driver, group_name, base_budget_name, cadence_label="Weekly")

        # Set 2 periods
        period_input = self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//input[@type='number' and @min='1']")
            )
        )
        self._set_react_input(driver, period_input, "2")

        self._click_next(driver)  # step 1 → 2
        self._click_next(driver)  # step 2 → 3

        # Click Create Plan and measure time to navigate
        create_btn = self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//button[normalize-space()='Create Plan']")
            )
        )
        create_btn.click()

        # Should navigate to /budgets quickly (within 5 seconds — async)
        start = time.time()
        self._wait(driver).until(
            EC.presence_of_element_located((By.XPATH, "//h1[normalize-space()='Budgets']"))
        )
        elapsed = time.time() - start
        assert elapsed < 10, f"Navigation took {elapsed:.1f}s, expected < 10s for async"

        # Wait for creation to complete in background
        time.sleep(8)

        # Re-select group and verify budgets
        self._select_group(driver, group_name)
        time.sleep(3)
        driver.save_screenshot(f"{screenshots_dir}/{tag}_created.png")

        budget_cells = driver.find_elements(
            By.XPATH, f"//td//button[contains(normalize-space(), '{base_budget_name}')]"
        )
        assert len(budget_cells) == 2, (
            f"Expected 2 budgets with name containing '{base_budget_name}', "
            f"found {len(budget_cells)}"
        )
