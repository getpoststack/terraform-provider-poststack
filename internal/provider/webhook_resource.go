package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &WebhookResource{}

type WebhookResource struct {
	client *Client
}

type WebhookResourceModel struct {
	ID     types.Int64    `tfsdk:"id"`
	URL    types.String   `tfsdk:"url"`
	Events []types.String `tfsdk:"events"`
	Secret types.String   `tfsdk:"secret"`
}

func NewWebhookResource() resource.Resource {
	return &WebhookResource{}
}

func (r *WebhookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *WebhookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A PostStack webhook subscription. Each event in `events` triggers a POST to `url` with an HMAC signature in the X-PostStack-Signature header.",
		Attributes: map[string]schema.Attribute{
			"id":  schema.Int64Attribute{Computed: true},
			"url": schema.StringAttribute{Required: true},
			"events": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Event names to subscribe to (e.g. email.delivered, email.bounced, contact.created).",
			},
			"secret": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "HMAC signing secret. Returned only at create time.",
			},
		},
	}
}

func (r *WebhookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client, _ = req.ProviderData.(*Client)
}

type webhookAPI struct {
	ID     int64    `json:"id"`
	URL    string   `json:"url"`
	Events []string `json:"events"`
	Secret string   `json:"secret,omitempty"`
}

func (r *WebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WebhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	events := make([]string, 0, len(plan.Events))
	for _, e := range plan.Events {
		events = append(events, e.ValueString())
	}
	body := map[string]interface{}{
		"url":    plan.URL.ValueString(),
		"events": events,
	}
	var created webhookAPI
	if err := r.client.do("POST", "/webhooks", body, &created); err != nil {
		resp.Diagnostics.AddError("Failed to create webhook", err.Error())
		return
	}
	plan.ID = types.Int64Value(created.ID)
	plan.Secret = types.StringValue(created.Secret)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *WebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WebhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var w webhookAPI
	if err := r.client.do("GET", fmt.Sprintf("/webhooks/%d", state.ID.ValueInt64()), nil, &w); err != nil {
		resp.Diagnostics.AddError("Failed to read webhook", err.Error())
		return
	}
	state.URL = types.StringValue(w.URL)
	events := make([]types.String, 0, len(w.Events))
	for _, e := range w.Events {
		events = append(events, types.StringValue(e))
	}
	state.Events = events
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *WebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WebhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	events := make([]string, 0, len(plan.Events))
	for _, e := range plan.Events {
		events = append(events, e.ValueString())
	}
	body := map[string]interface{}{
		"url":    plan.URL.ValueString(),
		"events": events,
	}
	if err := r.client.do("PATCH", fmt.Sprintf("/webhooks/%d", plan.ID.ValueInt64()), body, nil); err != nil {
		resp.Diagnostics.AddError("Failed to update webhook", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *WebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WebhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.do("DELETE", fmt.Sprintf("/webhooks/%d", state.ID.ValueInt64()), nil, nil); err != nil {
		resp.Diagnostics.AddError("Failed to delete webhook", err.Error())
		return
	}
}
