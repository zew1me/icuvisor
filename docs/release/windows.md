# Windows release: MSI signing and publish setup

Companion to the macOS release runbook (kept maintainer-only at `docs/internal/release.md`). This doc covers the **Windows** half of the release pipeline: building a signed `.msi`, publishing to the Scoop bucket, and submitting Winget manifests.

The macOS DMG flow remains owned by the maintainer runbook. After macOS drafts the release, the new `release-windows` job in `.github/workflows/release.yml` builds the MSI; the `finalize` job then regenerates `SHA256SUMS.txt` and publishes the draft. Winget submission lives in a separate workflow that fires on `release: published`.

**Signing is optional and auto-gated.** Without the `AZURE_TENANT_ID` secret, the MSI ships unsigned (Windows SmartScreen will warn end users; acceptable for pre-v1 betas). Without the `SCOOP_BUCKET_PAT` secret, the scoop bucket update is skipped. Without the `WINGET_PAT` secret, the winget submission is skipped. Set the secrets when you are ready for that signal — no code changes needed.

## Pipeline overview

```
preflight (ubuntu)
   |
   v
release (macos)              <- builds linux/windows zips, dmg, runs goreleaser scoops
   |
   v
release-windows (windows)    <- builds .msi for amd64+arm64, signs if Azure secrets present, uploads to draft
   |
   v
finalize (ubuntu)            <- regenerates SHA256SUMS, publishes draft

[release: published]
   |
   v
winget.yml                   <- submits PR to microsoft/winget-pkgs (skipped without WINGET_PAT)
```

Artifacts produced per release (added to the existing macOS/Linux set):

- `icuvisor_<version>_windows_amd64.msi` — Authenticode-signed (Azure Trusted Signing) when configured; unsigned otherwise
- `icuvisor_<version>_windows_arm64.msi` — Authenticode-signed when configured; unsigned otherwise
- Scoop manifest committed to `ricardocabral/scoop-icuvisor` at `bucket/icuvisor.json` (skipped without `SCOOP_BUCKET_PAT`)
- Winget PR opened against `microsoft/winget-pkgs` under `RicardoCabral.icuvisor` (skipped without `WINGET_PAT`)

## 1. Azure Trusted Signing setup (optional)

Skip this entire section to ship an unsigned MSI — useful for early betas. The `release-windows` job will publish an unsigned `.msi`, log a warning, and continue. Set the Azure secrets later to flip signing back on with no code change.

Azure Trusted Signing is Microsoft's managed Authenticode signing service. Cert lives in an Azure-managed HSM; CI authenticates with an Entra ID app and calls the signing endpoint per release.

### 1.1 Create the Trusted Signing account and certificate profile

- [ ] Sign in to <https://portal.azure.com/> with the account that should own the publisher identity.
- [ ] Create or pick a subscription. Note that Trusted Signing is only available in select regions (East US, West Central US, West Europe, North Europe at time of writing).
- [ ] Search the portal for **Trusted Signing Accounts** and click **Create**.
- [ ] Fill in:
  - **Account name:** `icuvisor-signing` (anything unique within the resource group)
  - **Region:** pick the closest supported region (e.g. East US)
  - **Pricing tier:** Basic is sufficient for an OSS project (~$10/month at time of writing)
- [ ] After the account is created, open it and go to **Identity validation**.
- [ ] Submit the identity validation request. Choose **Individual** or **Organization** — this determines the **CN=** name on the Authenticode signature shown to Windows users. Organization validation requires DUNS or equivalent verification and can take several days.
- [ ] Wait for validation to complete (status becomes **Completed**).
- [ ] In the Trusted Signing account, open **Certificate profiles** and click **Create**.
- [ ] Profile type: **Public Trust**. Name it `icuvisor-public`. The profile name becomes `TRUSTED_SIGNING_PROFILE`.

### 1.2 Create the Entra ID app and grant signing permission

- [ ] In Azure portal, go to **Microsoft Entra ID > App registrations > New registration**.
- [ ] Name: `icuvisor-ci-signing`. Single tenant. No redirect URI.
- [ ] After creation, record:
  - **Application (client) ID** -> `AZURE_CLIENT_ID`
  - **Directory (tenant) ID** -> `AZURE_TENANT_ID`
- [ ] Open **Certificates & secrets > New client secret**. Set a 12-month expiration. Copy the **Value** (only shown once) -> `AZURE_CLIENT_SECRET`. Set a calendar reminder ~11 months out to rotate.
- [ ] Open the Trusted Signing account again. Go to **Access control (IAM) > Add role assignment**.
- [ ] Role: **Trusted Signing Certificate Profile Signer**. Assign to the `icuvisor-ci-signing` app registration.

### 1.3 Record non-secret endpoint and account names

- [ ] In the Trusted Signing account overview, copy the **Account URI** (looks like `https://eus.codesigning.azure.net`) -> `TRUSTED_SIGNING_ENDPOINT`.
- [ ] Copy the **Account name** (e.g. `icuvisor-signing`) -> `TRUSTED_SIGNING_ACCOUNT`.
- [ ] Copy the **Certificate profile name** (e.g. `icuvisor-public`) -> `TRUSTED_SIGNING_PROFILE`.

## 2. GitHub Actions secrets

All secrets here are **optional**. Missing secrets auto-disable the matching pipeline step (signing / scoop publish / winget submission) and log a warning instead of failing the release.

Add via the GitHub web UI (**Settings > Secrets and variables > Actions > New repository secret**) or `gh secret set`.

| Secret | Source | Gates |
| --- | --- | --- |
| `AZURE_TENANT_ID` | Entra ID Directory (tenant) ID | MSI signing (presence of this one secret toggles the signing step) |
| `AZURE_CLIENT_ID` | App registration Application (client) ID | MSI signing |
| `AZURE_CLIENT_SECRET` | App registration client secret value | MSI signing |
| `TRUSTED_SIGNING_ENDPOINT` | Trusted Signing account URI (e.g. `https://eus.codesigning.azure.net`) | MSI signing |
| `TRUSTED_SIGNING_ACCOUNT` | Trusted Signing account name | MSI signing |
| `TRUSTED_SIGNING_PROFILE` | Certificate profile name | MSI signing |
| `SCOOP_BUCKET_PAT` | Fine-grained PAT with `contents: write` on `ricardocabral/scoop-icuvisor` | Scoop bucket commit |
| `WINGET_PAT` | Classic PAT with `public_repo` scope (used by winget-releaser to fork microsoft/winget-pkgs and open a PR) | Winget submission |

`gh` form:

```bash
gh secret set AZURE_TENANT_ID
gh secret set AZURE_CLIENT_ID
gh secret set AZURE_CLIENT_SECRET
gh secret set TRUSTED_SIGNING_ENDPOINT
gh secret set TRUSTED_SIGNING_ACCOUNT
gh secret set TRUSTED_SIGNING_PROFILE
gh secret set SCOOP_BUCKET_PAT
gh secret set WINGET_PAT
```

## 3. Scoop bucket

The bucket repo `ricardocabral/scoop-icuvisor` is private at first and is **already created and seeded** with an empty `bucket/` directory. GoReleaser commits `bucket/icuvisor.json` on every tag push using `SCOOP_BUCKET_PAT`; without that secret, the scoops step is skipped silently.

When ready for public adoption (v1.0):

- [ ] Flip the bucket repo to public: `gh repo edit ricardocabral/scoop-icuvisor --visibility public`.
- [ ] Smoke test: `scoop bucket add icuvisor https://github.com/ricardocabral/scoop-icuvisor && scoop install icuvisor`.

## 4. Winget submission

The `winget.yml` workflow uses `vedantmgoyal9/winget-releaser` (community action) to:

- detect the `.msi` artifacts attached to the published release,
- generate v1.6 manifest YAMLs,
- fork `microsoft/winget-pkgs`, push a branch, open a PR.

Package identifier: `RicardoCabral.icuvisor`.

Before the first run:

- [ ] Reserve the publisher namespace by reading [microsoft/winget-pkgs#new-package-guidelines](https://github.com/microsoft/winget-pkgs/blob/master/doc/Authoring.md). Identifier `RicardoCabral.icuvisor` is fine if no one else has claimed `RicardoCabral`.
- [ ] Confirm the `WINGET_PAT` token user (typically the maintainer's GitHub account) is allowed to fork `microsoft/winget-pkgs`. No prior PR is required, but you can also pre-create a fork manually.
- [ ] The first PR will be reviewed by the winget moderators. Subsequent releases auto-PR.

## 5. WiX MSI internals

The MSI source is `build/windows/icuvisor.wxs` (WiX v4). Highlights:

- **Per-user install** (no UAC prompt). Files land in `%LOCALAPPDATA%\Programs\icuvisor`.
- **PATH update** is per-user (`System="no"`) so the binary is callable from any new shell. New shells only — existing shells need to be reopened.
- **UpgradeCode** is `8F557483-04AB-41DF-B138-3F979AF6496F` and **must not change**. Changing it would break MSI upgrade behavior for already-installed users.
- **Version** must be N.N.N (max each component 65535). The CI strips any `-SNAPSHOT`/`-rc` suffix before passing `Version=` to `wix build`. Pre-release tags like `v0.5.0-beta.1` therefore ship `0.5.0` as the MSI ProductVersion. The full semver still appears in the filename and release notes.

The `wix build` invocation passes the binary, license, and icon paths in via WiX preprocessor variables (`$(var.BinarySource)`, etc.) so the same `.wxs` works for amd64 and arm64.

## 6. Pre-flight before first Windows release

### 6.1 Minimum (unsigned MSI, no scoop/winget — fastest path to a first release)

- [ ] `goreleaser check` passes locally.
- [ ] Tag a release candidate (`v0.5.1-rc.1`) and watch all jobs go green. Expect a warning in `release-windows` saying signing is disabled; the unsigned `.msi` is still uploaded to the draft. `winget.yml` will not fire (prerelease).
- [ ] Smoke test the MSI from a Windows VM:
  - install, run `icuvisor version` from a new shell, uninstall via Apps & Features, confirm clean removal.
  - Windows SmartScreen will warn "Unknown publisher" on first install — expected for unsigned builds.

### 6.2 Full pipeline (signed MSI + scoop + winget)

- [ ] Trusted Signing identity validation completed; certificate profile created.
- [ ] All 8 secrets in section 2 added to the repo.
- [ ] `scoop-icuvisor` repo exists with a `main` branch and a `bucket/` directory.
- [ ] Tag a release candidate (`v0.5.1-rc.1`) and watch all jobs go green. `winget.yml` will not fire because the release is a prerelease — that is intentional.
- [ ] Inspect the signed MSI from a Windows VM:
  - right-click the `.msi` > **Properties > Digital Signatures** — should show the Trusted Signing identity and a valid timestamp.
  - `Get-AuthenticodeSignature .\icuvisor_*.msi` — `Status` must be `Valid`.
  - install, run `icuvisor version` from a new shell, uninstall via Apps & Features, confirm clean removal.
- [ ] Promote to a stable tag (`v0.5.1`). The `winget.yml` workflow will fire on the publish event and open a PR in `microsoft/winget-pkgs`.

## 7. Failure modes and recovery

- **Trusted Signing returns 403**: the Entra app is missing the **Trusted Signing Certificate Profile Signer** role on the account. Fix the role assignment and retry; no need to rotate the secret.
- **`wix build` fails with version error**: the tag included pre-release metadata that survived the strip step. Inspect the `Compute MSI version` step output.
- **Scoop manifest never appears in bucket repo**: `SCOOP_BUCKET_PAT` is missing `contents: write` on `scoop-icuvisor`, or expired. Regenerate.
- **Winget PR never opens**: `WINGET_PAT` is missing `public_repo`, or the PAT user is rate-limited on `microsoft/winget-pkgs`. Re-run `winget.yml` via workflow_dispatch with the tag.
- **MSI installs but `icuvisor` is not on PATH**: existing shells were opened before install. Open a new shell. If still missing, `setx PATH "%PATH%;%LOCALAPPDATA%\Programs\icuvisor"` and re-test.

## 8. What is _not_ wired up here

- **No NSIS installer.** Roadmap calls for `.msi` only; NSIS is left for future iteration if user demand appears.
- **No GoReleaser-Pro MSI/winget/scoops-with-msi.** Everything here is OSS goreleaser plus a community action. If we adopt GoReleaser Pro later, the `release-windows` job becomes a thin wrapper around `goreleaser release --split`.
- **No code signing of the raw `icuvisor.exe`** inside the MSI. Only the MSI itself is signed. If SmartScreen reputation requires per-binary signing later, sign the `.exe` in the same `release-windows` job before `wix build`.
