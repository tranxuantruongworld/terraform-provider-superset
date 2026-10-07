package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/xtruong/terraform-provider-superset/internal/client"
)

var (
	_ resource.Resource                = &userResource{}
	_ resource.ResourceWithConfigure   = &userResource{}
	_ resource.ResourceWithImportState = &userResource{}
)

func NewUserResource() resource.Resource {
	return &userResource{}
}

type userResource struct {
	client *client.Client
}

type userResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Username  types.String `tfsdk:"username"`
	FirstName types.String `tfsdk:"first_name"`
	LastName  types.String `tfsdk:"last_name"`
	Email     types.String `tfsdk:"email"`
	Active    types.Bool   `tfsdk:"active"`
	Password  types.String `tfsdk:"password"`
	Roles     types.List   `tfsdk:"roles"`
}

func (r *userResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a user and their role memberships within Apache Superset.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Numeric identifier of the user as a string.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Description: "Unique username for the user.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"first_name": schema.StringAttribute{
				Description: "First name of the user.",
				Required:    true,
			},
			"last_name": schema.StringAttribute{
				Description: "Last name of the user.",
				Required:    true,
			},
			"email": schema.StringAttribute{
				Description: "User's email address.",
				Required:    true,
			},
			"active": schema.BoolAttribute{
				Description: "Whether the user account is active. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"password": schema.StringAttribute{
				Description: "Password for the user. Only sent on creation or password update.",
				Optional:    true,
				Sensitive:   true,
			},
			"roles": schema.ListAttribute{
				Description: "List of role IDs assigned to this user.",
				ElementType: types.Int64Type,
				Optional:    true,
			},
		},
	}
}

func (r *userResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var roleIDs []int64
	if !plan.Roles.IsNull() && !plan.Roles.IsUnknown() {
		diags = plan.Roles.ElementsAs(ctx, &roleIDs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	password := ""
	if !plan.Password.IsNull() {
		password = plan.Password.ValueString()
	}

	createReq := &client.CreateUserRequest{
		Username:  plan.Username.ValueString(),
		FirstName: plan.FirstName.ValueString(),
		LastName:  plan.LastName.ValueString(),
		Email:     plan.Email.ValueString(),
		Active:    plan.Active.ValueBool(),
		Password:  password,
		Roles:     roleIDs,
	}

	user, err := r.client.CreateUser(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Superset User",
			"Could not create user, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(strconv.FormatInt(user.ID, 10))
	plan.Username = types.StringValue(user.Username)
	plan.FirstName = types.StringValue(user.FirstName)
	plan.LastName = types.StringValue(user.LastName)
	plan.Email = types.StringValue(user.Email)
	plan.Active = types.BoolValue(user.Active)

	var currentRoleIDs []int64
	for _, role := range user.Roles {
		currentRoleIDs = append(currentRoleIDs, role.ID)
	}
	rolesVal, diags := types.ListValueFrom(ctx, types.Int64Type, currentRoleIDs)
	resp.Diagnostics.Append(diags...)
	plan.Roles = rolesVal

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid User ID", "Could not parse user ID: "+err.Error())
		return
	}

	user, err := r.client.GetUser(ctx, id)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Superset User",
			fmt.Sprintf("Could not read user with ID %d: %s", id, err.Error()),
		)
		return
	}

	state.Username = types.StringValue(user.Username)
	state.FirstName = types.StringValue(user.FirstName)
	state.LastName = types.StringValue(user.LastName)
	state.Email = types.StringValue(user.Email)
	state.Active = types.BoolValue(user.Active)

	var roleIDs []int64
	for _, role := range user.Roles {
		roleIDs = append(roleIDs, role.ID)
	}
	rolesVal, diags := types.ListValueFrom(ctx, types.Int64Type, roleIDs)
	resp.Diagnostics.Append(diags...)
	state.Roles = rolesVal

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(plan.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid User ID", "Could not parse user ID: "+err.Error())
		return
	}

	var roleIDs []int64
	if !plan.Roles.IsNull() && !plan.Roles.IsUnknown() {
		diags = plan.Roles.ElementsAs(ctx, &roleIDs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	firstName := plan.FirstName.ValueString()
	lastName := plan.LastName.ValueString()
	email := plan.Email.ValueString()
	active := plan.Active.ValueBool()

	updateReq := &client.UpdateUserRequest{
		FirstName: &firstName,
		LastName:  &lastName,
		Email:     &email,
		Active:    &active,
		Roles:     roleIDs,
	}

	if !plan.Password.IsNull() && !plan.Password.IsUnknown() && plan.Password.ValueString() != "" {
		pwd := plan.Password.ValueString()
		updateReq.Password = &pwd
	}

	user, err := r.client.UpdateUser(ctx, id, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Superset User",
			fmt.Sprintf("Could not update user %d: %s", id, err.Error()),
		)
		return
	}

	plan.FirstName = types.StringValue(user.FirstName)
	plan.LastName = types.StringValue(user.LastName)
	plan.Email = types.StringValue(user.Email)
	plan.Active = types.BoolValue(user.Active)

	var currentRoleIDs []int64
	for _, role := range user.Roles {
		currentRoleIDs = append(currentRoleIDs, role.ID)
	}
	rolesVal, diags := types.ListValueFrom(ctx, types.Int64Type, currentRoleIDs)
	resp.Diagnostics.Append(diags...)
	plan.Roles = rolesVal

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid User ID", "Could not parse user ID: "+err.Error())
		return
	}

	if err := r.client.DeleteUser(ctx, id); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Superset User",
			fmt.Sprintf("Could not delete user %d: %s", id, err.Error()),
		)
		return
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
