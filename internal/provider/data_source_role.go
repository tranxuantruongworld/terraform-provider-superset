package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/xtruong/terraform-provider-superset/internal/client"
)

var (
	_ datasource.DataSource              = &roleDataSource{}
	_ datasource.DataSourceWithConfigure = &roleDataSource{}
)

func NewRoleDataSource() datasource.DataSource {
	return &roleDataSource{}
}

type roleDataSource struct {
	client *client.Client
}

type roleDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *roleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (d *roleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source to read an existing role in Apache Superset by name or ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the role. If not specified, 'name' must be provided.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the role (e.g. 'Admin', 'Alpha', 'Gamma', 'Public'). If not specified, 'id' must be provided.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (d *roleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *roleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config roleDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var role *client.Role
	var err error

	if !config.ID.IsNull() && !config.ID.IsUnknown() && config.ID.ValueString() != "" {
		id, parseErr := strconv.ParseInt(config.ID.ValueString(), 10, 64)
		if parseErr != nil {
			resp.Diagnostics.AddError("Invalid Role ID", "Could not parse role ID: "+parseErr.Error())
			return
		}
		role, err = d.client.GetRole(ctx, id)
	} else if !config.Name.IsNull() && !config.Name.IsUnknown() && config.Name.ValueString() != "" {
		role, err = d.client.GetRoleByName(ctx, config.Name.ValueString())
	} else {
		resp.Diagnostics.AddError(
			"Missing Query Parameter",
			"Either 'id' or 'name' must be specified to query the superset_role data source.",
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Superset Role Data Source",
			"Could not find role: "+err.Error(),
		)
		return
	}

	config.ID = types.StringValue(strconv.FormatInt(role.ID, 10))
	config.Name = types.StringValue(role.Name)

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

