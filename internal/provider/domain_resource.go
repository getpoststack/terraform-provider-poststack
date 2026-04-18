package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DomainResource manages a sending domain.
//
// Update is intentionally a no-op for `name` (renaming a domain is not
// supported by the API — destroy + recreate). Other fields like
// open_tracking and click_tracking can be flipped without recreating.

var (
	_ resource.Resource                = &DomainResource{}
	_ resource.ResourceWithImportState = &DomainResource{}
)

type DomainResource struct {
	client *Client
}

type DomainResourceModel struct {
	ID            types.Int64  `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Status        types.String `tfsdk:"status"`
	Verified      types.Bool   `tfsdk:"verified"`
	OpenTracking  types.Bool   `tfsdk:"open_tracking"`
	ClickTracking types.Bool   `tfsdk:"click_tracking"`
}

func NewDomainResource() resource.Resource {
	return &DomainResource{}
}

func (r *DomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (r *DomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A PostStack sending domain.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "Numeric domain ID.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Domain name (e.g. mail.example.com). Forces replacement on change.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "DNS verification status (pending, verified, failed).",
			},
			"verified": schema.BoolAttribute{
				Computed:    true,
				Description: "True once SPF, DKIM, and DMARC have been verified.",
			},
			"open_tracking": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Inject open tracking pixels into outgoing HTML.",
			},
			"click_tracking": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Rewrite outgoing links to track clicks.",
			},
		},
	}
}

func (r *DomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Provider data type", fmt.Sprintf("expected *Client, got %T", req.ProviderData))
		return
	}
	r.client = client
}

type domainAPI struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	Verified      bool   `json:"verified"`
	OpenTracking  bool   `json:"open_tracking"`
	ClickTracking bool   `json:"click_tracking"`
}

func (r *DomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{"name": plan.Name.ValueString()}
	if !plan.OpenTracking.IsNull() && !plan.OpenTracking.IsUnknown() {
		body["open_tracking"] = plan.OpenTracking.ValueBool()
	}
	if !plan.ClickTracking.IsNull() && !plan.ClickTracking.IsUnknown() {
		body["click_tracking"] = plan.ClickTracking.ValueBool()
	}

	var created domainAPI
	if err := r.client.do("POST", "/domains", body, &created); err != nil {
		resp.Diagnostics.AddError("Failed to create domain", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, modelFromDomain(created))...)
}

func (r *DomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var d domainAPI
	if err := r.client.do("GET", fmt.Sprintf("/domains/%d", state.ID.ValueInt64()), nil, &d); err != nil {
		resp.Diagnostics.AddError("Failed to read domain", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, modelFromDomain(d))...)
}

func (r *DomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]interface{}{
		"open_tracking":  plan.OpenTracking.ValueBool(),
		"click_tracking": plan.ClickTracking.ValueBool(),
	}
	var updated domainAPI
	if err := r.client.do("PATCH", fmt.Sprintf("/domains/%d", plan.ID.ValueInt64()), body, &updated); err != nil {
		resp.Diagnostics.AddError("Failed to update domain", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, modelFromDomain(updated))...)
}

func (r *DomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.do("DELETE", fmt.Sprintf("/domains/%d", state.ID.ValueInt64()), nil, nil); err != nil {
		resp.Diagnostics.AddError("Failed to delete domain", err.Error())
		return
	}
}

func (r *DomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func modelFromDomain(d domainAPI) DomainResourceModel {
	return DomainResourceModel{
		ID:            types.Int64Value(d.ID),
		Name:          types.StringValue(d.Name),
		Status:        types.StringValue(d.Status),
		Verified:      types.BoolValue(d.Verified),
		OpenTracking:  types.BoolValue(d.OpenTracking),
		ClickTracking: types.BoolValue(d.ClickTracking),
	}
}
