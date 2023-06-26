package product

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &productResource{}
	_ resource.ResourceWithConfigure   = &productResource{}
	_ resource.ResourceWithImportState = &productResource{}
)

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &productResource{}
}

// orderResource is the resource implementation.
type productResource struct {
	client *platform.ByProjectKeyRequestBuilder
}

// Metadata returns the data source type name.
func (r *productResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_product"
}

// Schema defines the schema for the data source.
func (r *productResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "An abstract sellable good with a set of Attributes defined by a Product Type. " +
			"Products themselves are not sellable. Instead, they act as a parent structure for Product Variants. " +
			"Each Product must have at least one Product Variant, which is called the Master Variant. " +
			"A single Product representation contains the current and the staged representation of its product data.\n\n" +
			"See also the [Products API Documentation](https://docs.commercetools.com/api/projects/products)",
		Version: 1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Description: "User-defined unique identifier of the Product. This is different from the key of a ProductVariant.",
				Optional:    true,
			},
			"version": schema.Int64Attribute{
				Computed: true,
			},
			"product_type": schema.StringAttribute{
				Required:      true,
				Description:   "The ID of the Product Type defining the Attributes of the Product. Cannot be changed.",
				Validators:    nil, // TODO UUID?
				PlanModifiers: nil,
			},
			"tax_category": schema.StringAttribute{
				Optional:      true,
				Description:   "The ID of the TaxCategory of the Product.",
				Validators:    nil, // TODO UUID?
				PlanModifiers: nil,
			},
			// TODO state
			"price_mode": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						PriceModeEmbedded,
						PriceModeStandalone,
					),
				},
				Description:         "Type of Price to be used when looking up a price for the Product.",
				MarkdownDescription: "",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Default:  stringdefault.StaticString(PriceModeEmbedded),
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			// These items all have maximal one item. We don't use SingleNestedBlock
			// here since it isn't quite robust currently.
			// See https://github.com/hashicorp/terraform-plugin-framework/issues/603
			"master_data": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"published": schema.BoolAttribute{
							Optional:            true,
							Description:         "true if the Product is published.",
							MarkdownDescription: "",
							Validators:          nil,
							PlanModifiers:       nil,
							Default:             booldefault.StaticBool(true),
							Computed:            true,
						},
					},
					Blocks: map[string]schema.Block{
						"current": productDataSchema("Current (published) data of the Product."),
					},
					Validators:    nil,
					PlanModifiers: nil,
				},
				Description:         "Contains the current and the staged representation of the product information.",
				MarkdownDescription: "",

				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (r *productResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data := req.ProviderData.(*utils.ProviderData)
	r.client = data.Client
}

func (r *productResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan Product
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	draft := plan.draft(ctx)
	var product *platform.Product

	err := retry.RetryContext(ctx, 20*time.Second, func() *retry.RetryError {
		var err error
		product, err = r.client.Products().Post(draft).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating product",
			err.Error(),
		)
		return
	}

	current := NewProductFromNative(*product)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *productResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state Product
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	product, err := r.client.Products().WithId(state.ID.ValueString()).Get().Execute(ctx)
	if err != nil {
		if errors.Is(err, platform.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading product",
			"Could not retrieve product, unexpected error: "+err.Error(),
		)
		return
	}

	current := NewProductFromNative(*product)

	// Set refreshed state
	diags = resp.State.Set(ctx, &current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *productResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan Product
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state Product
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := state.calculateUpdateActions(ctx, plan)
	var product *platform.Product

	err := retry.RetryContext(ctx, 5*time.Second, func() *retry.RetryError {
		var err error
		product, err = r.client.Products().WithId(state.ID.ValueString()).Post(input).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating product",
			"Could not update product, unexpected error: "+err.Error(),
		)
		return
	}

	current := NewProductFromNative(*product)

	// Set refreshed state
	diags = resp.State.Set(ctx, &current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *productResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state Product
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := retry.RetryContext(ctx, 5*time.Second, func() *retry.RetryError {
		_, err := r.client.Products().WithId(state.ID.ValueString()).Delete().Version(int(state.Version.ValueInt64())).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting product",
			"Could not delete product, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *productResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
