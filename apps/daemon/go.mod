module github.com/daytonaio/daemon

go 1.23.0

// v0.5.0 breaks tailscale-connected docker clients so we need to pin it to v0.4.0
replace github.com/docker/go-connections => github.com/docker/go-connections v0.4.0

// samber/lo v1.47.0 - required by headscale breaks frp
replace github.com/samber/lo => github.com/samber/lo v1.39.0

require (
	github.com/creack/pty v1.1.23
	github.com/gin-gonic/gin v1.10.1
	github.com/gliderlabs/ssh v0.3.7
	github.com/go-git/go-git/v5 v5.12.1-0.20240617075238-c127d1b35535
	github.com/go-playground/validator/v10 v10.24.0
	github.com/go-vgo/robotgo v0.110.1
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.3
	github.com/kbinani/screenshot v0.0.0-20230812210009-b87d31814237
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/orcaman/concurrent-map/v2 v2.0.1
	github.com/pkg/sftp v1.13.6
	github.com/sirupsen/logrus v1.9.3
	github.com/sourcegraph/jsonrpc2 v0.2.0
	github.com/stretchr/testify v1.10.0
	golang.org/x/crypto v0.36.0
	golang.org/x/net v0.38.0
	golang.org/x/sys v0.31.0
	golang.org/x/text v0.23.0
	gopkg.in/ini.v1 v1.67.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	dario.cat/mergo v1.0.1 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v1.0.0 // indirect
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/bytedance/sonic v1.12.6 // indirect
	github.com/bytedance/sonic/loader v0.2.1 // indirect
	github.com/cakturk/go-netstat v0.0.0-20200220111822-e5b49efee7a5 // indirect
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/cyphar/filepath-securejoin v0.2.4 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/gen2brain/shm v0.0.0-20230802011745-f2460f5984f7 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.5.1-0.20240427054813-8453aa90c6ec // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jezek/xgb v1.1.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20230326075908-cb1d2100619a // indirect
	github.com/lxn/win v0.0.0-20210218163916-a377121e959e // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/onsi/gomega v1.33.1 // indirect
	github.com/otiai10/gosseract v2.2.1+incompatible // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pjbgf/sha1cd v0.3.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20221212215047-62379fc7944b // indirect
	github.com/robotn/xgb v0.0.0-20190912153532-2cb92d044934 // indirect
	github.com/robotn/xgbutil v0.0.0-20190912154524-c861d6f87770 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	github.com/shirou/gopsutil/v3 v3.23.8 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/skeema/knownhosts v1.2.2 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/vcaesar/gops v0.30.2 // indirect
	github.com/vcaesar/imgo v0.40.0 // indirect
	github.com/vcaesar/keycode v0.10.1 // indirect
	github.com/vcaesar/tt v0.20.0 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	golang.org/x/arch v0.12.0 // indirect
	golang.org/x/image v0.12.0 // indirect
	google.golang.org/protobuf v1.36.3 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
)
