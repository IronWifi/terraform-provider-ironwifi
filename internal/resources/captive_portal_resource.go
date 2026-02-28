package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ironwifi/terraform-provider-ironwifi/internal/client"
)

var (
	_ resource.Resource                = &CaptivePortalResource{}
	_ resource.ResourceWithImportState = &CaptivePortalResource{}
)

type CaptivePortalResource struct {
	client *client.Client
}

type CaptivePortalResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	Vendor            types.String `tfsdk:"vendor"`
	NetworkID         types.String `tfsdk:"network_id"`
	SplashPage        types.String `tfsdk:"splash_page"`
	SuccessPage       types.String `tfsdk:"success_page"`
	PortalTheme       types.String `tfsdk:"portal_theme"`
	MacAuthentication types.Bool   `tfsdk:"mac_authentication"`
	CloudCDN          types.Bool   `tfsdk:"cloud_cdn"`
	WebhookURL        types.String `tfsdk:"webhook_url"`
}

func NewCaptivePortalResource() resource.Resource {
	return &CaptivePortalResource{}
}

func (r *CaptivePortalResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_captive_portal"
}

func (r *CaptivePortalResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an IronWiFi captive portal.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Captive portal ID (UUID).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Captive portal name.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Captive portal description.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"vendor": schema.StringAttribute{
				Description: "Hardware vendor for the captive portal.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"network_id": schema.StringAttribute{
				Description: "Associated network ID.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"splash_page": schema.StringAttribute{
				Description: "Splash page URL or template.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"success_page": schema.StringAttribute{
				Description: "Success/redirect page URL.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"portal_theme": schema.StringAttribute{
				Description: "Portal theme identifier.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"mac_authentication": schema.BoolAttribute{
				Description: "Enable MAC-based authentication.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"cloud_cdn": schema.BoolAttribute{
				Description: "Enable Cloud CDN for portal assets.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"webhook_url": schema.StringAttribute{
				Description: "Webhook URL for portal events.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
		},
	}
}

func (r *CaptivePortalResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CaptivePortalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CaptivePortalResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"name": plan.Name.ValueString(),
	}
	setIfNotNull(body, "description", plan.Description)
	setIfNotNull(body, "vendor", plan.Vendor)
	setIfNotNull(body, "network_id", plan.NetworkID)
	setIfNotNull(body, "splash_page", plan.SplashPage)
	setIfNotNull(body, "success_page", plan.SuccessPage)
	setIfNotNull(body, "portal_theme", plan.PortalTheme)
	setBoolAsInt(body, "mac_authentication", plan.MacAuthentication)
	setBoolAsInt(body, "cloud_cdn", plan.CloudCDN)
	setIfNotNull(body, "webhook_url", plan.WebhookURL)

	result, err := r.client.Create("captive-portals", body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating captive portal", err.Error())
		return
	}

	mapCaptivePortalResponse(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CaptivePortalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CaptivePortalResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Read("captive-portals", state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading captive portal", err.Error())
		return
	}

	mapCaptivePortalResponse(result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CaptivePortalResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CaptivePortalResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state CaptivePortalResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"name": plan.Name.ValueString(),
	}
	setIfNotNull(body, "description", plan.Description)
	setIfNotNull(body, "vendor", plan.Vendor)
	setIfNotNull(body, "network_id", plan.NetworkID)
	setIfNotNull(body, "splash_page", plan.SplashPage)
	setIfNotNull(body, "success_page", plan.SuccessPage)
	setIfNotNull(body, "portal_theme", plan.PortalTheme)
	setBoolAsInt(body, "mac_authentication", plan.MacAuthentication)
	setBoolAsInt(body, "cloud_cdn", plan.CloudCDN)
	setIfNotNull(body, "webhook_url", plan.WebhookURL)

	result, err := r.client.Update("captive-portals", state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating captive portal", err.Error())
		return
	}

	mapCaptivePortalResponse(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CaptivePortalResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CaptivePortalResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete("captive-portals", state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting captive portal", err.Error())
	}
}

func (r *CaptivePortalResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapCaptivePortalResponse(data map[string]interface{}, model *CaptivePortalResourceModel) {
	model.ID = stringFromAPI(data, "id")
	model.Name = stringFromAPI(data, "name")
	model.Description = stringFromAPI(data, "description")
	model.Vendor = stringFromAPI(data, "vendor")
	model.NetworkID = stringFromAPI(data, "network_id")
	model.SplashPage = stringFromAPI(data, "splash_page")
	model.SuccessPage = stringFromAPI(data, "success_page")
	model.PortalTheme = stringFromAPI(data, "portal_theme")
	model.MacAuthentication = boolFromIntAPI(data, "mac_authentication")
	model.CloudCDN = boolFromIntAPI(data, "cloud_cdn")
	model.WebhookURL = stringFromAPI(data, "webhook_url")
}
