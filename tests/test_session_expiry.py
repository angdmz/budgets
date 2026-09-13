import time

import pytest
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.common.exceptions import TimeoutException
from conftest import dismiss_onboarding

from settings import settings


class TestSessionExpiry:
    """Integration tests for the "session expired" flow: when a call to the
    API fails because the access token is invalid/expired, or because the
    Auth0 SDK cannot silently refresh it, the app must show a session-expired
    dialog and log the user out.
    """

    TIMEOUT = 20

    # ── internal helpers ───────────────────────────────────────────────────────

    def _wait(self, driver):
        return WebDriverWait(driver, self.TIMEOUT)

    def _login(self, driver, base_url, credentials, screenshots_dir):
        """Navigate to /app and complete Auth0 Universal Login if redirected."""
        driver.get(f"{base_url}/app")
        time.sleep(2)
        driver.save_screenshot(f"{screenshots_dir}/session_01_pre_login.png")

        if "auth0.com" in driver.current_url or "auth0" in driver.page_source.lower():
            try:
                email_field = WebDriverWait(driver, 10).until(
                    EC.presence_of_element_located(
                        (By.CSS_SELECTOR,
                         "input[name='username'], input[type='email'], input[id='username']")
                    )
                )
            except TimeoutException:
                driver.save_screenshot(f"{screenshots_dir}/session_auth0_error.png")
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
                driver.save_screenshot(f"{screenshots_dir}/session_login_stuck.png")
                pytest.fail("Browser did not leave auth0.com after submitting credentials.")
            time.sleep(3)

        try:
            WebDriverWait(driver, 30).until(
                EC.presence_of_element_located(
                    (By.XPATH, "//nav//a[contains(normalize-space(),'Groups')]")
                )
            )
        except TimeoutException:
            driver.save_screenshot(f"{screenshots_dir}/session_post_login_timeout.png")
            pytest.fail("Authenticated app layout did not appear after login.")
        driver.save_screenshot(f"{screenshots_dir}/session_02_logged_in.png")
        print("\n  ✓ Logged in")
        dismiss_onboarding(driver, base_url)

    def _nav(self, driver, link_text):
        """Click a top-nav link by its visible text and wait for the page to settle."""
        self._wait(driver).until(
            EC.element_to_be_clickable(
                (By.XPATH, f"//nav//a[contains(normalize-space(),'{link_text}')]")
            )
        ).click()
        time.sleep(1)

    def _auth0_cache_key(self, driver):
        """Locate the auth0-spa-js token cache entry in localStorage.

        The SDK (with cacheLocation="localstorage") stores the token set under
        a key like ``@@auth0spajs@@::{client_id}::{audience}::{scope}``. We only
        know the client_id/audience, so match by prefix.
        """
        client_id = settings.auth0_client_id
        audience = settings.auth0_audience
        prefix = f"@@auth0spajs@@::{client_id}::{audience}"
        key = driver.execute_script(
            "const prefix = arguments[0];"
            "return Object.keys(window.localStorage)"
            "  .find(k => k.startsWith(prefix));",
            prefix,
        )
        if not key:
            pytest.fail(
                f"Could not find Auth0 token cache entry in localStorage with "
                f"prefix '{prefix}'. Keys present: "
                f"{driver.execute_script('return Object.keys(window.localStorage);')}"
            )
        return key

    def _get_auth0_cache_entry(self, driver, key):
        return driver.execute_script(
            "return JSON.parse(window.localStorage.getItem(arguments[0]));", key
        )

    def _set_auth0_cache_entry(self, driver, key, entry):
        driver.execute_script(
            "window.localStorage.setItem(arguments[0], JSON.stringify(arguments[1]));",
            key,
            entry,
        )

    def _corrupt_access_token(self, driver):
        """Overwrite the cached access token with a bogus value while keeping
        the cache entry "fresh" (expiresAt in the future) so the Auth0 SDK
        returns it as-is without attempting a silent refresh. The next API
        call will then be rejected by the backend with 401.
        """
        key = self._auth0_cache_key(driver)
        entry = self._get_auth0_cache_entry(driver, key)
        entry["body"]["access_token"] = "expired.invalid.token"
        self._set_auth0_cache_entry(driver, key, entry)

    def _expire_session_without_refresh_token(self, driver):
        """Simulate a fully-expired session: the cached token is expired and
        there is no refresh token to silently obtain a new one, so
        getAccessTokenSilently rejects with a login/refresh-token error.
        """
        key = self._auth0_cache_key(driver)
        entry = self._get_auth0_cache_entry(driver, key)
        entry["body"].pop("refresh_token", None)
        entry["expiresAt"] = int(time.time()) - 3600
        self._set_auth0_cache_entry(driver, key, entry)

    def _dialog(self, driver):
        return self._wait(driver).until(
            EC.visibility_of_element_located(
                (By.CSS_SELECTOR, "[data-testid='session-expired-dialog']")
            )
        )

    # ── tests ──────────────────────────────────────────────────────────────────

    def test_expired_access_token_shows_session_dialog(
        self, driver, base_url, credentials, screenshots_dir
    ):
        """An API call rejected with 401 (invalid/expired access token) must
        surface the session-expired dialog instead of a silent/generic error.
        """
        self._login(driver, base_url, credentials, screenshots_dir)

        self._corrupt_access_token(driver)

        # Trigger an API call by navigating to a data-driven page.
        self._nav(driver, "Budgets")

        try:
            dialog = self._dialog(driver)
        except TimeoutException:
            driver.save_screenshot(f"{screenshots_dir}/session_03_no_dialog_401.png")
            pytest.fail(
                "Session-expired dialog did not appear after the API rejected "
                "the tampered access token with 401."
            )

        assert dialog.is_displayed()
        driver.save_screenshot(f"{screenshots_dir}/session_03_dialog_after_401.png")

    def test_missing_refresh_token_shows_session_dialog(
        self, driver, base_url, credentials, screenshots_dir
    ):
        """When the cached token is expired and no refresh token is available,
        the Auth0 SDK's silent-refresh failure must also surface the
        session-expired dialog.
        """
        self._login(driver, base_url, credentials, screenshots_dir)

        self._expire_session_without_refresh_token(driver)

        self._nav(driver, "Budgets")

        try:
            dialog = self._dialog(driver)
        except TimeoutException:
            driver.save_screenshot(f"{screenshots_dir}/session_04_no_dialog_refresh.png")
            pytest.fail(
                "Session-expired dialog did not appear after silent token "
                "refresh failed (no refresh_token available)."
            )

        assert dialog.is_displayed()
        driver.save_screenshot(f"{screenshots_dir}/session_04_dialog_after_refresh_fail.png")

    def test_session_expired_dialog_logs_out_on_click(
        self, driver, base_url, credentials, screenshots_dir
    ):
        """Clicking the dialog's action button logs the user out and clears
        the local Auth0 session, so a fresh visit requires logging in again.
        """
        self._login(driver, base_url, credentials, screenshots_dir)

        self._corrupt_access_token(driver)
        self._nav(driver, "Budgets")
        self._dialog(driver)

        driver.find_element(
            By.CSS_SELECTOR, "[data-testid='session-expired-logout']"
        ).click()

        try:
            self._wait(driver).until(
                lambda d: "auth0.com" in d.current_url or "/app" not in d.current_url
            )
        except TimeoutException:
            driver.save_screenshot(f"{screenshots_dir}/session_05_logout_stuck.png")
            pytest.fail(
                "Clicking the session-expired logout action did not navigate "
                f"away from the app. Current URL: {driver.current_url}"
            )

        driver.save_screenshot(f"{screenshots_dir}/session_05_after_logout.png")
