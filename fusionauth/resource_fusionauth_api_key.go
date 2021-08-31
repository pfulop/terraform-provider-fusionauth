package fusionauth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/FusionAuth/go-client/pkg/fusionauth"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAPIKey() *schema.Resource {
	return &schema.Resource{
		Create: createAPIKey,
		Read:   readAPIKey,
		Update: updateAPIKey,
		Delete: deleteAPIKey,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The unique Id of the Tenant. This value is required if the key is meant to be tenant scoped. Tenant scoped keys can only be used to access users and other tenant scoped objects for the specified tenant. This value is read-only once the key is created.",
				ValidateFunc: validation.IsUUID,
				ForceNew:     true,
			},
			"key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The Id to use for the new Form. If not specified a secure random UUID will be generated.",
				ValidateFunc: validation.IsUUID,
			},
			"key": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "The Id to use for the new Form. If not specified a secure random UUID will be generated.",
				ValidateFunc: validation.IsUUID,
				Sensitive:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the key.",
			},
			"permissions_endpoints": {
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "Endpoint permissions for this key. Each key of the object is an endpoint, with the value being an array of the HTTP methods which can be used against the endpoint. An Empty permissions_endpoints object mean that this is a super key that authorizes this key for all the endpoints.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"/api/application",
								"/api/application/oauth-configuration",
								"/api/application/role",
								"/api/cleanspeak/notify",
								"/api/connector",
								"/api/consent",
								"/api/email/send",
								"/api/email/template",
								"/api/email/template/preview",
								"/api/entity",
								"/api/entity/grant",
								"/api/entity/grant/search",
								"/api/entity/search",
								"/api/entity/type",
								"/api/entity/type/permission",
								"/api/entity/type/search",
								"/api/form",
								"/api/form/field",
								"/api/group",
								"/api/group/member",
								"/api/identity-provider",
								"/api/identity-provider/link",
								"/api/integration",
								"/api/jwt/issue",
								"/api/jwt/refresh",
								"/api/jwt/validate",
								"/api/key",
								"/api/key/generate",
								"/api/key/import",
								"/api/lambda",
								"/api/logger",
								"/api/login",
								"/api/message/template",
								"/api/message/template/preview",
								"/api/messenger",
								"/api/passwordless/start",
								"/api/prometheus/metrics",
								"/api/reactor",
								"/api/report/daily-active-user",
								"/api/report/login",
								"/api/report/monthly-active-user",
								"/api/report/registration",
								"/api/report/totals",
								"/api/status",
								"/api/system-configuration",
								"/api/system/audit-log",
								"/api/system/audit-log/export",
								"/api/system/audit-log/search",
								"/api/system/event-log",
								"/api/system/event-log/search",
								"/api/system/log/export",
								"/api/system/login-record/export",
								"/api/system/login-record/search",
								"/api/system/reindex",
								"/api/system/version",
								"/api/tenant",
								"/api/theme",
								"/api/two-factor/secret",
								"/api/two-factor/send",
								"/api/two-factor/start",
								"/api/user",
								"/api/user-action",
								"/api/user-action-reason",
								"/api/user/action",
								"/api/user/bulk",
								"/api/user/change-password",
								"/api/user/comment",
								"/api/user/consent",
								"/api/user/family",
								"/api/user/family/pending",
								"/api/user/family/request",
								"/api/user/forgot-password",
								"/api/user/import",
								"/api/user/recent-login",
								"/api/user/refresh-token/import",
								"/api/user/registration",
								"/api/user/search",
								"/api/user/two-factor",
								"/api/user/two-factor/recovery-code",
								"/api/user/verify-email",
								"/api/user/verify-registration",
								"/api/webhook",
							}, false),
						},
						"delete": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "HTTP DELETE Verb",
						},
						"get": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "HTTP GET Verb",
						},
						"patch": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "HTTP PATCH Verb",
						},
						"post": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "HTTP POST Verb",
						},
						"put": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "HTTP PUT Verb",
						},
					},
				},
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func createAPIKey(data *schema.ResourceData, i interface{}) error {
	client := i.(Client)
	ak := buildAPIKey(data)

	oldTenantID := client.FAClient.TenantId
	client.FAClient.TenantId = ak.TenantId
	defer func() {
		client.FAClient.TenantId = oldTenantID
	}()
	kid := data.Get("key_id").(string)
	resp, faErrs, err := client.FAClient.CreateAPIKey(kid, fusionauth.APIKeyRequest{ApiKey: ak})
	if err != nil {
		return fmt.Errorf("createAPIKey errors: %v", err)
	}
	if err := checkResponse(resp.StatusCode, faErrs); err != nil {
		return err
	}
	data.SetId(resp.ApiKey.Id)
	return buildResourceDataFromAPIKey(data, resp.ApiKey)
}

func readAPIKey(data *schema.ResourceData, i interface{}) error {
	client := i.(Client)
	id := data.Id()

	resp, faErrs, err := client.FAClient.RetrieveAPIKey(id)
	if err != nil {
		return fmt.Errorf("readAPIKey errors: %v", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		data.SetId("")
		return nil
	}
	if err := checkResponse(resp.StatusCode, faErrs); err != nil {
		return err
	}

	if err := checkResponse(resp.StatusCode, nil); err != nil {
		return err
	}

	return buildResourceDataFromAPIKey(data, resp.ApiKey)
}

func updateAPIKey(data *schema.ResourceData, i interface{}) error {
	client := i.(Client)
	ak := buildAPIKey(data)

	oldTenantID := client.FAClient.TenantId
	client.FAClient.TenantId = ak.TenantId
	defer func() {
		client.FAClient.TenantId = oldTenantID
	}()

	resp, faErrs, err := client.FAClient.UpdateAPIKey(data.Id(), fusionauth.APIKeyRequest{ApiKey: ak})
	if err != nil {
		return fmt.Errorf("updateAPIKey errors: %v", err)
	}
	if err := checkResponse(resp.StatusCode, faErrs); err != nil {
		return err
	}

	data.SetId(resp.ApiKey.Id)
	return buildResourceDataFromAPIKey(data, resp.ApiKey)
}

func deleteAPIKey(data *schema.ResourceData, i interface{}) error {
	client := i.(Client)
	resp, faErrs, err := client.FAClient.DeleteAPIKey(data.Id())
	if err != nil {
		return err
	}
	if err := checkResponse(resp.StatusCode, faErrs); err != nil {
		return err
	}

	return nil
}
func buildAPIKey(data *schema.ResourceData) fusionauth.APIKey {
	ak := fusionauth.APIKey{
		Key:      data.Get("key").(string),
		TenantId: data.Get("tenant_id").(string),
		MetaData: fusionauth.APIKeyMetaData{
			Attributes: map[string]string{
				"description": data.Get("description").(string),
			},
		},
	}

	m := make(map[string][]string)
	s, ok := data.GetOk("permissions_endpoints")
	if !ok {
		return ak
	}
	set, ok := s.(*schema.Set)
	if !ok {
		return ak
	}
	l := set.List()
	for _, x := range l {
		ac := x.(map[string]interface{})
		ep := ac["endpoint"].(string)
		ss := []string{}
		if ac["delete"].(bool) {
			ss = append(ss, "DELETE")
		}
		if ac["get"].(bool) {
			ss = append(ss, "GET")
		}
		if ac["patch"].(bool) {
			ss = append(ss, "PATCH")
		}
		if ac["post"].(bool) {
			ss = append(ss, "POST")
		}
		if ac["put"].(bool) {
			ss = append(ss, "PUT")
		}
		m[ep] = ss
	}
	ak.Permissions.Endpoints = m
	return ak
}
func buildResourceDataFromAPIKey(data *schema.ResourceData, res fusionauth.APIKey) error {
	if err := data.Set("tenant_id", res.TenantId); err != nil {
		return fmt.Errorf("apiKey.tenant_id: %s", err.Error())
	}
	if err := data.Set("key", res.Key); err != nil {
		return fmt.Errorf("apiKey.key: %s", err.Error())
	}
	if desc, ok := res.MetaData.Attributes["description"]; ok {
		if err := data.Set("description", desc); err != nil {
			return fmt.Errorf("apiKey.description: %s", err.Error())
		}
	}
	if err := data.Set("tenant_id", res.TenantId); err != nil {
		return fmt.Errorf("apiKey.tenant_id: %s", err.Error())
	}

	pe := make([]map[string]interface{}, 0, len(res.Permissions.Endpoints))
	for ep := range res.Permissions.Endpoints {
		endpointPermission := map[string]interface{}{
			"delete": false,
			"get":    false,
			"patch":  false,
			"post":   false,
			"put":    false,
		}

		for _, s := range res.Permissions.Endpoints[ep] {
			method := strings.ToLower(s)
			if _, exists := endpointPermission[method]; exists {
				endpointPermission[method] = true
			}
		}

		endpointPermission["endpoint"] = ep
		pe = append(pe, endpointPermission)
	}

	if err := data.Set("permissions_endpoints", pe); err != nil {
		return fmt.Errorf("apiKey.permissions_endpoints: %s", err.Error())
	}
	return nil
}