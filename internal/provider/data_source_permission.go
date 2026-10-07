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
	_ datasource.DataSource              = &permissionDataSource{}
	_ datasource.DataSourceWithConfigure = &permissionDataSource{}
)

func NewPermissionDataSource() datasource.DataSource {
	return &permissionDataSource{}
}

type permissionDataSource struct {
	client *client.Client
}

type permissionDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	PermissionName types.String `tfsdk:"permission_name"`
	ViewMenuName   types.String `tfsdk:"view_menu_name"`
}

func (d *permissionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permission"
}

func (d *permissionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source to query a Permission View Menu (PVM) in Apache Superset by its permission and view_menu names, or by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The numeric ID of the Permission View Menu as a string.",
				Optional:    true,
				Computed:    true,
			},
			"permission_name": schema.StringAttribute{
				Description: "The action name of the permission (e.g., 'can_read', 'can_write', 'all_datasource_access', 'database_access').",
				Optional:    true,
				Computed:    true,
			},
			"view_menu_name": schema.StringAttribute{
				Description: "The target view/menu/resource (e.g., 'Dashboard', 'Chart', 'Dataset', 'Database', 'all_datasource_access').",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (d *permissionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *permissionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config permissionDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var pvm *client.PermissionViewMenu
	var err error

	if !config.ID.IsNull() && !config.ID.IsUnknown() && config.ID.ValueString() != "" {
		id, parseErr := strconv.ParseInt(config.ID.ValueString(), 10, 64)
		if parseErr != nil {
			resp.Diagnostics.AddError("Invalid Permission ID", "Could not parse permission ID: "+parseErr.Error())
			return
		}
		pvm, err = d.client.GetPermissionViewMenuByID(ctx, id)
	} else if !config.PermissionName.IsNull() && !config.ViewMenuName.IsNull() &&
		!config.PermissionName.IsUnknown() && !config.ViewMenuName.IsUnknown() &&
		config.PermissionName.ValueString() != "" && config.ViewMenuName.ValueString() != "" {
		pvm, err = d.client.GetPermissionViewMenu(ctx, config.PermissionName.ValueString(), config.ViewMenuName.ValueString())
	} else {
		resp.Diagnostics.AddError(
			"Missing Query Parameters",
			"Either 'id' or both 'permission_name' and 'view_menu_name' must be specified to query the superset_permission data source.",
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Superset Permission Data Source",
			"Could not find permission view menu: "+err.Error(),
		)
		return
	}

	config.ID = types.StringValue(strconv.FormatInt(pvm.ID, 10))
	config.PermissionName = types.StringValue(pvm.Permission.Name)
	config.ViewMenuName = types.StringValue(pvm.ViewMenu.Name)

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}
