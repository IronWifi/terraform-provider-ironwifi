package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ironwifi/terraform-provider-ironwifi/internal/client"
)

var (
	_ resource.Resource                = &PolicyResource{}
	_ resource.ResourceWithImportState = &PolicyResource{}
)

type PolicyResource struct {
	client *client.Client
}

type PolicyResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Priority    types.Int64  `tfsdk:"priority"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	MatchMode   types.String `tfsdk:"match_mode"`
	TargetType  types.String `tfsdk:"target_type"`
	TargetID    types.String `tfsdk:"target_id"`
	Conditions  types.String `tfsdk:"conditions"`
	Actions     types.String `tfsdk:"actions"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func NewPolicyResource() resource.Resource {
	return &PolicyResource{}
}

func (r *PolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

func (r *PolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an IronWiFi policy (conditional access).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Policy ID (UUID).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Policy name.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Policy description.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"priority": schema.Int64Attribute{
				Description: "Evaluation priority (lower values evaluated first).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(100),
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the policy is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"match_mode": schema.StringAttribute{
				Description: "Condition matching mode: 'all' or 'any'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("all"),
			},
			"target_type": schema.StringAttribute{
				Description: "Target type: 'global', 'network', 'group', etc.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("global"),
			},
			"target_id": schema.StringAttribute{
				Description: "Target entity ID (when target_type is not 'global').",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"conditions": schema.StringAttribute{
				Description: "JSON string defining policy conditions.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"actions": schema.StringAttribute{
				Description: "JSON string defining policy actions.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"created_at": schema.StringAttribute{
				Description: "Date the policy was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "Date the policy was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *PolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *PolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"name": plan.Name.ValueString(),
	}
	setIfNotNull(body, "description", plan.Description)
	setIntIfNotNull(body, "priority", plan.Priority)
	setBoolAsInt(body, "enabled", plan.Enabled)
	setIfNotNull(body, "match_mode", plan.MatchMode)
	setIfNotNull(body, "target_type", plan.TargetType)
	setIfNotNull(body, "target_id", plan.TargetID)
	setIfNotNull(body, "conditions", plan.Conditions)
	setIfNotNull(body, "actions", plan.Actions)

	result, err := r.client.Create("policies", body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating policy", err.Error())
		return
	}

	mapPolicyResponse(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Read("policies", state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading policy", err.Error())
		return
	}

	mapPolicyResponse(result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"name": plan.Name.ValueString(),
	}
	setIfNotNull(body, "description", plan.Description)
	setIntIfNotNull(body, "priority", plan.Priority)
	setBoolAsInt(body, "enabled", plan.Enabled)
	setIfNotNull(body, "match_mode", plan.MatchMode)
	setIfNotNull(body, "target_type", plan.TargetType)
	setIfNotNull(body, "target_id", plan.TargetID)
	setIfNotNull(body, "conditions", plan.Conditions)
	setIfNotNull(body, "actions", plan.Actions)

	result, err := r.client.Update("policies", state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating policy", err.Error())
		return
	}

	mapPolicyResponse(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete("policies", state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting policy", err.Error())
	}
}

func (r *PolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapPolicyResponse(data map[string]interface{}, model *PolicyResourceModel) {
	model.ID = stringFromAPI(data, "id")
	model.Name = stringFromAPI(data, "name")
	model.Description = stringFromAPI(data, "description")
	model.Priority = intFromAPI(data, "priority")
	model.Enabled = boolFromIntAPI(data, "enabled")
	model.MatchMode = stringFromAPI(data, "match_mode")
	model.TargetType = stringFromAPI(data, "target_type")
	model.TargetID = stringFromAPI(data, "target_id")
	model.Conditions = stringFromAPI(data, "conditions")
	model.Actions = stringFromAPI(data, "actions")
	model.CreatedAt = stringFromAPI(data, "created_at")
	model.UpdatedAt = stringFromAPI(data, "updated_at")
}
