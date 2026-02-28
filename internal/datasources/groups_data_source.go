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

var _ datasource.DataSource = &GroupsDataSource{}

// GroupsDataSource lists IronWiFi groups.
type GroupsDataSource struct {
	client *client.Client
}

// GroupsDataSourceModel is the schema model for the groups data source.
type GroupsDataSourceModel struct {
	NameFilter types.String     `tfsdk:"name_filter"`
	Items      []GroupItemModel `tfsdk:"items"`
}

// GroupItemModel represents a single group in the data source output.
type GroupItemModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Priority    types.Int64  `tfsdk:"priority"`
}

// NewGroupsDataSource returns a new groups data source.
func NewGroupsDataSource() datasource.DataSource {
	return &GroupsDataSource{}
}

func (d *GroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

func (d *GroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists IronWiFi groups.",
		Attributes: map[string]schema.Attribute{
			"name_filter": schema.StringAttribute{
				Description: "Optional substring filter applied to group names.",
				Optional:    true,
			},
			"items": schema.ListNestedAttribute{
				Description: "The list of groups.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true, Description: "Group ID."},
						"name":        schema.StringAttribute{Computed: true, Description: "Group name."},
						"description": schema.StringAttribute{Computed: true, Description: "Group description."},
						"priority":    schema.Int64Attribute{Computed: true, Description: "Group priority."},
					},
				},
			},
		},
	}
}

func (d *GroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state GroupsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, err := d.client.List("groups", "groups")
	if err != nil {
		resp.Diagnostics.AddError("Error listing groups", err.Error())
		return
	}

	for _, item := range items {
		name := fmt.Sprintf("%v", item["groupname"])
		if !state.NameFilter.IsNull() && !state.NameFilter.IsUnknown() {
			if !strings.Contains(strings.ToLower(name), strings.ToLower(state.NameFilter.ValueString())) {
				continue
			}
		}
		state.Items = append(state.Items, GroupItemModel{
			ID:          stringVal(item, "id"),
			Name:        types.StringValue(name),
			Description: stringVal(item, "description"),
			Priority:    intVal(item, "priority"),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
