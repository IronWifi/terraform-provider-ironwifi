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
	_ resource.Resource                = &DeviceResource{}
	_ resource.ResourceWithImportState = &DeviceResource{}
)

type DeviceResource struct {
	client *client.Client
}

type DeviceResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Email        types.String `tfsdk:"email"`
	Firstname    types.String `tfsdk:"firstname"`
	Lastname     types.String `tfsdk:"lastname"`
	Notes        types.String `tfsdk:"notes"`
	Mobilephone  types.String `tfsdk:"mobilephone"`
	Authsource   types.String `tfsdk:"authsource"`
	Orgunit      types.String `tfsdk:"orgunit"`
	Status       types.String `tfsdk:"status"`
	CreationDate types.String `tfsdk:"creationdate"`
}

func NewDeviceResource() resource.Resource {
	return &DeviceResource{}
}

func (r *DeviceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device"
}

func (r *DeviceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an IronWiFi device.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Device ID (UUID).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Device identifier (maps to API username).",
				Required:    true,
			},
			"email": schema.StringAttribute{
				Description: "Email address associated with the device.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"firstname": schema.StringAttribute{
				Description: "First name of the device owner.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"lastname": schema.StringAttribute{
				Description: "Last name of the device owner.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"notes": schema.StringAttribute{
				Description: "Notes about the device.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"mobilephone": schema.StringAttribute{
				Description: "Mobile phone number.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"authsource": schema.StringAttribute{
				Description: "Authentication source.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("local"),
			},
			"orgunit": schema.StringAttribute{
				Description: "Organizational unit.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"status": schema.StringAttribute{
				Description: "Device status.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
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

func (r *DeviceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DeviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DeviceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"username": plan.Name.ValueString(),
	}
	setIfNotNull(body, "email", plan.Email)
	setIfNotNull(body, "firstname", plan.Firstname)
	setIfNotNull(body, "lastname", plan.Lastname)
	setIfNotNull(body, "notes", plan.Notes)
	setIfNotNull(body, "mobilephone", plan.Mobilephone)
	setIfNotNull(body, "authsource", plan.Authsource)
	setIfNotNull(body, "orgunit", plan.Orgunit)
	setIfNotNull(body, "status", plan.Status)

	result, err := r.client.Create("devices", body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating device", err.Error())
		return
	}

	mapDeviceResponse(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DeviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DeviceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Read("devices", state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading device", err.Error())
		return
	}

	mapDeviceResponse(result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DeviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DeviceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state DeviceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"username": plan.Name.ValueString(),
	}
	setIfNotNull(body, "email", plan.Email)
	setIfNotNull(body, "firstname", plan.Firstname)
	setIfNotNull(body, "lastname", plan.Lastname)
	setIfNotNull(body, "notes", plan.Notes)
	setIfNotNull(body, "mobilephone", plan.Mobilephone)
	setIfNotNull(body, "authsource", plan.Authsource)
	setIfNotNull(body, "orgunit", plan.Orgunit)
	setIfNotNull(body, "status", plan.Status)

	result, err := r.client.Update("devices", state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating device", err.Error())
		return
	}

	mapDeviceResponse(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DeviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DeviceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete("devices", state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting device", err.Error())
	}
}

func (r *DeviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapDeviceResponse(data map[string]interface{}, model *DeviceResourceModel) {
	model.ID = stringFromAPI(data, "id")
	model.Name = stringFromAPI(data, "username")
	model.Email = stringFromAPI(data, "email")
	model.Firstname = stringFromAPI(data, "firstname")
	model.Lastname = stringFromAPI(data, "lastname")
	model.Notes = stringFromAPI(data, "notes")
	model.Mobilephone = stringFromAPI(data, "mobilephone")
	model.Authsource = stringFromAPI(data, "authsource")
	model.Orgunit = stringFromAPI(data, "orgunit")
	model.Status = stringFromAPI(data, "status")
	model.CreationDate = stringFromAPI(data, "creationdate")
}
