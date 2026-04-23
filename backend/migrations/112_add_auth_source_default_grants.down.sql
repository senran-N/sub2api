DROP TABLE IF EXISTS user_provider_default_grants;
DELETE FROM settings WHERE key IN (
  'auth_source_default_email_balance',
  'auth_source_default_email_concurrency',
  'auth_source_default_email_subscriptions',
  'auth_source_default_email_grant_on_signup',
  'auth_source_default_email_grant_on_first_bind',
  'auth_source_default_linuxdo_balance',
  'auth_source_default_linuxdo_concurrency',
  'auth_source_default_linuxdo_subscriptions',
  'auth_source_default_linuxdo_grant_on_signup',
  'auth_source_default_linuxdo_grant_on_first_bind',
  'auth_source_default_oidc_balance',
  'auth_source_default_oidc_concurrency',
  'auth_source_default_oidc_subscriptions',
  'auth_source_default_oidc_grant_on_signup',
  'auth_source_default_oidc_grant_on_first_bind',
  'auth_source_default_wechat_balance',
  'auth_source_default_wechat_concurrency',
  'auth_source_default_wechat_subscriptions',
  'auth_source_default_wechat_grant_on_signup',
  'auth_source_default_wechat_grant_on_first_bind',
  'force_email_on_third_party_signup'
);
