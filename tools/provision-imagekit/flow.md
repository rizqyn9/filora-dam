# ImageKit Registration → Verification → Login → Dashboard State

## Parameters

- `N`: Batch count
- `password`: Password from `config.yaml`
- `gmail`: Authorized mailbox used for verification emails
- `concurrency`: Maximum number of simultaneous accounts

---

# 1. Registration

For each account:

1. Create a dedicated isolated browser context.
2. Open the registration page.
3. Fill:
   - Business email
   - Name
   - Password
4. Uncheck the marketing communication checkbox.
5. Submit registration.

## Registration Result

### Expected

If:

`Email Verification Pending`

then set:

`EMAIL_VERIFICATION_PENDING`

### CAPTCHA / Security Challenge

If CAPTCHA/security verification appears:

1. Pause only this account.
2. Notify the operator.
3. Preserve the browser context.
4. Operator completes the challenge manually.
5. Re-check the registration state.
6. Continue only when the expected state is reached.

Do not attempt to bypass or defeat the CAPTCHA.

### Unexpected State

If the expected registration state is not reached:

- Capture screenshot for debugging.
- Capture URL.
- Capture visible non-sensitive error text.
- Mark account as `REGISTRATION_FAILED`.
- Do not continue blindly.

---

# 2. Email Verification

After registration reaches:

`EMAIL_VERIFICATION_PENDING`

wait for the verification email in the authorized Gmail mailbox.

## Email Matching

Do NOT simply use the newest email.

Match the email to the specific account using available
non-sensitive identifiers such as:

- recipient address
- sender
- subject
- timestamp
- registration/account ID if available

Example:

```text
Account A001
    ↓
verification email
    ↓
Gmail
    ↓
OTP/Email Worker
    ↓
A001 browser context
````

## Extract Verification URL

Read the email body and identify the verification URL.

Requirements:

* Only accept URLs from the expected authorized domain.
* Reject unrelated links.
* Do not log the complete URL if it contains sensitive tokens.
* Mark the email as consumed after successful processing.

Then navigate the account's browser context to the verification URL.

---

# 3. Verification Redirect

The verification URL normally redirects to the login page.

Detect:

`/login`

or the equivalent authorized login route.

Do not assume the redirect destination without checking the current URL.

---

# 4. Login

Fill:

* Email → account email
* Password → configured password

Submit the login form.

---

# 5. Login CAPTCHA / Human Intervention

If a CAPTCHA/security challenge appears after login:

1. Pause only the affected account.
2. Preserve the browser context.
3. Notify the operator.
4. Operator manually completes the challenge.
5. Wait for the page to become usable.
6. Re-check authentication state.
7. Continue the workflow.

Do not attempt to bypass the CAPTCHA.

State:

`LOGIN_HUMAN_REQUIRED`

After resolution:

`LOGIN_HUMAN_RESOLVED`

---

# 6. Determine Onboarding State

After successful authentication, inspect the resulting route.

Possible states:

### Already onboarded

```text
/dashboard
```

Set:

`ONBOARDING_COMPLETED`

### Not onboarded

```text
/dashboard/get-started
```

Set:

`ONBOARDING_REQUIRED`

The application may redirect `/dashboard` to the onboarding page when
the account has not completed onboarding.

Therefore, `/dashboard` should not be treated as proof of onboarding completion.

---

# 7. Dashboard State Verification

For an authenticated account:

1. Navigate to `/dashboard`.
2. Wait for navigation/redirect to settle.
3. Inspect the final URL.
4. Determine the account state.

Example:

```text
/dashboard
    ↓
final URL == /dashboard
    → ONBOARDING_COMPLETED

/dashboard
    ↓
redirect
    ↓
/dashboard/get-started
    → ONBOARDING_REQUIRED
```

This provides a second verification mechanism for the account's onboarding state.

---

# 8. Obtain Account Configuration

Only for authorized test accounts, obtain the required configuration
information from the dashboard/configuration UI.

Expected fields may include:

* ImageKit ID
* Endpoint
* Public key
* Private key

Treat all keys as secrets.

## Secret Handling

NEVER:

* print private keys to console
* include private keys in screenshots
* put private keys in Git
* expose private keys in error messages
* send private keys to external logging systems
* include private keys in normal automation traces

Use a secret store or encrypted local storage.

Example internal representation:

```yaml
account_id: A001
imagekit_id: ...
endpoint: ...
public_key: ...
private_key: "[SECRET]"
```

---

# 9. Private Key Reveal

The private key may require an additional verification step.

Expected flow:

```text
Dashboard
   ↓
Private Key
   ↓
Reveal
   ↓
OTP Required
   ↓
OTP sent to registered account
   ↓
Retrieve OTP from authorized mailbox
   ↓
Enter OTP
   ↓
Reveal private key
```

When OTP is requested:

1. Associate the OTP request with the account.
2. Wait for the corresponding verification email.
3. Extract only the required OTP.
4. Enter the OTP into the current account's browser context.
5. Verify the reveal operation succeeded.
6. Store the private key securely.

Never assume that the newest OTP belongs to the current account.

---

# 10. Account State

Maintain an explicit state for every account.

```text
CREATED
REGISTRATION
REGISTRATION_HUMAN_REQUIRED
EMAIL_VERIFICATION_PENDING
EMAIL_VERIFIED
LOGIN
LOGIN_HUMAN_REQUIRED
AUTHENTICATED
ONBOARDING_REQUIRED
ONBOARDING_COMPLETED
CONFIG_RETRIEVAL
PRIVATE_KEY_OTP_PENDING
PRIVATE_KEY_REVEALED
COMPLETED
FAILED
```

Example:

```text
A001 → COMPLETED
A002 → LOGIN_HUMAN_REQUIRED
A003 → EMAIL_VERIFICATION_PENDING
A004 → ONBOARDING_REQUIRED
A005 → FAILED
```

A human intervention event must pause only its associated account,
not the complete batch.

````

### One thing I'd change from your original design

Don't model CAPTCHA as:

```text
CAPTCHA → solve → continue
````

Model it as a **state transition**:

```text
RUNNING
   ↓
HUMAN_REQUIRED
   ↓
operator resolves
   ↓
VERIFY_STATE
   ↓
RUNNING
```

That matters because after a human solves the CAPTCHA, you should **verify the application's state again** rather than assuming the previous action succeeded.

And for the private key, I'd keep it out of the normal account object entirely. Use something like:

```js
{
  id: "A001",
  email: "...",
  status: "COMPLETED",
  imagekitId: "...",
  endpoint: "...",
  publicKey: "...",
  secretRef: "vault://imagekit/A001/private-key"
}
```

rather than:

```js
{
  privateKey: "actual-secret-here"
}
```

This also makes your Kiro-generated implementation considerably safer to debug and run concurrently.
