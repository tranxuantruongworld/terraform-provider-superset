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
	_ datasource.DataSource              = &userDataSource{}
	_ datasource.DataSourceWithConfigure = &userDataSource{}
)

func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

type userDataSource struct {
	client *client.Client
}

type userDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Username  types.String `tfsdk:"username"`
	FirstName types.String `tfsdk:"first_name"`
	LastName  types.String `tfsdk:"last_name"`
	Email     types.String `tfsdk:"email"`
	Active    types.Bool   `tfsdk:"active"`
	Roles     types.List   `tfsdk:"roles"`
}

func (d *userDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *userDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source to read an existing user in Apache Superset by username or ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the user. If not specified, 'username' must be provided.",
				Optional:    true,
				Computed:    true,
			},
			"username": schema.StringAttribute{
				Description: "The unique username of the user. If not specified, 'id' must be provided.",
				Optional:    true,
				Computed:    true,
			},
			"first_name": schema.StringAttribute{
				Description: "First name of the user.",
				Computed:    true,
			},
			"last_name": schema.StringAttribute{
				Description: "Last name of the user.",
				Computed:    true,
			},
			"email": schema.StringAttribute{
				Description: "Email address of the user.",
				Computed:    true,
			},
			"active": schema.BoolAttribute{
				Description: "Whether the user account is active.",
				Computed:    true,
			},
			"roles": schema.ListAttribute{
				Description: "List of role IDs assigned to the user.",
				ElementType: types.Int64Type,
				Computed:    true,
			},
		},
	}
}

func (d *userDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config userDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var user *client.User
	var err error

	if !config.ID.IsNull() && !config.ID.IsUnknown() && config.ID.ValueString() != "" {
		id, parseErr := strconv.ParseInt(config.ID.ValueString(), 10, 64)
		if parseErr != nil {
			resp.Diagnostics.AddError("Invalid User ID", "Could not parse user ID: "+parseErr.Error())
			return
		}
		user, err = d.client.GetUser(ctx, id)
	} else if !config.Username.IsNull() && !config.Username.IsUnknown() && config.Username.ValueString() != "" {
		user, err = d.client.GetUserByUsername(ctx, config.Username.ValueString())
	} else {
		resp.Diagnostics.AddError(
			"Missing Query Parameter",
			"Either 'id' or 'username' must be specified to query the superset_user data source.",
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Superset User Data Source",
			"Could not find user: "+err.Error(),
		)
		return
	}

	config.ID = types.StringValue(strconv.FormatInt(user.ID, 10))
	config.Username = types.StringValue(user.Username)
	config.FirstName = types.StringValue(user.FirstName)
	config.LastName = types.StringValue(user.LastName)
	config.Email = types.StringValue(user.Email)
	config.Active = types.BoolValue(user.Active)

	var roleIDs []int64
	for _, role := range user.Roles {
		roleIDs = append(roleIDs, role.ID)
	}
	rolesVal, diags := types.ListValueFrom(ctx, types.Int64Type, roleIDs)
	resp.Diagnostics.Append(diags...)
	config.Roles = rolesVal

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}
