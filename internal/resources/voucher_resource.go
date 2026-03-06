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
	_ resource.Resource                = &VoucherResource{}
	_ resource.ResourceWithImportState = &VoucherResource{}
)

type VoucherResource struct {
	client *client.Client
}

type VoucherResourceModel struct {
	ID                types.String `tfsdk:"id"`
	TemplateName      types.String `tfsdk:"template_name"`
	VoucherFormat     types.String `tfsdk:"voucher_format"`
	VoucherLength     types.Int64  `tfsdk:"voucher_length"`
	VoucherQuantity   types.Int64  `tfsdk:"voucher_quantity"`
	GroupID           types.String `tfsdk:"group_id"`
	OrgunitID         types.String `tfsdk:"orgunit_id"`
	VoucherDeleteDate types.String `tfsdk:"voucher_deletedate"`
	VoucherDevices    types.Int64  `tfsdk:"voucher_devices"`
	VoucherDuration   types.String `tfsdk:"voucher_duration"`
}

func NewVoucherResource() resource.Resource {
	return &VoucherResource{}
}

func (r *VoucherResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_voucher"
}

func (r *VoucherResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an IronWiFi voucher template.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Voucher ID (UUID).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"template_name": schema.StringAttribute{
				Description: "Voucher template name.",
				Required:    true,
			},
			"voucher_format": schema.StringAttribute{
				Description: "Voucher code format.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"voucher_length": schema.Int64Attribute{
				Description: "Length of voucher codes.",
				Optional:    true,
				Computed:    true,
			},
			"voucher_quantity": schema.Int64Attribute{
				Description: "Number of vouchers to generate.",
				Optional:    true,
				Computed:    true,
			},
			"group_id": schema.StringAttribute{
				Description: "Group ID to assign voucher users to.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"orgunit_id": schema.StringAttribute{
				Description: "Organizational unit ID.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"voucher_deletedate": schema.StringAttribute{
				Description: "Voucher expiration/delete date.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"voucher_devices": schema.Int64Attribute{
				Description: "Number of devices allowed per voucher.",
				Optional:    true,
				Computed:    true,
			},
			"voucher_duration": schema.StringAttribute{
				Description: "Voucher duration.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
		},
	}
}

func (r *VoucherResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VoucherResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VoucherResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"template_name": plan.TemplateName.ValueString(),
	}
	setIfNotNull(body, "voucher_format", plan.VoucherFormat)
	setIntIfNotNull(body, "voucher_length", plan.VoucherLength)
	setIntIfNotNull(body, "voucher_quantity", plan.VoucherQuantity)
	setIfNotNull(body, "group_id", plan.GroupID)
	setIfNotNull(body, "orgunitid", plan.OrgunitID)
	setIfNotNull(body, "voucher_deletedate", plan.VoucherDeleteDate)
	setIntIfNotNull(body, "voucher_devices", plan.VoucherDevices)
	setIfNotNull(body, "voucher_duration", plan.VoucherDuration)

	result, err := r.client.Create("vouchers", body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating voucher", err.Error())
		return
	}

	mapVoucherResponse(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *VoucherResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VoucherResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Read("vouchers", state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading voucher", err.Error())
		return
	}

	mapVoucherResponse(result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *VoucherResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan VoucherResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state VoucherResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"template_name": plan.TemplateName.ValueString(),
	}
	setIfNotNull(body, "voucher_format", plan.VoucherFormat)
	setIntIfNotNull(body, "voucher_length", plan.VoucherLength)
	setIntIfNotNull(body, "voucher_quantity", plan.VoucherQuantity)
	setIfNotNull(body, "group_id", plan.GroupID)
	setIfNotNull(body, "orgunitid", plan.OrgunitID)
	setIfNotNull(body, "voucher_deletedate", plan.VoucherDeleteDate)
	setIntIfNotNull(body, "voucher_devices", plan.VoucherDevices)
	setIfNotNull(body, "voucher_duration", plan.VoucherDuration)

	result, err := r.client.Update("vouchers", state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating voucher", err.Error())
		return
	}

	mapVoucherResponse(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *VoucherResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VoucherResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete("vouchers", state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting voucher", err.Error())
	}
}

func (r *VoucherResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapVoucherResponse(data map[string]interface{}, model *VoucherResourceModel) {
	model.ID = stringFromAPI(data, "id")
	model.TemplateName = stringFromAPI(data, "template_name")
	model.VoucherFormat = stringFromAPI(data, "voucher_format")
	model.VoucherLength = intFromAPINullable(data, "voucher_length")
	model.VoucherQuantity = intFromAPINullable(data, "voucher_quantity")
	model.GroupID = stringFromAPI(data, "group_id")
	model.OrgunitID = stringFromAPI(data, "orgunitid")
	model.VoucherDeleteDate = stringFromAPI(data, "voucher_deletedate")
	model.VoucherDevices = intFromAPINullable(data, "voucher_devices")
	model.VoucherDuration = stringFromAPI(data, "voucher_duration")
}
