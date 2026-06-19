# 🔐 Security Incident Response Template

## Incident Report

Use this template to document security incidents.

---

## Incident Header

| Field | Value |
|-------|-------|
| Incident ID | INC-YYYY-MM-DD-001 |
| Date Reported | YYYY-MM-DD |
| Date Resolved | YYYY-MM-DD |
| Duration | X hours Y minutes |
| Severity | Critical / High / Medium / Low |
| Reporter | Name (role) |
| Investigator | Name (role) |

---

## Executive Summary

One-paragraph summary suitable for non-technical stakeholders.

**Example:**
"On 2024-04-27, an unauthorized API token was exposed in a GitHub commit. The token was immediately revoked, access logs were reviewed, and no evidence of misuse was found. All developers were notified and the commit was removed from history."

---

## Incident Classification

### Severity Levels

- **Critical**: Data breach, RCE, complete system compromise
- **High**: Unauthorized access, confidential data exposure
- **Medium**: Configuration issue, potential vulnerability
- **Low**: Policy violation, security notice

### Incident Type

- [ ] Data Breach
- [ ] Unauthorized Access
- [ ] Malware / Ransomware
- [ ] DDoS / Availability
- [ ] Credential Exposure
- [ ] Misconfiguration
- [ ] Policy Violation
- [ ] Other: _______________

---

## Timeline (Chronological)

| Time | Event | Details |
|------|-------|---------|
| 2024-04-27 14:00 UTC | Detection | Alert triggered by secret scanning |
| 2024-04-27 14:05 UTC | Initial Response | On-call engineer notified |
| 2024-04-27 14:15 UTC | Investigation Started | Root cause analysis began |
| 2024-04-27 14:30 UTC | Credential Revoked | API token revoked in provider |
| 2024-04-27 15:00 UTC | Access Logs Reviewed | No unauthorized access detected |
| 2024-04-27 15:30 UTC | Remediation Complete | Commit removed, notification sent |
| 2024-04-27 16:00 UTC | Post-Incident Review | Team debriefing completed |

---

## Root Cause Analysis

### What Happened?

Describe the incident in detail, including:
- What was the security issue?
- How was it discovered?
- What systems/data were affected?
- What was the attack vector?

### Why Did It Happen?

Analyze the root cause:
- Was it a technical failure?
- Was it a process failure?
- Was it human error?
- Was it a third-party compromise?

### Contributing Factors

List other factors that enabled the incident:
- Lack of monitoring
- Outdated libraries
- Weak access controls
- Insufficient testing
- Poor documentation

---

## Impact Assessment

### Affected Systems

- [ ] Database
- [ ] API servers
- [ ] Web application
- [ ] Mobile application
- [ ] Infrastructure
- [ ] Other: _______________

### Data Exposure

| Data Category | Amount | Impact |
|---------------|--------|--------|
| User Accounts | 0 | None |
| Payment Data | 0 | None |
| Personal Data | 0 | None |
| Business Data | 0 | None |
| System Data | Limited | Low |

### Users Impacted

- Total affected: 0 users
- Notification sent: Yes / No
- User action required: Yes / No

### Financial Impact

- Estimated cost: $0
- Incident response: $X
- System recovery: $X
- Compliance/legal: $X
- Reputation: $X

---

## Immediate Response Actions

### Actions Taken

- [x] Identified root cause
- [x] Contained the incident
- [x] Isolated affected systems
- [x] Revoked credentials
- [x] Notified security team
- [x] Preserved evidence
- [x] Documented timeline
- [ ] Notified customers
- [ ] Reported to authorities
- [ ] Engaged external consultants

### Containment

Describe actions taken to prevent further damage:

1. **Immediate (0-1 hour)**
   - Revoked API token
   - Revoked SSH keys
   - Killed active sessions

2. **Short-term (1-24 hours)**
   - Updated firewall rules
   - Reset passwords
   - Applied security patches

3. **Long-term (1-30 days)**
   - Updated access controls
   - Enhanced monitoring
   - Implemented new safeguards

---

## Investigation Details

### Evidence Collected

- [ ] System logs
- [ ] Application logs
- [ ] Network traffic
- [ ] Firewall logs
- [ ] Access logs
- [ ] Code repository history
- [ ] Email/chat history
- [ ] Third-party reports

### Findings

| Finding | Severity | Status |
|---------|----------|--------|
| API token in Git history | High | Resolved |
| Weak password policy | Medium | In Progress |
| Missing rate limiting | Medium | Planned |

---

## Remediation & Recovery

### Short-term Fixes (Implemented)

- [ ] Revoked compromised credentials
- [ ] Updated affected systems
- [ ] Enhanced monitoring
- [ ] Notified stakeholders

### Long-term Improvements (Planned)

- [ ] Implement pre-commit hooks
- [ ] Enable secret scanning
- [ ] Enhance monitoring/alerting
- [ ] Update security policies
- [ ] Staff security training
- [ ] Regular penetration testing
- [ ] Improve incident response procedures

### Prevention Measures

1. **Technical Controls**
   - Pre-commit secret scanning
   - Code review requirements
   - Automated security testing
   - Rate limiting

2. **Process Controls**
   - Security training for all developers
   - Incident response drills quarterly
   - Regular security audits
   - Access reviews

3. **Personnel Controls**
   - Background checks
   - Security awareness training
   - Least privilege access
   - Separation of duties

---

## Lessons Learned

### What Went Well

1. Alert was triggered quickly
2. On-call engineer responded within 5 minutes
3. Credential was revoked immediately
4. No unauthorized access detected

### What Could Be Better

1. Pre-commit hooks not enabled by default
2. No rate limiting on API
3. Limited monitoring of Git activity
4. Lack of developer security awareness

### Action Items

| Action | Owner | Due Date | Priority |
|--------|-------|----------|----------|
| Enable pre-commit hooks | Dev Lead | 2024-05-04 | High |
| Implement rate limiting | DevOps | 2024-05-11 | High |
| Security training | HR | 2024-05-18 | Medium |
| Audit access logs | Security | 2024-05-01 | High |

---

## Compliance & Notification

### Regulatory Requirements

- [ ] GDPR: Data breach notification sent (if EU residents affected)
- [ ] CCPA: Consumer notification sent (if CA residents affected)
- [ ] HIPAA: Breach notification sent (if health data affected)
- [ ] PCI-DSS: Incident reported to acquiring bank
- [ ] Other: _______________

### Stakeholder Notifications

- [ ] Executive team
- [ ] Legal team
- [ ] Insurance carrier
- [ ] Affected customers
- [ ] Regulatory bodies
- [ ] Law enforcement

### Communication Template

```
Subject: Security Incident Notice - Immediate Action Required

Dear [Customer],

We are writing to inform you of a security incident that may have affected your account.

INCIDENT DETAILS:
- Date: YYYY-MM-DD
- Type: [Credential Exposure / Data Breach / etc.]
- Your Data Affected: [Yes / No]

WHAT WE'VE DONE:
- Identified and contained the issue
- Revoked compromised credentials
- Enhanced security controls
- Reviewed access logs

WHAT YOU SHOULD DO:
- Change your password immediately
- Enable two-factor authentication
- Monitor your account for unusual activity
- Contact us if you have questions

For more information: [Support URL]

Security Team
```

---

## Post-Incident Review Meeting

### Meeting Date: YYYY-MM-DD

### Attendees

- [ ] Incident Commander
- [ ] Security Lead
- [ ] Development Lead
- [ ] DevOps Lead
- [ ] CTO

### Key Discussion Points

1. **What happened?** (5 min)
   - Brief overview of incident

2. **Why did it happen?** (10 min)
   - Root cause analysis

3. **How did we respond?** (5 min)
   - Response effectiveness

4. **What will we do differently?** (15 min)
   - Preventive measures
   - Process improvements

5. **Action items** (5 min)
   - Assign owners
   - Set deadlines

---

## Sign-Off

| Role | Name | Signature | Date |
|------|------|-----------|------|
| Incident Commander | _____ | _____ | ____ |
| Security Lead | _____ | _____ | ____ |
| CTO | _____ | _____ | ____ |

---

## Appendices

### A. Evidence Files

- Access logs: `/var/log/auth.log`
- System logs: `/var/log/syslog`
- Application logs: `/var/log/app/*.log`

### B. Related Policies

- Security Policy: `/docs/SECURITY.md`
- Incident Response Plan: `/docs/INCIDENT-RESPONSE.md`
- Data Protection: `/docs/DATA-PROTECTION.md`

### C. External References

- NIST Incident Handling: https://nvlpubs.nist.gov/nistpubs/SpecialPublications/NIST.SP.800-61r2.pdf
- ISO 27035: Information security incident management

---

*Document Classification: CONFIDENTIAL*
*Retention Period: 7 years (regulatory requirement)*
