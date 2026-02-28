package datasources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ironwifi/terraform-provider-ironwifi/internal/client"
)

var _ datasource.DataSource = &PoliciesDataSource{}

// PoliciesDataSource lists IronWiFi conditional access policies.
type PoliciesDataSource struct {
	client *client.Client
}

// PoliciesDataSourceModel is the schema model for the policies data source.
type PoliciesDataSourceModel struct {
	NameFilter types.String      `tfsdk:"name_filter"`
	Items      []PolicyItemModel `tfsdk:"items"`
}

// PolicyItemModel represents a single policy in the data source output.
type PolicyItemModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Priority    types.Int64  `tfsdk:"priority"`
	Enabled     types.Bool   `tfsdk:"enabled"`
}

// NewPoliciesDataSource returns a new policies data source.
func NewPoliciesDataSource() datasource.DataSource {
	return &PoliciesDataSource{}
}

func (d *PoliciesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policies"
}

func (d *PoliciesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists IronWiFi conditional access policies.",
		Attributes: map[string]schema.Attribute{
			"name_filter": schema.StringAttribute{
				Description: "Optional substring filter applied to policy names.",
				Optional:    true,
			},
			"items": schema.ListNestedAttribute{
				Description: "The list of policies.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true, Description: "Policy ID."},
						"name":        schema.StringAttribute{Computed: true, Description: "Policy name."},
						"description": schema.StringAttribute{Computed: true, Description: "Policy description."},
						"priority":    schema.Int64Attribute{Computed: true, Description: "Policy priority."},
						"enabled":     schema.BoolAttribute{Computed: true, Description: "Whether the policy is enabled."},
					},
				},
			},
		},
	}
}

func (d *PoliciesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *PoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state PoliciesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, err := d.client.List("policies", "conditional_access_policies")
	if err != nil {
		resp.Diagnostics.AddError("Error listing policies", err.Error())
		return
	}

	for _, item := range items {
		name := fmt.Sprintf("%v", item["name"])
		if !state.NameFilter.IsNull() && !state.NameFilter.IsUnknown() {
			if !strings.Contains(strings.ToLower(name), strings.ToLower(state.NameFilter.ValueString())) {
				continue
			}
		}
		state.Items = append(state.Items, PolicyItemModel{
			ID:          stringVal(item, "id"),
			Name:        types.StringValue(name),
			Description: stringVal(item, "description"),
			Priority:    intVal(item, "priority"),
			Enabled:     boolVal(item, "enabled"),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
