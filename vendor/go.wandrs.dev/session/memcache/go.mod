module go.wandrs.dev/session/memcache

go 1.16

require (
	github.com/bradfitz/gomemcache v0.0.0-20190913173617-a41fca850d0b
	github.com/go-chi/chi v1.5.1
	github.com/smartystreets/goconvey v0.0.0-20190731233626-505e41936337
	go.wandrs.dev/session v0.0.0-20210531074458-568b38307750
)

replace go.wandrs.dev/session => ../
