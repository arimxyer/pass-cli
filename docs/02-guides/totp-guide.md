---
title: "TOTP & 2FA Guide"
weight: 7
toc: true
---

# TOTP & 2FA Support

Pass-CLI supports storing TOTP (Time-based One-Time Password) secrets and generating 6-digit or 8-digit codes directly from your terminal. This eliminates the need for a separate authenticator app on your phone for many services.

## Adding TOTP to a Credential

You can add TOTP support to a new or existing credential in three ways:

### 1. Interactive Prompt (Recommended)

When adding or updating a credential, use the `--totp` flag to be prompted for the secret:

```bash
# Add new credential with TOTP
pass-cli add github --totp

# Add TOTP to existing credential
pass-cli update github --totp
```

### 2. Using an otpauth:// URI

If you have the full `otpauth://` URI (often provided by services for manual entry), you can provide it directly:

```bash
pass-cli add github --totp-uri "otpauth://totp/GitHub:user?secret=JBSWY3DPEHPK3PXP&issuer=GitHub"
```

### 3. In the TUI

1. Launch `pass-cli` (TUI mode)
2. Select a credential and press `e` to edit
3. Navigate to the TOTP field and enter your secret or URI
4. Save the changes

## Generating TOTP Codes

Once configured, you can generate codes using the `get` command:

```bash
# Display credential info including current TOTP code
pass-cli get github

# Display ONLY the TOTP code (useful for scripts)
pass-cli get github --totp --quiet

# Copy TOTP code to clipboard in TUI
# Select credential and press 't'
```

## QR Code Support

Pass-CLI can display and export QR codes, making it easy to sync credentials with other authenticator apps (like Authy or Google Authenticator).

### Display QR Code in Terminal

```bash
pass-cli get github --totp-qr
```

### Export QR Code to File

```bash
pass-cli get github --totp-qr-file github-qr.png
```

> **Warning**: Keep QR code files secure as they contain your TOTP secret in plain text.

## How Service and Username are Used

When generating a TOTP URI or QR code, Pass-CLI uses your credential's fields to build the label that appears in your authenticator app.

### Field Mapping

The following logic is used to determine what appears in your authenticator app:

1.  **Issuer (Company Name)**:
    - Uses `TOTPIssuer` if set.
    - Falls back to `Service` if `TOTPIssuer` is empty.
    - Example: "GitHub", "Google", "AWS"

2.  **Account Identifier**:
    - Uses `Username` if set.
    - Falls back to `Service` if `Username` is empty.
    - Example: "user@example.com", "admin"

### Fallback Behavior

- If `TOTPIssuer` is empty, the **Service** name is used as the issuer.
- If `Username` is empty, the **Service** name is used as the account name.
- If **both** Service and Username are empty, Pass-CLI will return an error when trying to generate a QR code, as it cannot build a valid URI.

### Best Practices

- **Set Username**: Always set the `Username` field if you have multiple accounts for the same service (e.g., personal vs. work GitHub accounts). This ensures they are distinguishable in your authenticator app.
- **Use Service for Identity**: If you only have one account for a service, leaving `Username` empty is fine; the `Service` name will be used for both the issuer and account fields.
- **Special Characters**: Spaces, hyphens, and other special characters in Service or Username are properly URL-encoded. Most authenticator apps (Google Authenticator, Authy, Microsoft Authenticator) decode and display them correctly.

### Example

If you have a credential with:
- **Service**: `GitHub - Work`
- **Username**: `dev-admin`

The QR code will show as **GitHub - Work (dev-admin)** in most authenticator apps.
