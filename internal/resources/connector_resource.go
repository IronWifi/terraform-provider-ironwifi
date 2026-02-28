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
	_ resource.Resource                = &ConnectorResource{}
	_ resource.ResourceWithImportState = &ConnectorResource{}
)

type ConnectorResource struct {
	client *client.Client
}

type ConnectorResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	Domain       types.String `tfsdk:"domain"`
	Group        types.String `tfsdk:"group"`
	Groupname    types.String `tfsdk:"groupname"`
	Status       types.String `tfsdk:"status"`
	Authsource   types.String `tfsdk:"authsource"`
	BaseDN       types.String `tfsdk:"basedn"`
	Bind         types.String `tfsdk:"bind"`
	Password     types.String `tfsdk:"password"`
	SyncInterval types.Int64  `tfsdk:"sync_interval"`
	UserTakeover types.Bool   `tfsdk:"user_takeover"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	CreationDate types.String `tfsdk:"creationdate"`
}

func NewConnectorResource() resource.Resource {
	return &ConnectorResource{}
}

func (r *ConnectorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connector"
}

func (r *ConnectorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an IronWiFi authentication connector (LDAP, AD, SAML, OAuth).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Connector ID (UUID).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Connector name.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Connector type: ldap, ad, saml, or oauth.",
				Required:    true,
			},
			"domain": schema.StringAttribute{
				Description: "Domain for the connector.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"group": schema.StringAttribute{
				Description: "Group identifier.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"groupname": schema.StringAttribute{
				Description: "Group name.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"status": schema.StringAttribute{
				Description: "Connector status.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("enabled"),
			},
			"authsource": schema.StringAttribute{
				Description: "Authentication source.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"basedn": schema.StringAttribute{
				Description: "Base DN for LDAP.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"bind": schema.StringAttribute{
				Description: "Bind DN for LDAP.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"password": schema.StringAttribute{
				Description: "Password (maps to API dbpassword).",
				Optional:    true,
				Sensitive:   true,
			},
			"sync_interval": schema.Int64Attribute{
				Description: "Sync interval in minutes.",
				Optional:    true,
				Computed:    true,
			},
			"user_takeover": schema.BoolAttribute{
				Description: "Enable user takeover.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"client_id": schema.StringAttribute{
				Description: "Client ID for OAuth/SAML.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"client_secret": schema.StringAttribute{
				Description: "Client secret for OAuth/SAML.",
				Optional:    true,
				Sensitive:   true,
			},
			"creationdate": schema.StringAttribute{
				Description: "Creation date.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ConnectorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConnectorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ConnectorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"name":   plan.Name.ValueString(),
		"dbtype": plan.Type.ValueString(),
	}
	setIfNotNull(body, "domain", plan.Domain)
	setIfNotNull(body, "group", plan.Group)
	setIfNotNull(body, "groupname", plan.Groupname)
	setIfNotNull(body, "status", plan.Status)
	setIfNotNull(body, "authsource", plan.Authsource)
	setIfNotNull(body, "basedn", plan.BaseDN)
	setIfNotNull(body, "bind", plan.Bind)
	setIfNotNull(body, "dbpassword", plan.Password)
	setIntIfNotNull(body, "sync_interval", plan.SyncInterval)
	setBoolAsInt(body, "user_takeover", plan.UserTakeover)
	setIfNotNull(body, "client_id", plan.ClientID)
	setIfNotNull(body, "client_secret", plan.ClientSecret)

	result, err := r.client.Create("connectors", body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating connector", err.Error())
		return
	}

	mapConnectorResponse(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ConnectorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ConnectorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Read("connectors", state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading connector", err.Error())
		return
	}

	// Preserve write-only fields since the API does not return them.
	pw := state.Password
	cs := state.ClientSecret
	mapConnectorResponse(result, &state)
	state.Password = pw
	state.ClientSecret = cs
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ConnectorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ConnectorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ConnectorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"name":   plan.Name.ValueString(),
		"dbtype": plan.Type.ValueString(),
	}
	setIfNotNull(body, "domain", plan.Domain)
	setIfNotNull(body, "group", plan.Group)
	setIfNotNull(body, "groupname", plan.Groupname)
	setIfNotNull(body, "status", plan.Status)
	setIfNotNull(body, "authsource", plan.Authsource)
	setIfNotNull(body, "basedn", plan.BaseDN)
	setIfNotNull(body, "bind", plan.Bind)
	setIfNotNull(body, "dbpassword", plan.Password)
	setIntIfNotNull(body, "sync_interval", plan.SyncInterval)
	setBoolAsInt(body, "user_takeover", plan.UserTakeover)
	setIfNotNull(body, "client_id", plan.ClientID)
	setIfNotNull(body, "client_secret", plan.ClientSecret)

	result, err := r.client.Update("connectors", state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating connector", err.Error())
		return
	}

	mapConnectorResponse(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ConnectorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ConnectorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete("connectors", state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting connector", err.Error())
	}
}

func (r *ConnectorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapConnectorResponse(data map[string]interface{}, model *ConnectorResourceModel) {
	model.ID = stringFromAPI(data, "id")
	model.Name = stringFromAPI(data, "name")
	model.Type = stringFromAPI(data, "dbtype")
	model.Domain = stringFromAPI(data, "domain")
	model.Group = stringFromAPI(data, "group")
	model.Groupname = stringFromAPI(data, "groupname")
	model.Status = stringFromAPI(data, "status")
	model.Authsource = stringFromAPI(data, "authsource")
	model.BaseDN = stringFromAPI(data, "basedn")
	model.Bind = stringFromAPI(data, "bind")
	// Password is write-only; not read back from API
	model.SyncInterval = intFromAPINullable(data, "sync_interval")
	model.UserTakeover = boolFromIntAPI(data, "user_takeover")
	model.ClientID = stringFromAPI(data, "client_id")
	// ClientSecret is write-only; not read back from API
	model.CreationDate = stringFromAPI(data, "creationdate")
}
