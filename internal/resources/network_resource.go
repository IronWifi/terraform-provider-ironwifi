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
	_ resource.Resource                = &NetworkResource{}
	_ resource.ResourceWithImportState = &NetworkResource{}
)

type NetworkResource struct {
	client *client.Client
}

type NetworkResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Region       types.String `tfsdk:"region"`
	AuthPort     types.Int64  `tfsdk:"auth_port"`
	AcctPort     types.Int64  `tfsdk:"acct_port"`
	Secret       types.String `tfsdk:"secret"`
	PrimaryIP    types.String `tfsdk:"primary_ip"`
	BackupIP     types.String `tfsdk:"backup_ip"`
	IPv6         types.Bool   `tfsdk:"ipv6"`
	UnknownUsers types.String `tfsdk:"unknown_users"`
	OpenRoaming  types.Bool   `tfsdk:"open_roaming"`
	Eduroam      types.Bool   `tfsdk:"eduroam"`
	COA          types.Bool   `tfsdk:"coa"`
	RadSec       types.Bool   `tfsdk:"radsec"`
}

func NewNetworkResource() resource.Resource {
	return &NetworkResource{}
}

func (r *NetworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (r *NetworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an IronWiFi network (RADIUS NAS).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Network ID (UUID).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Network name (NAS identifier).",
				Required:    true,
			},
			"region": schema.StringAttribute{
				Description: "RADIUS region for this network.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"auth_port": schema.Int64Attribute{
				Description: "RADIUS authentication port.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1812),
			},
			"acct_port": schema.Int64Attribute{
				Description: "RADIUS accounting port.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1813),
			},
			"secret": schema.StringAttribute{
				Description: "RADIUS shared secret.",
				Computed:    true,
				Sensitive:   true,
			},
			"primary_ip": schema.StringAttribute{
				Description: "Primary RADIUS server IP.",
				Computed:    true,
			},
			"backup_ip": schema.StringAttribute{
				Description: "Backup RADIUS server IP.",
				Computed:    true,
			},
			"ipv6": schema.BoolAttribute{
				Description: "Enable IPv6 support.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"unknown_users": schema.StringAttribute{
				Description: "Action for unknown users: 'reject' or 'accept'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("reject"),
			},
			"open_roaming": schema.BoolAttribute{
				Description: "Enable OpenRoaming.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"eduroam": schema.BoolAttribute{
				Description: "Enable eduroam federation.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"coa": schema.BoolAttribute{
				Description: "Enable Change of Authorization (CoA).",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"radsec": schema.BoolAttribute{
				Description: "Enable RadSec (RADIUS over TLS).",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *NetworkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NetworkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"nasname": plan.Name.ValueString(),
	}
	setIfNotNull(body, "region", plan.Region)
	setIntIfNotNull(body, "auth_port", plan.AuthPort)
	setIntIfNotNull(body, "acct_port", plan.AcctPort)
	setBoolAsInt(body, "ipv6", plan.IPv6)
	setIfNotNull(body, "unknown_users", plan.UnknownUsers)
	setBoolAsInt(body, "open_roaming", plan.OpenRoaming)
	setBoolAsInt(body, "eduroam", plan.Eduroam)
	setBoolAsInt(body, "coa", plan.COA)
	setBoolAsInt(body, "radsec", plan.RadSec)

	result, err := r.client.Create("networks", body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating network", err.Error())
		return
	}

	mapNetworkResponse(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NetworkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Read("networks", state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading network", err.Error())
		return
	}

	mapNetworkResponse(result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NetworkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state NetworkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"nasname": plan.Name.ValueString(),
	}
	setIfNotNull(body, "region", plan.Region)
	setBoolAsInt(body, "ipv6", plan.IPv6)
	setIfNotNull(body, "unknown_users", plan.UnknownUsers)
	setBoolAsInt(body, "open_roaming", plan.OpenRoaming)
	setBoolAsInt(body, "eduroam", plan.Eduroam)
	setBoolAsInt(body, "coa", plan.COA)
	setBoolAsInt(body, "radsec", plan.RadSec)

	result, err := r.client.Update("networks", state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating network", err.Error())
		return
	}

	mapNetworkResponse(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NetworkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete("networks", state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting network", err.Error())
	}
}

func (r *NetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapNetworkResponse(data map[string]interface{}, model *NetworkResourceModel) {
	model.ID = stringFromAPI(data, "id")
	model.Name = stringFromAPI(data, "nasname")
	model.Region = stringFromAPI(data, "region")
	model.AuthPort = intFromAPI(data, "auth_port")
	model.AcctPort = intFromAPI(data, "acct_port")
	model.Secret = stringFromAPI(data, "secret")
	model.PrimaryIP = stringFromAPI(data, "primary_ip")
	model.BackupIP = stringFromAPI(data, "backup_ip")
	model.IPv6 = boolFromIntAPI(data, "ipv6")
	model.UnknownUsers = stringFromAPI(data, "unknown_users")
	model.OpenRoaming = boolFromIntAPI(data, "open_roaming")
	model.Eduroam = boolFromIntAPI(data, "eduroam")
	model.COA = boolFromIntAPI(data, "coa")
	model.RadSec = boolFromIntAPI(data, "radsec")
}
