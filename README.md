# Terraform Provider for PostStack

Manage [PostStack](https://poststack.dev) sending domains, API keys, and
webhook subscriptions as code — for the
[EU-hosted, GDPR-compliant email API](https://poststack.dev) that's up to
75% cheaper than Resend, SendGrid, and Postmark.

Provider docs: [poststack.dev/integrations](https://poststack.dev/integrations).

## Resources

| Resource            | API endpoint | Notes                                       |
| ------------------- | ------------ | ------------------------------------------- |
| `poststack_domain`  | `/domains`   | Sending domain with SPF/DKIM/DMARC tracking |
| `poststack_api_key` | `/api-keys`  | Plaintext key returned only at create time  |
| `poststack_webhook` | `/webhooks`  | HMAC-signed webhook subscription            |

## Quickstart

```hcl
terraform {
  required_providers {
    poststack = {
      source  = "getpoststack/poststack"
      version = "~> 0.1"
    }
  }
}

provider "poststack" {
  api_key = var.poststack_api_key
}

resource "poststack_domain" "primary" {
  name = "mail.example.com"
}
```

See [`examples/main.tf`](./examples/main.tf) for a fuller example with
all three resource types.

## Build

```bash
go build -o terraform-provider-poststack
```

For local development, drop the binary into your
`~/.terraform.d/plugins/registry.terraform.io/getpoststack/poststack/0.1.0/<os>_<arch>/`
directory and `terraform init` will pick it up.

## Release

The provider is built and signed with [goreleaser](https://goreleaser.com)
and published to the [Terraform Registry](https://registry.terraform.io/).

```bash
goreleaser release --clean
```

## Links

- [PostStack — EU email API](https://poststack.dev)
- [API documentation](https://poststack.dev/docs)
- [Pricing](https://poststack.dev/pricing) · [Compare providers](https://poststack.dev/pricing/compare)
- [Integrations](https://poststack.dev/integrations) — Terraform, Zapier, Make.com, WordPress, Vercel, CLI
- [Status](https://poststack.dev/status) · [Security](https://poststack.dev/security)

## License

MIT — see [LICENSE](./LICENSE).
