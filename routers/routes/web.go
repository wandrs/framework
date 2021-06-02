// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package routes

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	// to registers all internal adapters

	"go.wandrs.dev/captcha"
	"go.wandrs.dev/framework/modules/context"
	"go.wandrs.dev/framework/modules/httpcache"
	"go.wandrs.dev/framework/modules/log"
	"go.wandrs.dev/framework/modules/metrics"
	"go.wandrs.dev/framework/modules/public"
	_ "go.wandrs.dev/framework/modules/session"
	"go.wandrs.dev/framework/modules/setting"
	"go.wandrs.dev/framework/modules/storage"
	"go.wandrs.dev/framework/modules/templates"
	"go.wandrs.dev/framework/modules/validation"
	"go.wandrs.dev/framework/modules/web"
	"go.wandrs.dev/framework/routers"
	"go.wandrs.dev/framework/routers/admin"
	apiv1 "go.wandrs.dev/framework/routers/api/v1"
	"go.wandrs.dev/framework/routers/api/v1/misc"
	"go.wandrs.dev/framework/routers/dev"
	"go.wandrs.dev/framework/routers/org"
	"go.wandrs.dev/framework/routers/private"
	"go.wandrs.dev/framework/routers/user"
	userSetting "go.wandrs.dev/framework/routers/user/setting"
	"go.wandrs.dev/framework/services/forms"
	"go.wandrs.dev/framework/services/mailer"
	"go.wandrs.dev/session"

	"github.com/NYTimes/gziphandler"
	"github.com/chi-middleware/proxy"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tstranex/u2f"
)

const (
	// GzipMinSize represents min size to compress for the body size of response
	GzipMinSize = 1400
)

func commonMiddlewares() []func(http.Handler) http.Handler {
	var handlers = []func(http.Handler) http.Handler{
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
				next.ServeHTTP(context.NewResponse(resp), req)
			})
		},
	}

	if setting.ReverseProxyLimit > 0 {
		opt := proxy.NewForwardedHeadersOptions().
			WithForwardLimit(setting.ReverseProxyLimit).
			ClearTrustedProxies()
		for _, n := range setting.ReverseProxyTrustedProxies {
			if !strings.Contains(n, "/") {
				opt.AddTrustedProxy(n)
			} else {
				opt.AddTrustedNetwork(n)
			}
		}
		handlers = append(handlers, proxy.ForwardedHeaders(opt))
	}

	handlers = append(handlers, middleware.StripSlashes)

	if !setting.DisableRouterLog && setting.RouterLogLevel != log.NONE {
		if log.GetLogger("router").GetLevel() <= setting.RouterLogLevel {
			handlers = append(handlers, LoggerHandler(setting.RouterLogLevel))
		}
	}
	if setting.EnableAccessLog {
		handlers = append(handlers, context.AccessLogger())
	}

	handlers = append(handlers, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			// Why we need this? The Recovery() will try to render a beautiful
			// error page for user, but the process can still panic again, and other
			// middleware like session also may panic then we have to recover twice
			// and send a simple error page that should not panic any more.
			defer func() {
				if err := recover(); err != nil {
					combinedErr := fmt.Sprintf("PANIC: %v\n%s", err, string(log.Stack(2)))
					log.Error("%v", combinedErr)
					if setting.IsProd() {
						http.Error(resp, http.StatusText(500), 500)
					} else {
						http.Error(resp, combinedErr, 500)
					}
				}
			}()
			next.ServeHTTP(resp, req)
		})
	})
	return handlers
}

var corsHandler func(http.Handler) http.Handler

// NormalRoutes represents non install routes
func NormalRoutes() *web.Route {
	r := web.NewRoute()
	for _, middle := range commonMiddlewares() {
		r.Use(middle)
	}

	if setting.CORSConfig.Enabled {
		corsHandler = cors.Handler(cors.Options{
			//Scheme:           setting.CORSConfig.Scheme, // FIXME: the cors middleware needs scheme option
			AllowedOrigins: setting.CORSConfig.AllowDomain,
			//setting.CORSConfig.AllowSubdomain // FIXME: the cors middleware needs allowSubdomain option
			AllowedMethods:   setting.CORSConfig.Methods,
			AllowCredentials: setting.CORSConfig.AllowCredentials,
			MaxAge:           int(setting.CORSConfig.MaxAge.Seconds()),
		})
	} else {
		corsHandler = func(next http.Handler) http.Handler {
			return next
		}
	}

	r.Mount("/", WebRoutes())
	r.Mount("/api/v1", apiv1.Routes())
	r.Mount("/api/internal", private.Routes())
	return r
}

// WebRoutes returns all web routes
func WebRoutes() *web.Route {
	routes := web.NewRoute()

	routes.Use(public.AssetsHandler(&public.Options{
		Directory:   path.Join(setting.StaticRootPath, "public"),
		Prefix:      "/assets",
		CorsHandler: corsHandler,
	}))

	routes.Use(session.Sessioner(session.Options{
		Provider:       setting.SessionConfig.Provider,
		ProviderConfig: setting.SessionConfig.ProviderConfig,
		CookieName:     setting.SessionConfig.CookieName,
		CookiePath:     setting.SessionConfig.CookiePath,
		Gclifetime:     setting.SessionConfig.Gclifetime,
		Maxlifetime:    setting.SessionConfig.Maxlifetime,
		Secure:         setting.SessionConfig.Secure,
		Domain:         setting.SessionConfig.Domain,
	}))

	routes.Use(Recovery())

	// We use r.Route here over r.Use because this prevents requests that are not for avatars having to go through this additional handler
	routes.Route("/avatars/*", "GET, HEAD", storageHandler(setting.Avatar.Storage, "avatars", storage.Avatars))

	// for health check - doeesn't need to be passed through gzip handler
	routes.Head("/", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// this png is very likely to always be below the limit for gzip so it doesn't need to pass through gzip
	routes.Get("/apple-touch-icon.png", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, path.Join(setting.StaticURLPrefix, "/assets/img/apple-touch-icon.png"), 301)
	})

	gob.Register(&u2f.Challenge{})

	common := []interface{}{}

	if setting.EnableGzip {
		h, err := gziphandler.GzipHandlerWithOpts(gziphandler.MinSize(GzipMinSize))
		if err != nil {
			log.Fatal("GzipHandlerWithOpts failed: %v", err)
		}
		common = append(common, h)
	}

	mailer.InitMailRender(templates.Mailer())

	if setting.Service.EnableCaptcha {
		// The captcha http.Handler should only fire on /captcha/* so we can just mount this on that url
		routes.Route("/captcha/*", "GET,HEAD", append(common, captcha.Captchaer(context.GetImageCaptcha()))...)
	}

	if setting.HasRobotsTxt {
		routes.Get("/robots.txt", append(common, func(w http.ResponseWriter, req *http.Request) {
			filePath := path.Join(setting.CustomPath, "robots.txt")
			fi, err := os.Stat(filePath)
			if err == nil && httpcache.HandleTimeCache(req, w, fi) {
				return
			}
			http.ServeFile(w, req, filePath)
		})...)
	}

	// prometheus metrics endpoint - do not need to go through contexter
	if setting.Metrics.Enabled {
		c := metrics.NewCollector()
		prometheus.MustRegister(c)

		routes.Get("/metrics", append(common, routers.Metrics)...)
	}

	// Removed: toolbox.Toolboxer middleware will provide debug informations which seems unnecessary
	common = append(common, context.Contexter())

	// GetHead allows a HEAD request redirect to GET if HEAD method is not defined for that route
	common = append(common, middleware.GetHead)

	if setting.API.EnableSwagger {
		// Note: The route moved from apiroutes because it's in fact want to render a web page
		routes.Get("/api/swagger", append(common, misc.Swagger)...) // Render V1 by default
	}

	others := web.NewRoute()
	for _, middle := range common {
		others.Use(middle)
	}

	RegisterRoutes(others)
	routes.Mount("", others)
	return routes
}

// RegisterRoutes register routes
func RegisterRoutes(m *web.Route) {
	reqSignIn := context.Toggle(&context.ToggleOptions{SignInRequired: true})
	ignSignIn := context.Toggle(&context.ToggleOptions{SignInRequired: setting.Service.RequireSignInView})
	ignExploreSignIn := context.Toggle(&context.ToggleOptions{SignInRequired: setting.Service.RequireSignInView || setting.Service.Explore.RequireSigninView})
	ignSignInAndCsrf := context.Toggle(&context.ToggleOptions{DisableCSRF: true})
	reqSignOut := context.Toggle(&context.ToggleOptions{SignOutRequired: true})

	//bindIgnErr := binding.BindIgnErr
	bindIgnErr := web.Bind
	validation.AddBindingRules()

	openIDSignInEnabled := func(ctx *context.Context) {
		if !setting.Service.EnableOpenIDSignIn {
			ctx.Error(http.StatusForbidden)
			return
		}
	}

	openIDSignUpEnabled := func(ctx *context.Context) {
		if !setting.Service.EnableOpenIDSignUp {
			ctx.Error(http.StatusForbidden)
			return
		}
	}

	// FIXME: not all routes need go through same middleware.
	// Especially some AJAX requests, we can reduce middleware number to improve performance.
	// Routers.
	// for health check
	m.Get("/", routers.Home)
	m.Get("/.well-known/openid-configuration", user.OIDCWellKnown)
	m.Group("/explore", func() {
		m.Get("", func(ctx *context.Context) {
			ctx.Redirect(setting.AppSubURL + "/explore/users")
		})
		m.Get("/users", routers.ExploreUsers)
		m.Get("/organizations", routers.ExploreOrganizations)
	}, ignExploreSignIn)

	// ***** START: User *****
	m.Group("/user", func() {
		m.Get("/login", user.SignIn)
		m.Post("/login", bindIgnErr(forms.SignInForm{}), user.SignInPost)
		m.Group("", func() {
			m.Combo("/login/openid").
				Get(user.SignInOpenID).
				Post(bindIgnErr(forms.SignInOpenIDForm{}), user.SignInOpenIDPost)
		}, openIDSignInEnabled)
		m.Group("/openid", func() {
			m.Combo("/connect").
				Get(user.ConnectOpenID).
				Post(bindIgnErr(forms.ConnectOpenIDForm{}), user.ConnectOpenIDPost)
			m.Group("/register", func() {
				m.Combo("").
					Get(user.RegisterOpenID, openIDSignUpEnabled).
					Post(bindIgnErr(forms.SignUpOpenIDForm{}), user.RegisterOpenIDPost)
			}, openIDSignUpEnabled)
		}, openIDSignInEnabled)
		m.Get("/sign_up", user.SignUp)
		m.Post("/sign_up", bindIgnErr(forms.RegisterForm{}), user.SignUpPost)
		m.Group("/oauth2", func() {
			m.Get("/{provider}", user.SignInOAuth)
			m.Get("/{provider}/callback", user.SignInOAuthCallback)
		})
		m.Get("/link_account", user.LinkAccount)
		m.Post("/link_account_signin", bindIgnErr(forms.SignInForm{}), user.LinkAccountPostSignIn)
		m.Post("/link_account_signup", bindIgnErr(forms.RegisterForm{}), user.LinkAccountPostRegister)
		m.Group("/two_factor", func() {
			m.Get("", user.TwoFactor)
			m.Post("", bindIgnErr(forms.TwoFactorAuthForm{}), user.TwoFactorPost)
			m.Get("/scratch", user.TwoFactorScratch)
			m.Post("/scratch", bindIgnErr(forms.TwoFactorScratchAuthForm{}), user.TwoFactorScratchPost)
		})
		m.Group("/u2f", func() {
			m.Get("", user.U2F)
			m.Get("/challenge", user.U2FChallenge)
			m.Post("/sign", bindIgnErr(u2f.SignResponse{}), user.U2FSign)

		})
	}, reqSignOut)

	m.Group("/login/oauth", func() {
		m.Get("/authorize", bindIgnErr(forms.AuthorizationForm{}), user.AuthorizeOAuth)
		m.Post("/grant", bindIgnErr(forms.GrantApplicationForm{}), user.GrantApplicationOAuth)
		// TODO manage redirection
		m.Post("/authorize", bindIgnErr(forms.AuthorizationForm{}), user.AuthorizeOAuth)
	}, ignSignInAndCsrf, reqSignIn)
	m.Get("/login/oauth/userinfo", ignSignInAndCsrf, user.InfoOAuth)
	m.Post("/login/oauth/access_token", corsHandler, bindIgnErr(forms.AccessTokenForm{}), ignSignInAndCsrf, user.AccessTokenOAuth)

	m.Group("/user/settings", func() {
		m.Get("", userSetting.Profile)
		m.Post("", bindIgnErr(forms.UpdateProfileForm{}), userSetting.ProfilePost)
		m.Get("/change_password", user.MustChangePassword)
		m.Post("/change_password", bindIgnErr(forms.MustChangePasswordForm{}), user.MustChangePasswordPost)
		m.Post("/avatar", bindIgnErr(forms.AvatarForm{}), userSetting.AvatarPost)
		m.Post("/avatar/delete", userSetting.DeleteAvatar)
		m.Group("/account", func() {
			m.Combo("").Get(userSetting.Account).Post(bindIgnErr(forms.ChangePasswordForm{}), userSetting.AccountPost)
			m.Post("/email", bindIgnErr(forms.AddEmailForm{}), userSetting.EmailPost)
			m.Post("/email/delete", userSetting.DeleteEmail)
			m.Post("/delete", userSetting.DeleteAccount)
			m.Post("/theme", bindIgnErr(forms.UpdateThemeForm{}), userSetting.UpdateUIThemePost)
		})
		m.Group("/security", func() {
			m.Get("", userSetting.Security)
			m.Group("/two_factor", func() {
				m.Post("/regenerate_scratch", userSetting.RegenerateScratchTwoFactor)
				m.Post("/disable", userSetting.DisableTwoFactor)
				m.Get("/enroll", userSetting.EnrollTwoFactor)
				m.Post("/enroll", bindIgnErr(forms.TwoFactorAuthForm{}), userSetting.EnrollTwoFactorPost)
			})
			m.Group("/u2f", func() {
				m.Post("/request_register", bindIgnErr(forms.U2FRegistrationForm{}), userSetting.U2FRegister)
				m.Post("/register", bindIgnErr(u2f.RegisterResponse{}), userSetting.U2FRegisterPost)
				m.Post("/delete", bindIgnErr(forms.U2FDeleteForm{}), userSetting.U2FDelete)
			})
			m.Group("/openid", func() {
				m.Post("", bindIgnErr(forms.AddOpenIDForm{}), userSetting.OpenIDPost)
				m.Post("/delete", userSetting.DeleteOpenID)
				m.Post("/toggle_visibility", userSetting.ToggleOpenIDVisibility)
			}, openIDSignInEnabled)
			m.Post("/account_link", userSetting.DeleteAccountLink)
		})
		m.Group("/applications/oauth2", func() {
			m.Get("/{id}", userSetting.OAuth2ApplicationShow)
			m.Post("/{id}", bindIgnErr(forms.EditOAuth2ApplicationForm{}), userSetting.OAuthApplicationsEdit)
			m.Post("/{id}/regenerate_secret", userSetting.OAuthApplicationsRegenerateSecret)
			m.Post("", bindIgnErr(forms.EditOAuth2ApplicationForm{}), userSetting.OAuthApplicationsPost)
			m.Post("/delete", userSetting.DeleteOAuth2Application)
			m.Post("/revoke", userSetting.RevokeOAuth2Grant)
		})
		m.Combo("/applications").Get(userSetting.Applications).
			Post(bindIgnErr(forms.NewAccessTokenForm{}), userSetting.ApplicationsPost)
		m.Post("/applications/delete", userSetting.DeleteApplication)
		m.Get("/organization", userSetting.Organization)
	}, reqSignIn, func(ctx *context.Context) {
		ctx.Data["PageIsUserSettings"] = true
		ctx.Data["AllThemes"] = setting.UI.Themes
	})

	m.Group("/user", func() {
		// r.Get("/feeds", binding.Bind(auth.FeedsForm{}), user.Feeds)
		m.Get("/activate", user.Activate, reqSignIn)
		m.Post("/activate", user.ActivatePost, reqSignIn)
		m.Any("/activate_email", user.ActivateEmail)
		m.Get("/avatar/{username}/{size}", user.Avatar)
		m.Get("/email2user", user.Email2User)
		m.Get("/recover_account", user.ResetPasswd)
		m.Post("/recover_account", user.ResetPasswdPost)
		m.Get("/forgot_password", user.ForgotPasswd)
		m.Post("/forgot_password", user.ForgotPasswdPost)
		m.Post("/logout", user.SignOut)
	})
	// ***** END: User *****

	m.Get("/avatar/{hash}", user.AvatarByEmailHash)

	adminReq := context.Toggle(&context.ToggleOptions{SignInRequired: true, AdminRequired: true})

	// ***** START: Admin *****
	m.Group("/admin", func() {
		m.Get("", adminReq, admin.Dashboard)
		m.Post("", adminReq, bindIgnErr(forms.AdminDashboardForm{}), admin.DashboardPost)
		m.Get("/config", admin.Config)
		m.Post("/config/test_mail", admin.SendTestMail)
		m.Group("/monitor", func() {
			m.Get("", admin.Monitor)
			m.Post("/cancel/{pid}", admin.MonitorCancel)
			m.Group("/queue/{qid}", func() {
				m.Get("", admin.Queue)
				m.Post("/set", admin.SetQueueSettings)
				m.Post("/add", admin.AddWorkers)
				m.Post("/cancel/{pid}", admin.WorkerCancel)
				m.Post("/flush", admin.Flush)
			})
		})

		m.Group("/users", func() {
			m.Get("", admin.Users)
			m.Combo("/new").Get(admin.NewUser).Post(bindIgnErr(forms.AdminCreateUserForm{}), admin.NewUserPost)
			m.Combo("/{userid}").Get(admin.EditUser).Post(bindIgnErr(forms.AdminEditUserForm{}), admin.EditUserPost)
			m.Post("/{userid}/delete", admin.DeleteUser)
		})

		m.Group("/emails", func() {
			m.Get("", admin.Emails)
			m.Post("/activate", admin.ActivateEmail)
		})

		m.Group("/orgs", func() {
			m.Get("", admin.Organizations)
		})

		m.Group("/auths", func() {
			m.Get("", admin.Authentications)
			m.Combo("/new").Get(admin.NewAuthSource).Post(bindIgnErr(forms.AuthenticationForm{}), admin.NewAuthSourcePost)
			m.Combo("/{authid}").Get(admin.EditAuthSource).
				Post(bindIgnErr(forms.AuthenticationForm{}), admin.EditAuthSourcePost)
			m.Post("/{authid}/delete", admin.DeleteAuthSource)
		})

		m.Group("/notices", func() {
			m.Get("", admin.Notices)
			m.Post("/delete", admin.DeleteNotices)
			m.Post("/empty", admin.EmptyNotices)
		})
	}, adminReq)
	// ***** END: Admin *****

	m.Group("", func() {
		m.Get("/{username}", user.Profile)
	}, ignSignIn)

	m.Group("/{username}", func() {
		m.Post("/action/{action}", user.Action)
	}, reqSignIn)

	if !setting.IsProd() {
		m.Get("/template/*", dev.TemplatePreview)
	}

	// ***** START: Organization *****
	m.Group("/org", func() {
		m.Group("", func() {
			m.Get("/create", org.Create)
			m.Post("/create", bindIgnErr(forms.CreateOrgForm{}), org.CreatePost)
		})

		m.Group("/{org}", func() {
			m.Get("/dashboard", user.Dashboard)
			m.Get("/dashboard/{team}", user.Dashboard)
			m.Get("/members", org.Members)
			m.Get("/teams", org.Teams)
		}, context.OrgAssignment(true, false, true))

		m.Group("/{org}", func() {
			m.Get("/teams/{team}", org.TeamMembers)
		}, context.OrgAssignment(true, false, true))

		m.Group("/{org}", func() {
			m.Get("/teams/new", org.NewTeam)
			m.Post("/teams/new", bindIgnErr(forms.CreateTeamForm{}), org.NewTeamPost)
			m.Get("/teams/{team}/edit", org.EditTeam)
			m.Post("/teams/{team}/edit", bindIgnErr(forms.CreateTeamForm{}), org.EditTeamPost)
			m.Post("/teams/{team}/delete", org.DeleteTeam)

			m.Group("/settings", func() {
				m.Combo("").Get(org.Settings).
					Post(bindIgnErr(forms.UpdateOrgSettingForm{}), org.SettingsPost)
				m.Post("/avatar", bindIgnErr(forms.AvatarForm{}), org.SettingsAvatar)
				m.Post("/avatar/delete", org.SettingsDeleteAvatar)

				m.Route("/delete", "GET,POST", org.SettingsDelete)
			})
		}, context.OrgAssignment(true, true))
	}, reqSignIn)
	// ***** END: Organization *****

	if setting.API.EnableSwagger {
		m.Get("/swagger.v1.json", routers.SwaggerV1Json)
	}

	// Not found handler.
	m.NotFound(web.Wrap(routers.NotFound))
}
