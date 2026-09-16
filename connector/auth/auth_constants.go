package auth

var (
	//DEPRECATED: DO NOT USE
	ToChannelFromBotLoginURL = []string{
		"https://login.microsoftonline.com/botframework.com/oauth2/v2.0/token",
	}

	ToBotFromChannelOpenIDMetadataURL = []string{
		"https://login.botframework.com/v1/.well-known/openidconfiguration",
	}

	ToBotFromEnterpriseChannelOpenIDMetadataURLFormat = []string{
		"https://{channelService}.enterprisechannel.botframework.com",
		"/v1/.well-known/openidconfiguration",
	}

	ToBotFromEmulatorOpenIDMetadataURL = []string{
		"https://login.microsoftonline.com/common/v2.0/.well-known/openid-configuration",
	}

	AllowedSigningAlgorithms = []string{"RS256", "RS384", "RS512"}
)

const (
	ToChannelFromBotLoginURLPrefix = "https://login.microsoftonline.com/"

	ToChannelFromBotTokenEndpointPathTOCHANNELFROMBOTTOKENENDPOINTPATH = "/oauth2/v2.0/token"

	DefaultChannelAuthTenant = "botframework.com"

	ToChannelFromBotOauthScope = "https://api.botframework.com/.default"

	ToBotFromChannelTokenIssuer = "https://api.botframework.com"

	BotOpenIDMetadataKey = "BotOpenIdMetadata"

	ChannelService = "ChannelService"

	OauthURLKey = "OAuthApiEndpoint"

	EmulateOauthCardsKey = "EmulateOAuthCards"

	AuthorizedParty = "azp"

	AudienceClaim = "aud"

	IssuerClaim = "iss"

	KeyIDHeader = "kid"

	VersionClaim = "ver"

	AppIDClaim = "appid"

	ServiceURLClaim = "serviceurl"
)
