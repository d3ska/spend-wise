# Threat Model — SpendWise

**Date:** 2026-02-17
**Methodology:** STRIDE + MITRE ATT&CK
**Target:** SpendWise collaborative household financial ledger
**Total Threats Identified:** 20 (across 6 STRIDE categories)
**MITRE ATT&CK Techniques Mapped:** 18

---

## Executive Summary

SpendWise handles sensitive financial data including bank account connections via Enable Banking, expense tracking with monetary values, and workspace-based multi-tenancy. The application's threat landscape is dominated by three P0 critical issues: a committed private RSA key granting bank API access, a predictable JWT default secret enabling token forgery, and an unguarded authentication bypass flag. These, combined with OAuth state validation gaps and sensitive data logging, create a significant risk profile for a financial application.

---

## Trust Boundaries

```
┌─────────────────────────────────────────────────────┐
│  EXTERNAL (Untrusted)                                │
│  ┌──────────┐  ┌──────────┐  ┌──────────────────┐  │
│  │ Browser  │  │ OAuth    │  │ Enable Banking   │  │
│  │ Client   │  │ Providers│  │ API              │  │
│  └────┬─────┘  └────┬─────┘  └────────┬─────────┘  │
│       │              │                  │            │
├───────┼──────────────┼──────────────────┼────────────┤
│  BOUNDARY: TLS / Internet Gateway                    │
├───────┼──────────────┼──────────────────┼────────────┤
│  DMZ  │              │                  │            │
│  ┌────▼──────────────▼──────────────────▼─────────┐ │
│  │              Chi Router (handler/)              │ │
│  │   - CORS middleware                             │ │
│  │   - JWT validation middleware                   │ │
│  │   - Request ID / Logger / Recoverer             │ │
│  └────────────────────┬───────────────────────────┘ │
│                       │                              │
├───────────────────────┼──────────────────────────────┤
│  BOUNDARY: Auth Middleware                           │
├───────────────────────┼──────────────────────────────┤
│  INTERNAL             │                              │
│  ┌────────────────────▼───────────────────────────┐ │
│  │            service/ (Business Logic)            │ │
│  └────────────────────┬───────────────────────────┘ │
│                       │                              │
│  ┌────────────────────▼───────────────────────────┐ │
│  │        store/ (PostgreSQL via pgx/sqlc)         │ │
│  └────────────────────┬───────────────────────────┘ │
│                       │                              │
├───────────────────────┼──────────────────────────────┤
│  BOUNDARY: Database Network                          │
├───────────────────────┼──────────────────────────────┤
│  ┌────────────────────▼───────────────────────────┐ │
│  │              PostgreSQL Database                 │ │
│  └────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────┘
```

---

## STRIDE Threat Analysis

### S — Spoofing Identity (5 threats)

| ID | Threat | Vulnerability | CVSS | Priority |
|----|--------|---------------|------|----------|
| S-1 | JWT token forgery via default secret | VULN-002 | 9.1 | P0 |
| S-2 | Auth bypass via dev flag in production | VULN-003 | 9.0 | P0 |
| S-3 | OAuth CSRF — attacker links own account to victim session | VULN-004 | 8.1 | P1 |
| S-4 | Bank API impersonation via leaked RSA key | VULN-001 | 9.8 | P0 |
| S-5 | Session hijacking via missing HttpOnly/Secure cookie flags | VULN-010 | 5.0 | P2 |

### T — Tampering with Data (3 threats)

| ID | Threat | Vulnerability | CVSS | Priority |
|----|--------|---------------|------|----------|
| T-1 | Expense amount manipulation via forged JWT | VULN-002 | 9.1 | P0 |
| T-2 | Cross-workspace data modification (tenant isolation bypass) | New finding | 7.0 | P1 |
| T-3 | Category rule injection via malicious regex | VULN-005 | 7.5 | P1 |

### R — Repudiation (2 threats)

| ID | Threat | Vulnerability | CVSS | Priority |
|----|--------|---------------|------|----------|
| R-1 | Unaudited financial transactions — no security event logging | VULN-015 | 5.0 | P2 |
| R-2 | No audit trail for workspace membership changes | New finding | 4.0 | P3 |

### I — Information Disclosure (4 threats)

| ID | Threat | Vulnerability | CVSS | Priority |
|----|--------|---------------|------|----------|
| I-1 | Bank credentials/account data exposed in logs | VULN-006 | 7.0 | P1 |
| I-2 | Private RSA key accessible from repository | VULN-001 | 9.8 | P0 |
| I-3 | Internal error details leaked to clients | VULN-013 | 3.1 | P3 |
| I-4 | Database connection string exposure | VULN-002 (related) | 4.0 | P3 |

### D — Denial of Service (3 threats)

| ID | Threat | Vulnerability | CVSS | Priority |
|----|--------|---------------|------|----------|
| D-1 | ReDoS via user-controlled regex patterns | VULN-005 | 7.5 | P1 |
| D-2 | Auth endpoint brute force (no rate limiting) | VULN-009 | 5.3 | P2 |
| D-3 | SSE connection exhaustion (no connection limits) | New finding | 5.0 | P2 |

### E — Elevation of Privilege (3 threats)

| ID | Threat | Vulnerability | CVSS | Priority |
|----|--------|---------------|------|----------|
| E-1 | Full admin access via dev bypass | VULN-003 | 9.0 | P0 |
| E-2 | Cross-workspace access via IDOR on workspace IDs | New finding | 7.5 | P1 |
| E-3 | Privilege escalation via JWT claims manipulation | VULN-002 | 9.1 | P0 |

---

## Risk Matrix

```
                    IMPACT
           Low    Medium    High    Critical
         ┌────────┬─────────┬───────┬──────────┐
 High    │        │ D-2,D-3 │ T-2,  │ S-1,S-2, │  ← LIKELIHOOD
         │        │ R-1     │ S-3,  │ S-4,E-1, │
         │        │         │ D-1,  │ E-3,T-1  │
         │        │         │ I-1   │          │
         ├────────┼─────────┼───────┼──────────┤
 Medium  │        │ S-5     │ E-2,  │ I-2      │
         │        │         │ T-3   │          │
         ├────────┼─────────┼───────┼──────────┤
 Low     │        │ I-3,I-4 │ R-2   │          │
         │        │         │       │          │
         └────────┴─────────┴───────┴──────────┘
```

---

## MITRE ATT&CK Mapping

| Technique ID | Technique Name | Related Threats |
|-------------|----------------|-----------------|
| T1078 | Valid Accounts | S-1, S-2, E-1, E-3 |
| T1552.001 | Unsecured Credentials: Credentials in Files | S-4, I-2 (RSA key in repo) |
| T1539 | Steal Web Session Cookie | S-5 |
| T1550.001 | Use Alternate Authentication Material: Application Access Token | S-1, E-3 (JWT forgery) |
| T1185 | Browser Session Hijacking | S-3 (OAuth CSRF) |
| T1565.001 | Data Manipulation: Stored Data Manipulation | T-1, T-2 |
| T1059 | Command and Scripting Interpreter | T-3 (regex injection) |
| T1070.003 | Indicator Removal: Clear Command History | R-1, R-2 (no audit trail) |
| T1530 | Data from Cloud Storage Object | I-1 (log data exposure) |
| T1190 | Exploit Public-Facing Application | D-1, D-2, D-3 |
| T1498 | Network Denial of Service | D-2, D-3 |
| T1499.004 | Application or System Exploitation | D-1 (ReDoS) |
| T1068 | Exploitation for Privilege Escalation | E-2 (IDOR) |
| T1548 | Abuse Elevation Control Mechanism | E-1 (dev bypass) |
| T1557 | Adversary-in-the-Middle | S-5, VULN-010 (missing HSTS) |
| T1071.001 | Application Layer Protocol: Web Protocols | All API threats |
| T1586 | Compromise Accounts | S-1 via JWT forgery |
| T1199 | Trusted Relationship | S-4 (bank API impersonation) |

---

## Attack Scenarios

### Scenario 1: Complete Account Takeover via JWT Forgery (P0)
**Vector:** Attacker discovers default JWT secret → forges JWT for any user → full access to all expenses, workspaces, bank connections
**Impact:** Total compromise of all user data, financial records, and bank account access
**Likelihood:** HIGH — default secret is well-known pattern, discoverable via source code
**MITRE ATT&CK:** T1078 → T1550.001 → T1565.001

### Scenario 2: Bank Account Data Theft via Leaked RSA Key (P0)
**Vector:** Attacker extracts RSA key from git repo → authenticates to Enable Banking API → accesses all connected bank accounts
**Impact:** Unauthorized access to users' bank account data, transaction history
**Likelihood:** HIGH — key is in the repository, trivially extractable
**MITRE ATT&CK:** T1552.001 → T1199 → T1530

### Scenario 3: Authentication Bypass to Admin Access (P0)
**Vector:** AUTH_BYPASS flag set to true in production → all requests authenticated as dev user → full API access without credentials
**Impact:** Complete bypass of all authentication and authorization
**Likelihood:** MEDIUM — depends on deployment practices, but risk of misconfiguration is real
**MITRE ATT&CK:** T1078 → T1548

### Scenario 4: OAuth CSRF Account Linking (P1)
**Vector:** Attacker initiates OAuth flow → obtains callback URL → tricks victim into visiting → attacker's OAuth account linked to victim's session
**Impact:** Account takeover, persistent access to victim's financial data
**Likelihood:** MEDIUM — requires social engineering
**MITRE ATT&CK:** T1185 → T1078

### Scenario 5: Cross-Workspace Data Breach via IDOR (P1)
**Vector:** Authenticated user manipulates workspace ID in API requests → accesses expenses/data from other workspaces
**Impact:** Privacy breach, unauthorized access to other households' financial data
**Likelihood:** MEDIUM — requires authenticated access, depends on authorization checks
**MITRE ATT&CK:** T1068 → T1530

### Scenario 6: Denial of Service via ReDoS (P1)
**Vector:** Authenticated user creates category rule with catastrophic regex → pattern compilation/matching consumes all CPU
**Impact:** Application becomes unresponsive for all users
**Likelihood:** MEDIUM — requires authentication but trivial to exploit
**MITRE ATT&CK:** T1499.004

---

## Business Impact Analysis

| Impact Area | Severity | Description |
|-------------|----------|-------------|
| **Financial Data Exposure** | Critical | Bank account numbers, transaction history, expense details of all users |
| **Regulatory/Legal** | High | Potential GDPR violations (financial PII), banking data protection regulations |
| **Reputational Damage** | Critical | Users trust app with bank connections; breach destroys trust permanently |
| **Financial Loss** | High | Potential liability for unauthorized bank access, legal costs |
| **Service Availability** | Medium | ReDoS and SSE exhaustion could cause outages for all users |
| **Data Integrity** | High | Forged JWT tokens could manipulate expense records, split calculations |

---

## Prioritized Remediation Roadmap

### P0 — Immediate (24-48 hours)
1. **Revoke and rotate** the Enable Banking RSA key (VULN-001/S-4/I-2)
2. **Remove JWT default secret** — fail-fast if not configured (VULN-002/S-1/E-3/T-1)
3. **Guard auth bypass** — build tags or remove from production builds (VULN-003/S-2/E-1)
4. **Purge RSA key from git history** (BFG Repo Cleaner)

### P1 — Urgent (1 week)
5. **Validate OAuth state parameter** on callback (VULN-004/S-3)
6. **Add regex pattern length limits** and validation (VULN-005/D-1/T-3)
7. **Redact bank API data from logs** (VULN-006/I-1)
8. **Add workspace-scoped authorization checks** (E-2/T-2)

### P2 — Important (2 weeks)
9. **Add rate limiting** on auth endpoints (VULN-009/D-2)
10. **Add security headers** middleware (VULN-010/S-5)
11. **Add SSE connection limits** per user (D-3)
12. **Implement security event logging** (VULN-015/R-1)

### P3 — Standard (1 month)
13. **Update vulnerable dependencies** — chi, x/crypto (VULN-007/VULN-008)
14. **Sanitize error messages** (VULN-013/I-3)
15. **Add audit trail** for workspace membership (R-2)
16. **Restrict CORS origins** (VULN-011)

### P4 — Hardening (ongoing)
17. **Account lockout mechanism** (VULN-014)
18. **Consistent request size limits** (VULN-012)
19. **Comprehensive security test suite**
20. **CI/CD security scanning integration**
