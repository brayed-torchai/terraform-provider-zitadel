package human_user

import (
	"context"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/user"

	"github.com/zitadel/terraform-provider-zitadel/v2/zitadel/helper"
)

func GetResource() *schema.Resource {
	return &schema.Resource{
		Description: "Resource representing a human user situated under an organization, which then can be authorized through memberships or direct grants on other resources.",
		Schema: map[string]*schema.Schema{
			helper.OrgIDVar: helper.OrgIDResourceField,
			userStateVar: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "State of the user",
				/* Not necessary as long as only active users are created
				ValidateDiagFunc: func(value interface{}, path cty.Path) diag.Diagnostics {
					return EnumValueValidation(userStateVar, value.(string), user.UserState_value)
				},*/
			},
			UserNameVar: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Username",
			},
			loginNamesVar: {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "Loginnames",
				ForceNew:    true,
			},
			preferredLoginNameVar: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Preferred login name",
				ForceNew:    true,
			},

			firstNameVar: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "First name of the user",
			},
			lastNameVar: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Last name of the user",
			},
			nickNameVar: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Nick name of the user",
			},
			DisplayNameVar: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Display name of the user",
				Computed:    true,
			},
			preferredLanguageVar: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Preferred language of the user",
				Default:     defaultPreferredLanguage,
			},
			genderVar: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Gender of the user" + helper.DescriptionEnumValuesList(user.Gender_name),
				ValidateDiagFunc: func(value interface{}, path cty.Path) diag.Diagnostics {
					return helper.EnumValueValidation(genderVar, value.(string), user.Gender_value)
				},
				Default: defaultGenderString,
			},
			emailVar: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Email of the user",
			},
			isEmailVerifiedVar: {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Is the email verified of the user, can only be true if password of the user is set",
			},
			phoneVar: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Phone of the user",
			},
			isPhoneVerifiedVar: {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Is the phone verified of the user",
			},
			InitialPasswordVar: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Initially set password for the user, not changeable after creation",
				Sensitive:   true,
				// We ignore if the value changes after creation or import
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool { return d.Id() != "" },
			},
			initialHashedPasswordVar: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Initial hashed password for the user, not changeable after creation. Being able to pass an initial hashed password is useful in migration scenarios.",
				// We ignore if the value changes after creation or import
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool { return d.Id() != "" },
				Sensitive:        true,
			},
			initialSkipPasswordChange: {
				Type:     schema.TypeBool,
				Optional: true,
				// We ignore if the value changes after creation or import
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool { return d.Id() != "" },
				Description:      "Whether the user has to change the password on first login.",
			},
		},
		ReadContext:   readFunc(false),
		CreateContext: create,
		DeleteContext: delete,
		UpdateContext: update,
		CustomizeDiff: customdiff.All(
			customdiff.IfValue(DisplayNameVar, func(ctx context.Context, value, meta interface{}) bool {
				if value == "" {
					return true
				}
				return false
			}, func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
				return diff.SetNew(DisplayNameVar, defaultDisplayName(diff.Get(firstNameVar).(string), diff.Get(lastNameVar).(string)))
			}),
			customdiff.IfValue(genderVar, func(ctx context.Context, value, meta interface{}) bool {
				if value == "" {
					return true
				}
				return false
			}, func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
				return diff.SetNew(genderVar, defaultGenderString)
			}),
			customdiff.IfValue(preferredLanguageVar, func(ctx context.Context, value, meta interface{}) bool {
				if value == "" {
					return true
				}
				return false
			}, func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
				return diff.SetNew(preferredLanguageVar, defaultPreferredLanguage)
			}),
		),
		Importer: helper.ImportWithIDAndOptionalOrgAndSecret(UserIDVar, InitialPasswordVar),
	}
}
