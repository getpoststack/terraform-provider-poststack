// Package main is the entry point for the PostStack Terraform provider.
//
// Run `terraform init` against any module that has the provider in its
// `required_providers` block, and Terraform will exec this binary over
// the gRPC plugin protocol. The actual provider definition lives in
// internal/provider.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/getpoststack/terraform-provider-poststack/internal/provider"
)

// Set via -ldflags by goreleaser at release time.
var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with delve")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/getpoststack/poststack",
		Debug:   debug,
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}
