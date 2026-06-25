package stats

import "strings"

// GafamCompany represents one of the major tech companies tracked for GAFAM
// dominance statistics.
type GafamCompany int

const (
	GafamGoogle GafamCompany = iota
	GafamAmazon
	GafamMeta
	GafamApple
	GafamMicrosoft

	gafamCount
)

var gafamDomains = [gafamCount][]string{
	GafamGoogle: {
		// Search & Core
		"google.com", "google.co", "google.org", "google.net",
		"google.co.uk", "google.ca", "google.com.au", "google.de",
		"google.fr", "google.co.jp", "google.co.in", "google.com.br",
		"google.ru", "google.es", "google.it", "google.nl",
		"google.com.vn", "google.com.tw", "google.co.kr", "google.com.sg",
		"google.co.th", "google.com.ph", "google.co.id", "google.com.mx",
		"google.com.ar", "google.com.co", "google.cl", "google.com.tr",
		"google.com.sa", "google.ae", "google.co.za", "google.com.eg",
		"google.com.ng", "google.co.ke", "google.com.pk", "google.com.bd",
		"google.com.hk", "google.se", "google.no", "google.dk",
		"google.fi", "google.pl", "google.pt", "google.at", "google.ch",
		"google.be", "google.ie", "google.co.nz", "google.gr",
		"google.ro", "google.cz", "google.hu", "google.bg",
		"google.sk", "google.rs", "google.hr", "google.si",
		"google.lt", "google.lv", "google.ee",
		"goog", "abc.xyz",
		"googleapis.com", "gstatic.com",
		"googleusercontent.com", "googlesyndication.com",
		"googletagmanager.com", "googleadservices.com",
		"googledomains.com", "googlesource.com",
		"googlezip.net", "googleapps.com",
		"googlelabs.com", "googlegroups.com",
		"googlecommerce.com", "googleplex.com",
		"googlecnapps.cn",
		"1e100.net",
		"gvt1.com", "gvt2.com", "gvt3.com",

		// Gmail
		"gmail.com", "googlemail.com",

		// YouTube
		"youtube.com", "youtu.be", "ytimg.com",
		"googlevideo.com", "youtube-nocookie.com",
		"youtube-ui.l.google.com", "youtubekids.com",
		"youtubei.googleapis.com",
		"yt.be", "yt3.ggpht.com",

		// Ads & Analytics
		"google-analytics.com", "googleanalytics.com",
		"doubleclick.net", "2mdn.net",
		"googletagservices.com",
		"googlesyndication.com",
		"googleoptimize.com",
		"googleadservices.com",
		"admob.com", "app-measurement.com",

		// Cloud & Firebase
		"cloud.google.com",
		"firebaseio.com", "firebaseapp.com", "firebase.com",
		"firebasedatabase.app", "firebaseinstallations.googleapis.com",
		"crashlytics.com",
		"appspot.com",
		"run.app", "cloudfunctions.net",
		"googlehosted.com",
		"withgoogle.com",
		"withyoutube.com",

		// Maps & Earth
		"maps.google.com", "maps.googleapis.com",
		"earth.google.com", "earthengine.google.com",
		"khms.google.com",

		// Android & Chrome
		"android.com", "android.clients.google.com",
		"chromium.org", "chrome.com", "chromeexperiments.com",
		"chromestatus.com", "chromebook.com",

		// Workspace & Productivity
		"docs.google.com", "drive.google.com", "meet.google.com",
		"calendar.google.com", "keep.google.com",
		"sites.google.com", "groups.google.com",
		"chat.google.com", "currents.google.com",
		"jamboard.google.com",

		// Blogger & Sites
		"blogger.com", "blogspot.com",
		"blogblog.com",

		// Other Google services
		"ggpht.com",
		"recaptcha.net",
		"waze.com",
		"nest.com",
		"dropcam.com",
		"waymo.com",
		"x.company", "x.team",
		"deepmind.com", "deepmind.google",
		"kaggle.com",
		"tensorflow.org",
		"dartlang.org", "dart.dev", "flutter.dev", "pub.dev",
		"golang.org", "go.dev",
		"material.io",
		"web.dev",
		"bazel.build",
		"chromium.googlesource.com",
		"gerrit.googlesource.com",
		"area120.google.com",
		"madewithcode.com",
		"poly.google.com",
		"artsandculture.google.com",
		"stadia.com", "stadia.dev",
		"fitbit.com",
		"google.ai", "ai.google",
		"bard.google.com", "gemini.google.com",
		"looker.com",
		"mandiant.com",
		"virustotal.com",
		"chronicle.security",
		"siemplify.co",
		"dialogflow.com",
		"capitalone.cloud.google.com",
		"apigee.com",
		"fabric.io",
		"crbug.com",
		"debug.com",
		"g.co", "g.page", "goo.gl",
		"google.dev", "google.cloud",
		"amplitude.com",
		"panoramio.com",
		"zagat.com",
		"snapseed.com",
		"google.store",
	},
	GafamAmazon: {
		// Amazon Retail (global)
		"amazon.com", "amazon.co", "amazon.co.uk", "amazon.de",
		"amazon.fr", "amazon.co.jp", "amazon.in", "amazon.com.br",
		"amazon.es", "amazon.it", "amazon.ca", "amazon.com.au",
		"amazon.com.mx", "amazon.sg", "amazon.nl", "amazon.sa",
		"amazon.ae", "amazon.pl", "amazon.se", "amazon.com.tr",
		"amazon.com.be", "amazon.eg", "amazon.cn", "amazon.co.za",
		"amazon.com.ng", "amazon.cl", "amazon.com.co",
		"amzn.to", "amzn.com", "amzn.asia",
		"media-amazon.com", "ssl-images-amazon.com",
		"images-amazon.com", "images-na.ssl-images-amazon.com",
		"a2z.com",
		"aboutamazon.com", "aboutamazon.co.uk", "aboutamazon.eu",

		// AWS (Amazon Web Services)
		"amazonaws.com", "amazonaws.cn", "amazonaws.com.cn",
		"aws.amazon.com", "aws.dev",
		"awsstatic.com", "awscloud.com", "awsapps.com",
		"elasticbeanstalk.com", "elasticloadbalancing.com",
		"elasticmapreduce.com",
		"cloudfront.net",
		"awsglobalaccelerator.com",
		"amazonwebservices.com",
		"aws.a2z.com",
		"amazoncognito.com",
		"amazonses.com",

		// Amazon Trust & Security
		"amazontrust.com", "amazonssl.com",

		// Amazon Ads
		"amazon-adsystem.com", "amazonadsi.com",
		"aax.amazon.com",

		// Amazon Payments
		"amazonpay.com", "amazonpay.in",
		"pay.amazon.com",

		// Amazon Video & Prime
		"amazonvideo.com", "primevideo.com",
		"aiv-cdn.net", "aiv-delivery.net",
		"atv-ps.amazon.com", "pv-cdn.net",

		// Amazon Music
		"music.amazon.com", "amazonmusic.com",

		// Alexa & Echo
		"alexa.com", "alexa.amazon.com",
		"amazonalexa.com", "echospatial.com",

		// Kindle & Books
		"kindle.com", "amazonkindle.com",
		"kindlefc.com",
		"goodreads.com",

		// Audible
		"audible.com", "audible.co.uk", "audible.de",
		"audible.fr", "audible.co.jp", "audible.com.au",
		"audible.co.in", "audible.it", "audible.es", "audible.ca",
		"audtd.com",

		// Twitch
		"twitch.tv", "twitchcdn.net", "twitchsvc.net",
		"jtvnw.net", "ttvnw.net", "ext-twitch.tv",
		"twitchstatic.com",

		// Ring
		"ring.com", "ring.me",
		"fw.ring.com",

		// Whole Foods
		"wholefoodsmarket.com", "wholefoods.com",

		// IMDb & MGM
		"imdb.com", "imdbws.com",
		"imdb-video.media-imdb.com",
		"mgm.com",

		// Zappos
		"zappos.com", "6pm.com",

		// Pill Pack & Health
		"pillpack.com", "amazon.care",
		"one.amazon.com",

		// Eero (mesh WiFi)
		"eero.com",

		// Blink
		"blinkforhome.com", "immedia-semi.com",

		// Woot
		"woot.com",

		// Amazon Logistics & Delivery
		"amazonlogistics.com", "amazonlogistics.eu",
		"logistics.amazon.com",

		// Other Amazon services
		"amazonbusiness.com",
		"amazonfresh.com",
		"amazongames.com",
		"amazonstudios.com",
		"amazoninspect.com",
		"aws.training",
		"amazonappstore.com",
		"developer.amazon.com",
		"selling.amazon.com",
		"sellercentral.amazon.com",
		"fakespot.com",
		"boxofficemojo.com",
		"dpreview.com",
		"comixology.com",
		"lovefilm.com",
		"souq.com",
		"junglee.com",
		"amazon.jobs",
		"sustainability.aboutamazon.com",
		"amazonsmile.com",
	},
	GafamMeta: {
		// Facebook
		"facebook.com", "facebook.net",
		"fb.com", "fb.me", "fb.gg", "fb.watch",
		"fbcdn.net", "fbsbx.com",
		"facebookmail.com", "facebookcorewwi.onion",
		"fbpigeon.com",
		"facebook.com.br", "facebook.com.mx",
		"facebookblueprint.com",
		"facebookconnect.com",
		"facebookbrand.com",
		"facebook-hardware.com",
		"facebookenterprise.com",
		"facebookvirtualassistant.com",
		"facebookrecruiting.com",

		// Instagram
		"instagram.com", "cdninstagram.com",
		"instagram.net",
		"ig.me", "instagr.am",

		// WhatsApp
		"whatsapp.com", "whatsapp.net",
		"wa.me",
		"whatsapp-plus.info",

		// Messenger
		"messenger.com", "m.me",
		"messengerdevelopers.com",

		// Threads
		"threads.net", "threads.com",

		// Meta (parent company)
		"meta.com", "meta.net",
		"meta.ai",
		"metacareers.com",
		"metaquest.com",
		"about.meta.com",
		"developers.meta.com",
		"transparency.meta.com",

		// Oculus / Quest (VR/AR)
		"oculus.com", "oculuscdn.com",
		"oculusbrand.com",
		"oculusvr.com",

		// Workplace
		"workplace.com", "workplacemsolutions.com",
		"workplace.meta.com",

		// Horizon (metaverse)
		"horizon.meta.com",

		// GIPHY
		"giphy.com", "giphy-analytics.com",
		"gph.is",

		// Mapillary
		"mapillary.com",

		// CrowdTangle
		"crowdtangle.com",

		// Kustomer
		"kustomer.com",

		// Wit.ai (NLP)
		"wit.ai",

		// Spark AR
		"spark.meta.com", "sparkar.com",

		// Novi / Diem (crypto)
		"novi.com", "diem.com",

		// NPE (New Product Experimentation)
		"npe.com",

		// Other Meta services
		"accountkit.com",
		"atscaleconference.com",
		"bonfire.com",
		"bulletin.com",
		"expresswifi.com",
		"freebasics.com", "internet.org",
		"loom.com",
		"movefast.com",
		"pytorch.org",
		"onavo.com",
		"redkix.com",
		"relay.meta.com",
		"slingshot.com",
		"tfbnw.net",
		"thefacebook.com",
		"viewpoints.com",
		"wal.li",
		"llama.meta.com",
		"ai.meta.com",
		"producthunt.com",
	},
	GafamApple: {
		// Apple Core
		"apple.com", "apple.co",
		"apple.com.cn", "apple.co.uk", "apple.co.kr", "apple.co.jp",
		"apple.de", "apple.fr", "apple.it", "apple.es",
		"apple.com.au", "apple.com.br", "apple.com.mx",
		"apple.com.tw", "apple.com.tr",
		"apple.news",
		"apple.org",

		// iCloud
		"icloud.com", "icloud.com.cn",
		"icloud-content.com",
		"icloud.apple.com",

		// App Store & iTunes
		"itunes.com", "itunes.apple.com",
		"appstore.com",
		"apps.apple.com",
		"itunesconnect.apple.com",
		"appstoreconnect.apple.com",

		// CDN & Infrastructure
		"cdn-apple.com", "apple-dns.net",
		"mzstatic.com", "aaplimg.com",
		"apple-cloudkit.com", "apple-livephotoskit.com", "apple-mapkit.com",
		"swcdn.apple.com", "swdist.apple.com",
		"swdownload.apple.com", "swscan.apple.com",
		"updates.cdn-apple.com",
		"blobstore.apple.com",
		"devimages-cdn.apple.com",
		"gc.apple.com",
		"configuration.apple.com",
		"cds.apple.com",
		"ls.apple.com",
		"gs.apple.com",
		"osxapps.itunes.apple.com",
		"sylvan.apple.com",
		"mesu.apple.com",
		"gdmf.apple.com",
		"iosapps.itunes.apple.com",

		// Apple Services
		"push.apple.com",
		"siri.apple.com",
		"smoot.apple.com",
		"me.com",

		// Apple Music & Beats
		"applemusic.com", "music.apple.com",
		"beats1.apple.com",
		"beatsbydre.com", "beatsbydre.co.uk",

		// Shazam
		"shazam.com",
		"shazamid.com",

		// Apple TV+
		"tv.apple.com",
		"trailers.apple.com",

		// Apple Maps
		"maps.apple.com",

		// Apple Pay
		"applepaydomain.apple.com",
		"applepay.apple.com",

		// Apple Developer
		"developer.apple.com",
		"devforums.apple.com",
		"devstreaming-cdn.apple.com",
		"download.developer.apple.com",

		// Apple ID & Auth
		"appleid.apple.com",
		"appleid.cdn-apple.com",
		"gsa.apple.com",
		"idmsa.apple.com",

		// Apple Fitness & Health
		"fitness.apple.com",
		"health.apple.com",

		// Apple Education
		"education.apple.com",
		"school.apple.com",

		// Apple Business
		"business.apple.com",
		"businesschat.apple.com",

		// Apple Advertising
		"searchads.apple.com",
		"iad.apple.com",

		// TestFlight
		"testflight.apple.com",

		// Swift & Dev Tools
		"swift.org", "swiftpackageindex.com",

		// macOS / OS updates
		"osrecovery.apple.com",
		"support.apple.com",
		"locate.apple.com",
		"xp.apple.com",
		"setup.icloud.com",
		"mask.icloud.com",
		"mask-h2.icloud.com",
		"fmf.icloud.com",
		"fmfmobile.icloud.com",
		"fmipmobile.icloud.com",
		"feedbackassistant.apple.com",
		"reportaproblem.apple.com",

		// Other Apple acquisitions & services
		"claris.com", "filemaker.com",
		"texture.com",
		"workflow.is",
		"beddit.com",
		"primephonic.com",
		"mubi.com",
		"darksky.net",
		"tupleapp.com",
		"fleetsmith.com",
		"spektral.com",
		"voysis.com",
		"vilynx.com",
		"xnor.ai",
	},
	GafamMicrosoft: {
		// Microsoft Core
		"microsoft.com", "microsoft.net",
		"microsoftstore.com", "microsoftedge.com",
		"msftncsi.com", "msftconnecttest.com",
		"msedge.net",

		// Windows
		"windows.com", "windows.net",
		"windowsupdate.com", "windowsupdate.org",
		"windowsphone.com",

		// Office / Microsoft 365
		"office.com", "office.net", "office365.com", "office365.us",
		"officeppe.net", "officeapps.live.com",
		"sharepoint.com", "sharepoint.us",
		"onedrive.com", "onedrive.live.com",
		"onenote.com", "onenote.net",
		"sway.com", "sway.office.com",
		"powerbi.com",
		"powerapps.com",
		"powerautomate.com",
		"dynamics.com", "dynamics365.com",
		"microsoftstream.com",
		"yammer.com",

		// Outlook & Email
		"outlook.com", "outlook.live.com",
		"hotmail.com", "hotmail.co.uk", "hotmail.fr", "hotmail.de",
		"live.com", "live.net",
		"msn.com",

		// Teams
		"teams.microsoft.com", "teams.live.com",
		"statics.teams.cdn.office.net",

		// Azure & Cloud
		"azure.com", "azure.net",
		"azureedge.net", "azurefd.net",
		"azurewebsites.net", "azurestaticapps.net",
		"azure-dns.com", "azure-dns.net", "azure-dns.org", "azure-dns.info",
		"azurecontainer.io", "azurecr.io",
		"azure-api.net", "azure-mobile.net",
		"azure.microsoft.com",
		"trafficmanager.net",
		"cloudapp.net", "cloudapp.azure.com",
		"database.windows.net",
		"servicebus.windows.net",
		"blob.core.windows.net",
		"table.core.windows.net",
		"queue.core.windows.net",
		"file.core.windows.net",
		"vault.azure.net",

		// Auth & Identity
		"microsoftonline.com", "microsoftonline-p.com",
		"microsoftazuread-sso.com",
		"msftauth.net", "msauth.net", "msauthimages.net",
		"login.microsoftonline.com",
		"aadcdn.microsoftonline-p.com",
		"msocsp.com",

		// Bing & Search
		"bing.com", "bing.net",
		"bingapis.com", "bingsandbox.com",

		// LinkedIn
		"linkedin.com", "linkedin.cn",
		"licdn.com", "lnkd.in",
		"lynda.com",

		// GitHub
		"github.com", "github.io", "github.dev",
		"githubassets.com", "githubcopilot.com",
		"github.blog", "github.community",
		"githubusercontent.com",
		"github.githubassets.com",
		"copilot.github.com",
		"npm.com", "npmjs.com", "npmjs.org",
		"nuget.org",

		// Visual Studio & Dev Tools
		"visualstudio.com",
		"vsassets.io",
		"vscode.dev", "vscode.cdn.dev",
		"devdiv.microsoft.com",
		"dev.azure.com",
		"marketplace.visualstudio.com",
		"dotnet.microsoft.com",
		"typescriptlang.org",

		// Xbox & Gaming
		"xbox.com", "xbox.live.com",
		"xboxlive.com", "xboxservices.com",
		"xboxcontent.com", "xboxab.com",
		"minecraft.net", "mojang.com",
		"activision.com", "activisionblizzard.com",
		"blizzard.com", "blizzard.cn",
		"battle.net",
		"callofduty.com",
		"worldofwarcraft.com",
		"diablo.com",
		"overwatchleague.com", "overwatch.com",
		"hearthstone.com",
		"starcraft.com",
		"king.com", "midasplayer.com",
		"candycrush.com",
		"bethesda.net",
		"zenimax.com",

		// Skype & Communication
		"skype.com", "skypeecs.net",
		"skypeassets.com", "skypeforbusiness.com",

		// Cortana & AI
		"cortana.ai",
		"bing.com",
		"microsoftcognitiveservices.com",
		"openai.com",

		// Surface
		"surface.com",

		// Other Microsoft URLs
		"aka.ms", "1drv.ms", "onenote.com",
		"aspnetcdn.com", "dotnetfoundation.org",
		"azureiotcentral.com",
		"mixpanel.com",
		"playfab.com",
		"havok.com",
		"flip.com", "flipgrid.com",
		"msecnd.net",
		"gears5.com",
		"forzamotorsport.net", "forza.net",
		"halowaypoint.com",
		"rare.co.uk",
		"obsidian.net",
		"inxile-entertainment.com",
		"undead-labs.com",
		"ninjatheorycam.com",
		"hellblade.com",
		"compulsiongames.com",
		"doublefine.com",
		"playground-games.com",
		"343industries.com",
		"thecoalitionstudio.com",
		"nuance.com", "nuancemedicare.com",
		"swiftkey.com",
		"semantic-machines.com",
		"affirmednetworks.com",
		"metaswitch.com",
		"cloudknox.io",
		"xandr.com", "appnexus.com",
		"clarity.ms",
		"msidentity.com",
		"microsoftpressstore.com",
	},
}

// matchGafamCompany returns the GafamCompany that the domain belongs to, or -1
// if the domain does not belong to any tracked company.
func matchGafamCompany(domain string) GafamCompany {
	lower := strings.ToLower(domain)
	for company := GafamCompany(0); company < gafamCount; company++ {
		for _, d := range gafamDomains[company] {
			if lower == d || strings.HasSuffix(lower, "."+d) {
				return company
			}
		}
	}

	return -1
}
