import pytest
import time
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait, Select
from selenium.webdriver.support import expected_conditions as EC
from selenium.common.exceptions import TimeoutException


class TestThemePreference:
    """End-to-end integration tests for the user theme preference:
    selecting a theme -> PATCH /preferences -> persisted in the database ->
    re-applied on reload without a flash of the previous theme.
    """

    TIMEOUT = 20
    THEME_STORAGE_KEY = "budgets.theme"

    # ── internal helpers ───────────────────────────────────────────────────────

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
                        (By.CSS_SELECTOR,
                         "input[name='username'], input[type='email'], input[id='username']")
                    )
                )
            except TimeoutException:
                driver.save_screenshot(f"{screenshots_dir}/theme_auth0_error.png")
                pytest.skip(
                    "Auth0 login form did not appear — the redirect_uri may not be "
                    "registered in the Auth0 tenant (add the test base URL as an "
                    "Allowed Callback URL in the Auth0 dashboard)."
                )
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

            try:
                self._wait(driver).until(lambda d: "auth0.com" not in d.current_url)
            except TimeoutException:
                driver.save_screenshot(f"{screenshots_dir}/theme_login_stuck.png")
                page_text = driver.find_element(By.TAG_NAME, "body").text[:400]
                pytest.fail(
                    f"Browser did not leave auth0.com after submitting credentials.\n"
                    f"Current URL: {driver.current_url}\n"
                    f"Page text snippet: {page_text}"
                )
            time.sleep(3)

        try:
            WebDriverWait(driver, 30).until(
                EC.presence_of_element_located(
                    (By.XPATH, "//nav//a[contains(normalize-space(),'Groups')]")
                )
            )
        except TimeoutException:
            driver.save_screenshot(f"{screenshots_dir}/theme_post_login_timeout.png")
            page_text = driver.find_element(By.TAG_NAME, "body").text[:500]
            pytest.fail(
                f"Authenticated app layout did not appear after login.\n"
                f"Current URL: {driver.current_url}\n"
                f"Page title: {driver.title}\n"
                f"Page text snippet: {page_text}"
            )
        print("\n  ✓ Logged in")

    def _theme_select(self, driver):
        """Locate the theme dropdown in the top nav.

        Identified by the presence of the LIGHT option value, which distinguishes
        it from the sibling language dropdown.
        """
        element = self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, "//nav//select[option[@value='LIGHT']]")
            )
        )
        return Select(element)

    def _root_classes(self, driver):
        """Class list currently applied to the <html> element."""
        return driver.execute_script("return document.documentElement.className;") or ""

    def _cached_theme(self, driver):
        """Theme value cached in localStorage, or None."""
        return driver.execute_script(
            f"return window.localStorage.getItem('{self.THEME_STORAGE_KEY}');"
        )

    def _select_theme(self, driver, value):
        """Pick a theme and wait until the server round-trip has completed.

        The localStorage cache is only written from server-confirmed query data,
        so waiting for it to hold the new value proves PATCH /preferences
        succeeded and the response was applied.
        """
        # Wait until the dropdown has hydrated (its value matches the cached
        # theme or the server-confirmed theme) so we don't interact with a
        # stale default that swallows the change event.
        self._wait(driver).until(
            lambda d: self._theme_select(d).first_selected_option.get_attribute("value")
            == (self._cached_theme(d) or "LIGHT")
        )

        current = self._theme_select(driver).first_selected_option.get_attribute("value")
        if current == value:
            # The dropdown already shows the target — no change event will
            # fire, so just verify the cache is already consistent.
            if self._cached_theme(driver) != value:
                pytest.fail(
                    f"Dropdown already shows '{value}' but localStorage"
                    f"['{self.THEME_STORAGE_KEY}'] is "
                    f"'{self._cached_theme(driver)}'. The theme was never "
                    f"persisted server-side."
                )
            return

        self._theme_select(driver).select_by_value(value)
        try:
            self._wait(driver).until(
                lambda d: self._cached_theme(d) == value
            )
        except TimeoutException:
            pytest.fail(
                f"Theme '{value}' was never confirmed by the API — "
                f"localStorage['{self.THEME_STORAGE_KEY}'] is "
                f"'{self._cached_theme(driver)}' and <html> class is "
                f"'{self._root_classes(driver)}'. The PATCH /preferences call "
                f"likely failed."
            )

    def _reset_theme(self, driver):
        """Restore LIGHT so the stored preference does not leak into other tests."""
        try:
            self._select_theme(driver, "LIGHT")
        except Exception as exc:
            print(f"\n  ⚠ Could not reset theme to LIGHT: {exc}")

    # ── tests ──────────────────────────────────────────────────────────────────

    def test_theme_dropdown_excludes_dim(
        self, driver, base_url, credentials, screenshots_dir
    ):
        """Only LIGHT and DARK are user-selectable; DIM stays out of the UI."""
        self._login(driver, base_url, credentials, screenshots_dir)

        options = [o.get_attribute("value") for o in self._theme_select(driver).options]
        print(f"\n  Theme options exposed: {options}")

        assert options == ["LIGHT", "DARK"], (
            f"Expected exactly ['LIGHT', 'DARK'] in the theme dropdown, got {options}"
        )
        assert "DIM" not in options, (
            "DIM must not be selectable from the UI even though it remains a "
            "valid theme in the API and stylesheet"
        )
        driver.save_screenshot(f"{screenshots_dir}/theme_01_dropdown_options.png")

    def test_selecting_dark_applies_class_and_caches_it(
        self, driver, base_url, credentials, screenshots_dir
    ):
        """Choosing DARK adds the dark class and caches the confirmed value."""
        self._login(driver, base_url, credentials, screenshots_dir)

        try:
            self._select_theme(driver, "DARK")

            classes = self._root_classes(driver)
            print(f"\n  <html> class after selecting DARK: '{classes}'")

            assert "dark" in classes.split(), (
                f"Expected 'dark' class on <html>, got '{classes}'"
            )
            assert self._cached_theme(driver) == "DARK"
            driver.save_screenshot(f"{screenshots_dir}/theme_02_dark_applied.png")
        finally:
            self._reset_theme(driver)

    def test_theme_persists_across_reload_without_flash(
        self, driver, base_url, credentials, screenshots_dir
    ):
        """After reload the dark class is present immediately, before the API responds."""
        self._login(driver, base_url, credentials, screenshots_dir)

        try:
            self._select_theme(driver, "DARK")

            driver.refresh()

            # Read the class as early as possible: the cached theme is applied
            # synchronously while the preferences module is evaluated, so it must
            # already be set by the time the load event fires.
            classes_on_load = self._root_classes(driver)
            print(f"\n  <html> class immediately after reload: '{classes_on_load}'")

            assert "dark" in classes_on_load.split(), (
                f"Expected 'dark' class to be applied synchronously on reload "
                f"(no theme flash), but <html> class was '{classes_on_load}'. "
                f"The cached theme is probably not being applied before render."
            )

            # And it must still be dark once the app has fully settled.
            self._wait(driver).until(
                EC.presence_of_element_located(
                    (By.XPATH, "//nav//a[contains(normalize-space(),'Groups')]")
                )
            )
            assert "dark" in self._root_classes(driver).split(), (
                "Theme reverted after the app finished loading"
            )
            driver.save_screenshot(f"{screenshots_dir}/theme_03_reload_no_flash.png")
        finally:
            self._reset_theme(driver)

    def test_theme_restored_from_api_when_local_cache_cleared(
        self, driver, base_url, credentials, screenshots_dir
    ):
        """With the local cache dropped, the theme still comes back from the database."""
        self._login(driver, base_url, credentials, screenshots_dir)

        try:
            self._select_theme(driver, "DARK")

            # Drop only the theme cache, keeping the Auth0 tokens intact so the
            # reload stays authenticated and the value can only come from the API.
            driver.execute_script(
                f"window.localStorage.removeItem('{self.THEME_STORAGE_KEY}');"
            )
            assert self._cached_theme(driver) is None

            driver.refresh()
            self._wait(driver).until(
                EC.presence_of_element_located(
                    (By.XPATH, "//nav//a[contains(normalize-space(),'Groups')]")
                )
            )

            try:
                self._wait(driver).until(
                    lambda d: "dark" in self._root_classes(d).split()
                )
            except TimeoutException:
                pytest.fail(
                    f"Theme was not restored from the API after clearing the local "
                    f"cache — <html> class is '{self._root_classes(driver)}'. "
                    f"The preference was probably never persisted server-side."
                )

            # The dropdown must reflect the stored value too.
            assert self._theme_select(driver).first_selected_option.get_attribute(
                "value"
            ) == "DARK"
            assert self._cached_theme(driver) == "DARK", (
                "Theme fetched from the API should be re-cached locally"
            )
            driver.save_screenshot(f"{screenshots_dir}/theme_04_restored_from_api.png")
        finally:
            self._reset_theme(driver)

    def test_switching_back_to_light_removes_dark_class(
        self, driver, base_url, credentials, screenshots_dir
    ):
        """Switching DARK -> LIGHT clears the class and persists the new value."""
        self._login(driver, base_url, credentials, screenshots_dir)

        self._select_theme(driver, "DARK")
        assert "dark" in self._root_classes(driver).split()

        self._select_theme(driver, "LIGHT")

        classes = self._root_classes(driver)
        print(f"\n  <html> class after switching back to LIGHT: '{classes}'")

        assert "dark" not in classes.split(), (
            f"Expected 'dark' class to be removed, got '{classes}'"
        )
        assert "dim" not in classes.split()
        assert self._cached_theme(driver) == "LIGHT"

        driver.refresh()
        self._wait(driver).until(
            EC.presence_of_element_located(
                (By.XPATH, "//nav//a[contains(normalize-space(),'Groups')]")
            )
        )
        assert "dark" not in self._root_classes(driver).split(), (
            "LIGHT theme did not survive a reload"
        )
        driver.save_screenshot(f"{screenshots_dir}/theme_05_back_to_light.png")
