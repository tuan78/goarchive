# Security Policy

## Supported Versions

We currently support the following versions with security updates:

| Version | Supported          |
| ------- | ------------------ |
| main    | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

If you discover a security vulnerability, please send an email to <tuantla0708@gmail.com> with:

- A description of the vulnerability
- Steps to reproduce the issue
- Potential impact
- Any suggested fixes (optional)

### What to Expect

- **Initial Response**: We aim to acknowledge your report within 48 hours
- **Status Updates**: We'll keep you informed about our progress
- **Disclosure**: Once the vulnerability is fixed, we'll work with you on responsible disclosure
- **Credit**: We'll credit you in our security advisory (unless you prefer to remain anonymous)

### Security Considerations for GoArchive

When using GoArchive, please consider:

1. **Credentials Management**
   - Never commit credentials to version control
   - Use environment variables or secure secret management (e.g., AWS Secrets Manager, HashiCorp Vault)
   - Rotate credentials regularly

2. **Database Connections**
   - Always use SSL/TLS for production database connections
   - Set `DB_SSLMODE=require` or `verify-full` for PostgreSQL
   - Limit database user permissions to only what's needed for backups

3. **Storage Security**
   - Enable encryption at rest for S3 buckets
   - Use bucket policies to restrict access
   - Enable versioning and MFA delete for production backups
   - Consider using VPC endpoints for private connectivity

4. **Network Security**
   - Run backups from trusted networks
   - Use network policies in Kubernetes
   - Consider using PrivateLink/VPC peering for cloud resources

5. **Access Control**
   - Use IAM roles instead of access keys when possible
   - Follow principle of least privilege
   - Audit access logs regularly

6. **Backup Data**
   - Remember that backups may contain sensitive data
   - Apply appropriate data classification
   - Consider encrypting backups before upload
   - Implement retention and deletion policies

## Security Best Practices

### For Contributors

- Keep dependencies up to date
- Run `go mod tidy` and review changes
- Avoid introducing known vulnerable dependencies
- Use `gosec` or similar tools for security scanning

### For Users

- Keep GoArchive updated to the latest version
- Review and audit plugin code before use
- Test backups and restoration regularly
- Monitor backup jobs for failures or anomalies
- Implement backup verification

## Known Security Considerations

### pg_dump and Passwords

GoArchive uses `PGPASSWORD` environment variable for PostgreSQL backups. While convenient, be aware:

- The password may be visible in process listings
- Consider using `.pgpass` file or connection URIs for production
- We're exploring more secure authentication methods

### Backup Contents

- Backups are **not encrypted by default** during transfer or storage
- Enable S3 encryption if storing sensitive data
- Consider implementing application-level encryption for highly sensitive data

## Security Updates

We'll announce security updates through:

- GitHub Security Advisories
- Release notes
- Git tags with `security` label

Subscribe to repository releases to stay informed.

---

Thank you for helping keep GoArchive and its users safe!
