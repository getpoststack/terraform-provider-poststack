package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// APIKeyResource manages a PostStack API key.
//
// The plaintext key is only returned by the create endpoint, so we
// capture it in state on Create and never re-fetch it on Read. Marking
// the secret with RequiresReplace ensures Terraform recreates the key
// rather than ever leaving it null in state.

var _ resource.Resource = &APIKeyResource{}

type APIKeyResource struct {
	client *Client
}

type APIKeyResourceModel struct {
	ID         types.Int64  `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Permission types.String `tfsdk:"permission"`
	Mode       types.String `tfsdk:"mode"`
	Key        types.String `tfsdk:"key"`
}

func NewAPIKeyResource() resource.Resource {
	return &APIKeyResource{}
}

func (r *APIKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *APIKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A PostStack API key. The plaintext secret is exposed once at create time and stored in Terraform state — protect that state file.",
		Attributes: map[string]schema.Attribute{
			"id":   schema.Int64Attribute{Computed: true},
			"name": schema.StringAttribute{Required: true},
			"permission": schema.StringAttribute{
				Required:    true,
				Description: "Either `full_access` or `sending_access`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mode": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Either `live` or `test`. Defaults to live.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Plaintext API key. Returned only at create time.",
			},
		},
	}
}

func (r *APIKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client, _ = req.ProviderData.(*Client)
}

type apiKeyAPI struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Permission string `json:"permission"`
	Mode       string `json:"mode"`
	Key        string `json:"key,omitempty"`
}

func (r *APIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan APIKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]interface{}{
		"name":       plan.Name.ValueString(),
		"permission": plan.Permission.ValueString(),
	}
	if !plan.Mode.IsNull() && !plan.Mode.IsUnknown() {
		body["mode"] = plan.Mode.ValueString()
	}

	var created apiKeyAPI
	if err := r.client.do("POST", "/api-keys", body, &created); err != nil {
		resp.Diagnostics.AddError("Failed to create API key", err.Error())
		return
	}

	plan.ID = types.Int64Value(created.ID)
	plan.Permission = types.StringValue(created.Permission)
	plan.Mode = types.StringValue(created.Mode)
	plan.Key = types.StringValue(created.Key)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *APIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state APIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var k apiKeyAPI
	if err := r.client.do("GET", fmt.Sprintf("/api-keys/%d", state.ID.ValueInt64()), nil, &k); err != nil {
		resp.Diagnostics.AddError("Failed to read API key", err.Error())
		return
	}
	state.Name = types.StringValue(k.Name)
	state.Permission = types.StringValue(k.Permission)
	state.Mode = types.StringValue(k.Mode)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *APIKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Only `name` is mutable. Plaintext key + permission/mode require replacement.
	var plan APIKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]interface{}{"name": plan.Name.ValueString()}
	if err := r.client.do("PATCH", fmt.Sprintf("/api-keys/%d", plan.ID.ValueInt64()), body, nil); err != nil {
		resp.Diagnostics.AddError("Failed to update API key", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *APIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state APIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.do("DELETE", fmt.Sprintf("/api-keys/%d", state.ID.ValueInt64()), nil, nil); err != nil {
		resp.Diagnostics.AddError("Failed to delete API key", err.Error())
		return
	}
}

