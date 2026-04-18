// Package provider implements the PostStack Terraform provider.
//
// The provider exposes three resources today: poststack_domain,
// poststack_api_key, and poststack_webhook. All three are thin wrappers
// around the corresponding /api endpoints, authenticated with a bearer
// token configured at the provider level.
package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Compile-time assertion that the provider satisfies the framework
// interface — catches signature drift on framework upgrades.
var _ provider.Provider = &PostStackProvider{}

type PostStackProvider struct {
	version string
}

type PostStackProviderModel struct {
	APIKey  types.String `tfsdk:"api_key"`
	BaseURL types.String `tfsdk:"base_url"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PostStackProvider{version: version}
	}
}

func (p *PostStackProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "poststack"
	resp.Version = p.version
}

func (p *PostStackProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "PostStack API key. May also be set via the POSTSTACK_API_KEY environment variable.",
			},
			"base_url": schema.StringAttribute{
				Optional:    true,
				Description: "PostStack API base URL. Defaults to https://api.poststack.dev.",
			},
		},
	}
}

func (p *PostStackProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data PostStackProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("POSTSTACK_API_KEY")
	if !data.APIKey.IsNull() && !data.APIKey.IsUnknown() {
		apiKey = data.APIKey.ValueString()
	}
	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing API key",
			"Set the api_key attribute or the POSTSTACK_API_KEY environment variable.",
		)
		return
	}

	baseURL := "https://api.poststack.dev"
	if v := os.Getenv("POSTSTACK_BASE_URL"); v != "" {
		baseURL = v
	}
	if !data.BaseURL.IsNull() && !data.BaseURL.IsUnknown() {
		baseURL = data.BaseURL.ValueString()
	}

	client := NewClient(apiKey, baseURL)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *PostStackProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDomainResource,
		NewAPIKeyResource,
		NewWebhookResource,
	}
}

func (p *PostStackProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
