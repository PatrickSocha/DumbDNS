package database

var blockListSources = []sources{
	{
		regex: `0.0.0.0\s+(?P<url>\S+)`,
		url:   "https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
	},
}

var whitelistDatabase = map[string]struct{}{
	"spclient.wg.spotify.com": {}, // spotify
	"api-partner.spotify.com": {}, // spotify
	"i.scdn.co":               {}, // spotify
	"encore.scdn.co":          {}, // spotify

	// core
	"cdn.jsdelivr.net":                  {},
	"cdnjs.com":                         {},
	"unpkg.com":                         {},
	"cdnjs.cloudflare.com":              {},
	"downloaddispatch.itunes.apple.com": {}, // app store downloads
	"xp.apple.com":                      {}, // app store images
	"gsa.apple.com":                     {}, // sign in with apple
	"init.push.apple.com":               {}, // apple push / i think apple pay also
}

var hosts = map[string]string{
	"archive.ph": "23.137.248.133",
	"archive.is": "23.137.248.133",
}
