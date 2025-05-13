resource "zitadel_human_user" "default" {
  org_id                       = data.zitadel_org.default.id
  user_name                    = "humanfull@localhost.com"
  first_name                   = "firstname"
  last_name                    = "lastname"
  nick_name                    = "nickname"
  display_name                 = "displayname"
  preferred_language           = "de"
  gender                       = "GENDER_MALE"
  phone                        = "+41799999999"
  is_phone_verified            = true
  email                        = "test@zitadel.com"
  is_email_verified            = true
  initial_password             = "Password1!"
  initial_skip_password_change = true
}
