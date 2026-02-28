package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ironwifi/terraform-provider-ironwifi/internal/client"
)

var (
	_ resource.Resource                = &AuthProviderResource{}
	_ resource.ResourceWithImportState = &AuthProviderResource{}
)

type AuthProviderResource struct {
	client *client.Client
}

type AuthProviderResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Type            types.String `tfsdk:"type"`
	CaptivePortalID types.String `tfsdk:"captive_portal_id"`
	GroupID         types.String `tfsdk:"group_id"`
	Status          types.String `tfsdk:"status"`
	Configuration   types.String `tfsdk:"configuration"`
}

func NewAuthProviderResource() resource.Resource {
	return &AuthProviderResource{}
}

func (r *AuthProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_authentication_provider"
}

func (r *AuthProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an IronWiFi authentication provider.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Authentication provider ID (UUID).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Authentication provider name.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Provider type: ldap, saml, social, twilio, etc.",
				Required:    true,
			},
			"captive_portal_id": schema.StringAttribute{
				Description: "Associated captive portal ID.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"group_id": schema.StringAttribute{
				Description: "Associated group ID.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"status": schema.StringAttribute{
				Description: "Provider status (enabled/disabled).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("enabled"),
			},
			"configuration": schema.StringAttribute{
				Description: "JSON string with type-specific configuration.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
		},
	}
}

func (r *AuthProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AuthProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AuthProviderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"name": plan.Name.ValueString(),
		"type": plan.Type.ValueString(),
	}
	setIfNotNull(body, "captive_portal_id", plan.CaptivePortalID)
	setIfNotNull(body, "group_id", plan.GroupID)
	setIfNotNull(body, "status", plan.Status)
	setIfNotNull(body, "configuration", plan.Configuration)

	result, err := r.client.Create("authentication-providers", body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating authentication provider", err.Error())
		return
	}

	mapAuthProviderResponse(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AuthProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AuthProviderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Read("authentication-providers", state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading authentication provider", err.Error())
		return
	}

	mapAuthProviderResponse(result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AuthProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AuthProviderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state AuthProviderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"name": plan.Name.ValueString(),
		"type": plan.Type.ValueString(),
	}
	setIfNotNull(body, "captive_portal_id", plan.CaptivePortalID)
	setIfNotNull(body, "group_id", plan.GroupID)
	setIfNotNull(body, "status", plan.Status)
	setIfNotNull(body, "configuration", plan.Configuration)

	result, err := r.client.Update("authentication-providers", state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating authentication provider", err.Error())
		return
	}

	mapAuthProviderResponse(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AuthProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AuthProviderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete("authentication-providers", state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting authentication provider", err.Error())
	}
}

func (r *AuthProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapAuthProviderResponse(data map[string]interface{}, model *AuthProviderResourceModel) {
	model.ID = stringFromAPI(data, "id")
	model.Name = stringFromAPI(data, "name")
	model.Type = stringFromAPI(data, "type")
	model.CaptivePortalID = stringFromAPI(data, "captive_portal_id")
	model.GroupID = stringFromAPI(data, "group_id")
	model.Status = stringFromAPI(data, "status")
	model.Configuration = stringFromAPI(data, "configuration")
}
