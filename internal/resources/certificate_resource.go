package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ironwifi/terraform-provider-ironwifi/internal/client"
)

var (
	_ resource.Resource                = &CertificateResource{}
	_ resource.ResourceWithImportState = &CertificateResource{}
)

type CertificateResource struct {
	client *client.Client
}

type CertificateResourceModel struct {
	ID             types.String `tfsdk:"id"`
	UserID         types.String `tfsdk:"user_id"`
	Serial         types.String `tfsdk:"serial"`
	Status         types.String `tfsdk:"status"`
	CN             types.String `tfsdk:"cn"`
	Subject        types.String `tfsdk:"subject"`
	Validity       types.Int64  `tfsdk:"validity"`
	Distribution   types.String `tfsdk:"distribution"`
	Hash           types.String `tfsdk:"hash"`
	ExpirationDate types.String `tfsdk:"expirationdate"`
	RevocationDate types.String `tfsdk:"revocationdate"`
	CreationDate   types.String `tfsdk:"creationdate"`
}

func NewCertificateResource() resource.Resource {
	return &CertificateResource{}
}

func (r *CertificateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

func (r *CertificateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an IronWiFi certificate.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Certificate ID (UUID).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				Description: "UUID of the user this certificate belongs to.",
				Required:    true,
			},
			"serial": schema.StringAttribute{
				Description: "Certificate serial number.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Description: "Certificate status: valid, revoked, or pending.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("valid"),
			},
			"cn": schema.StringAttribute{
				Description: "Common name.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"subject": schema.StringAttribute{
				Description: "Certificate subject.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"validity": schema.Int64Attribute{
				Description: "Validity period in days.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(365),
			},
			"distribution": schema.StringAttribute{
				Description: "Distribution method.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("email"),
			},
			"hash": schema.StringAttribute{
				Description: "Hash algorithm.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("sha2"),
			},
			"expirationdate": schema.StringAttribute{
				Description: "Expiration date.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"revocationdate": schema.StringAttribute{
				Description: "Revocation date.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

func (r *CertificateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CertificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"user_id": plan.UserID.ValueString(),
	}
	setIfNotNull(body, "status", plan.Status)
	setIfNotNull(body, "cn", plan.CN)
	setIfNotNull(body, "subject", plan.Subject)
	setIntIfNotNull(body, "validity", plan.Validity)
	setIfNotNull(body, "distribution", plan.Distribution)
	setIfNotNull(body, "hash", plan.Hash)

	result, err := r.client.Create("certificates", body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating certificate", err.Error())
		return
	}

	mapCertificateResponse(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Read("certificates", state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading certificate", err.Error())
		return
	}

	mapCertificateResponse(result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CertificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state CertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"user_id": plan.UserID.ValueString(),
	}
	setIfNotNull(body, "status", plan.Status)
	setIfNotNull(body, "cn", plan.CN)
	setIfNotNull(body, "subject", plan.Subject)
	setIntIfNotNull(body, "validity", plan.Validity)
	setIfNotNull(body, "distribution", plan.Distribution)
	setIfNotNull(body, "hash", plan.Hash)

	result, err := r.client.Update("certificates", state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating certificate", err.Error())
		return
	}

	mapCertificateResponse(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete("certificates", state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting certificate", err.Error())
	}
}

func (r *CertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapCertificateResponse(data map[string]interface{}, model *CertificateResourceModel) {
	model.ID = stringFromAPI(data, "id")
	model.UserID = stringFromAPI(data, "user_id")
	model.Serial = stringFromAPI(data, "serial")
	model.Status = stringFromAPI(data, "status")
	model.CN = stringFromAPI(data, "cn")
	model.Subject = stringFromAPI(data, "subject")
	model.Validity = intFromAPI(data, "validity")
	model.Distribution = stringFromAPI(data, "distribution")
	model.Hash = stringFromAPI(data, "hash")
	model.ExpirationDate = stringFromAPI(data, "expirationdate")
	model.RevocationDate = stringFromAPI(data, "revocationdate")
	model.CreationDate = stringFromAPI(data, "creationdate")
}
