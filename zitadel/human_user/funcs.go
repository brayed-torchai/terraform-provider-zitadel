package human_user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/management"
	"github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/user"

	"github.com/zitadel/terraform-provider-zitadel/v2/zitadel/helper"
)

func delete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, "started read")

	clientinfo, ok := m.(*helper.ClientInfo)
	if !ok {
		return diag.Errorf("failed to get client")
	}

	client, err := helper.GetManagementClient(ctx, clientinfo)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = client.RemoveUser(helper.CtxWithOrgID(ctx, d), &management.RemoveUserRequest{
		Id: d.Id(),
	})
	if err != nil {
		return diag.Errorf("failed to delete user: %v", err)
	}
	return nil
}

func create(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, "started read")

	clientinfo, ok := m.(*helper.ClientInfo)
	if !ok {
		return diag.Errorf("failed to get client")
	}

	client, err := helper.GetManagementClient(ctx, clientinfo)
	if err != nil {
		return diag.FromErr(err)
	}

	firstName := d.Get(firstNameVar).(string)
	lastName := d.Get(lastNameVar).(string)
	importUser := &management.ImportHumanUserRequest{
		UserName: d.Get(UserNameVar).(string),
		Profile: &management.ImportHumanUserRequest_Profile{
			FirstName:         firstName,
			LastName:          lastName,
			Gender:            user.Gender(user.Gender_value[d.Get(genderVar).(string)]),
			PreferredLanguage: d.Get(preferredLanguageVar).(string),
			NickName:          d.Get(nickNameVar).(string),
		},
		Password:               d.Get(InitialPasswordVar).(string),
		PasswordChangeRequired: !d.Get(initialSkipPasswordChange).(bool),
	}

	if hashedPassword, ok := d.GetOk(initialHashedPasswordVar); ok {
		importUser.HashedPassword = &management.ImportHumanUserRequest_HashedPassword{
			Value: hashedPassword.(string),
		}
	}

	if displayname, ok := d.GetOk(DisplayNameVar); ok {
		importUser.Profile.DisplayName = displayname.(string)
	} else {
		if err := d.Set(DisplayNameVar, defaultDisplayName(firstName, lastName)); err != nil {
			return diag.Errorf("failed to set default display name for human user: %v", err)
		}
	}

	if email, ok := d.GetOk(emailVar); ok {
		isVerified, isVerifiedOk := d.GetOk(isEmailVerifiedVar)
		importUser.Email = &management.ImportHumanUserRequest_Email{
			Email:           email.(string),
			IsEmailVerified: false,
		}
		if isVerifiedOk {
			importUser.Email.IsEmailVerified = isVerified.(bool)
		}
	}

	if phone, ok := d.GetOk(phoneVar); ok {
		isVerified, isVerifiedOk := d.GetOk(isPhoneVerifiedVar)
		importUser.Phone = &management.ImportHumanUserRequest_Phone{
			Phone:           phone.(string),
			IsPhoneVerified: false,
		}
		if isVerifiedOk {
			importUser.Phone.IsPhoneVerified = isVerified.(bool)
		}
	}

	respUser, err := client.ImportHumanUser(helper.CtxWithOrgID(ctx, d), importUser)
	if err != nil {
		return diag.Errorf("failed to create human user: %v", err)
	}
	d.SetId(respUser.UserId)
	// To avoid diffs for terraform plan -refresh=false right after creation, we query and set the computed values.
	// The acceptance tests rely on this, too.
	return read(ctx, d, m)
}

func update(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, "started update")

	clientinfo, ok := m.(*helper.ClientInfo)
	if !ok {
		return diag.Errorf("failed to get client")
	}

	client, err := helper.GetManagementClient(ctx, clientinfo)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange(UserNameVar) {
		_, err = client.UpdateUserName(helper.CtxWithOrgID(ctx, d), &management.UpdateUserNameRequest{
			UserId:   d.Id(),
			UserName: d.Get(UserNameVar).(string),
		})
		if err != nil {
			return diag.Errorf("failed to update username: %v", err)
		}
	}

	if d.HasChanges(firstNameVar, lastNameVar, nickNameVar, DisplayNameVar, preferredLanguageVar, genderVar) {
		_, err := client.UpdateHumanProfile(helper.CtxWithOrgID(ctx, d), &management.UpdateHumanProfileRequest{
			UserId:            d.Id(),
			FirstName:         d.Get(firstNameVar).(string),
			LastName:          d.Get(lastNameVar).(string),
			NickName:          d.Get(nickNameVar).(string),
			DisplayName:       d.Get(DisplayNameVar).(string),
			PreferredLanguage: d.Get(preferredLanguageVar).(string),
			Gender:            user.Gender(user.Gender_value[d.Get(genderVar).(string)]),
		})
		if err != nil {
			return diag.Errorf("failed to update human profile: %v", err)
		}
	}

	if d.HasChanges(emailVar, isEmailVerifiedVar) {
		_, err = client.UpdateHumanEmail(helper.CtxWithOrgID(ctx, d), &management.UpdateHumanEmailRequest{
			UserId:          d.Id(),
			Email:           d.Get(emailVar).(string),
			IsEmailVerified: d.Get(isEmailVerifiedVar).(bool),
		})
		if err != nil {
			return diag.Errorf("failed to update human email: %v", err)
		}
	}

	if d.HasChanges(phoneVar, isPhoneVerifiedVar) {
		_, err = client.UpdateHumanPhone(helper.CtxWithOrgID(ctx, d), &management.UpdateHumanPhoneRequest{
			UserId:          d.Id(),
			Phone:           d.Get(phoneVar).(string),
			IsPhoneVerified: d.Get(isPhoneVerifiedVar).(bool),
		})
		if err != nil {
			return diag.Errorf("failed to update human phone: %v", err)
		}
	}
	return nil
}

func read(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, "started read")

	clientinfo, ok := m.(*helper.ClientInfo)
	if !ok {
		return diag.Errorf("failed to get client")
	}

	client, err := helper.GetManagementClient(ctx, clientinfo)
	if err != nil {
		return diag.FromErr(err)
	}

	respUser, err := client.GetUserByID(helper.CtxWithOrgID(ctx, d), &management.GetUserByIDRequest{Id: helper.GetID(d, UserIDVar)})
	if err != nil && helper.IgnoreIfNotFoundError(err) == nil {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.Errorf("failed to get user")
	}

	user := respUser.GetUser()
	set := map[string]interface{}{
		helper.OrgIDVar:       user.GetDetails().GetResourceOwner(),
		userStateVar:          user.GetState().String(),
		UserNameVar:           user.GetUserName(),
		loginNamesVar:         user.GetLoginNames(),
		preferredLoginNameVar: user.GetPreferredLoginName(),
		// This will be ignored using the CustomizeDiff function.
		// However, we should explicitly set it to true or false so that importing a user doesn't produce an immediate plan diff.
		initialSkipPasswordChange: false,
	}

	if human := user.GetHuman(); human != nil {
		if profile := human.GetProfile(); profile != nil {
			set[firstNameVar] = profile.GetFirstName()
			set[lastNameVar] = profile.GetLastName()
			set[DisplayNameVar] = profile.GetDisplayName()
			set[nickNameVar] = profile.GetNickName()
			set[preferredLanguageVar] = profile.GetPreferredLanguage()
			if gender := profile.GetGender().String(); gender != "" {
				set[genderVar] = gender
			}
		}
		if email := human.GetEmail(); email != nil {
			set[emailVar] = email.GetEmail()
			set[isEmailVerifiedVar] = email.GetIsEmailVerified()
		}
		if phone := human.GetPhone(); phone != nil {
			set[phoneVar] = phone.GetPhone()
			set[isPhoneVerifiedVar] = phone.GetIsPhoneVerified()
		}
	}
	for k, v := range set {
		if err := d.Set(k, v); err != nil {
			return diag.Errorf("failed to set %s of user: %v", k, err)
		}
	}
	d.SetId(user.GetId())
	return nil
}

func defaultDisplayName(firstName, lastName string) string {
	return firstName + " " + lastName
}
