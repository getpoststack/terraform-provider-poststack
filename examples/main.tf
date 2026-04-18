terraform {
  required_providers {
    poststack = {
      source  = "getpoststack/poststack"
      version = "~> 0.1"
    }
  }
}

provider "poststack" {
  # Or set POSTSTACK_API_KEY in the environment.
  api_key = var.poststack_api_key
}

variable "poststack_api_key" {
  type      = string
  sensitive = true
}

resource "poststack_domain" "marketing" {
  name           = "mail.example.com"
  open_tracking  = true
  click_tracking = true
}

resource "poststack_api_key" "ci" {
  name       = "CI deploy bot"
  permission = "sending_access"
}

resource "poststack_webhook" "events" {
  url    = "https://hooks.example.com/poststack"
  events = ["email.delivered", "email.bounced", "contact.created"]
}

output "webhook_secret" {
  value     = poststack_webhook.events.secret
  sensitive = true
}
