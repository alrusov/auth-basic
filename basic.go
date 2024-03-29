package basic

import (
	"fmt"
	"net/http"

	"github.com/alrusov/auth"
	"github.com/alrusov/config"
	"github.com/alrusov/log"
	"github.com/alrusov/misc"
	"github.com/alrusov/stdhttp"
)

//----------------------------------------------------------------------------------------------------------------------------//

type (
	// AuthHandler --
	AuthHandler struct {
		http    *stdhttp.HTTP
		authCfg *config.Auth
		cfg     *config.AuthMethod
		options *methodOptions
	}

	methodOptions struct {
		HashedPassword bool `toml:"hashed-password"` // Пароль передается в хешированном виде
	}
)

const (
	module = "basic"
	method = "Basic"
)

//----------------------------------------------------------------------------------------------------------------------------//

// Автоматическая регистрация при запуске приложения
func init() {
	config.AddAuthMethod(module, &methodOptions{})
}

// Проверка валидности дополнительных опций метода
func (options *methodOptions) Check(cfg any) (err error) {
	msgs := misc.NewMessages()

	err = msgs.Error()
	return
}

//----------------------------------------------------------------------------------------------------------------------------//

// Init --
func (ah *AuthHandler) Init(cfg *config.Listener) (err error) {
	ah.authCfg = nil
	ah.cfg = nil
	ah.options = nil

	methodCfg, exists := cfg.Auth.Methods[module]
	if !exists || !methodCfg.Enabled || methodCfg.Options == nil {
		return nil
	}

	options, ok := methodCfg.Options.(*methodOptions)
	if !ok {
		return fmt.Errorf(`options for module "%s" is "%T", expected "%T"`, module, methodCfg.Options, options)
	}

	ah.authCfg = &cfg.Auth
	ah.cfg = methodCfg
	ah.options = options
	return nil
}

//----------------------------------------------------------------------------------------------------------------------------//

// Add --
func Add(http *stdhttp.HTTP) (err error) {
	return http.AddAuthHandler(
		&AuthHandler{
			http: http,
		},
	)
}

//----------------------------------------------------------------------------------------------------------------------------//

// Enabled --
func (ah *AuthHandler) Enabled() bool {
	return ah.cfg != nil && ah.cfg.Enabled
}

//----------------------------------------------------------------------------------------------------------------------------//

// Score --
func (ah *AuthHandler) Score() int {
	return ah.cfg.Score
}

//----------------------------------------------------------------------------------------------------------------------------//

// WWWAuthHeader --
func (ah *AuthHandler) WWWAuthHeader() (name string, withRealm bool) {
	return method, true
}

//----------------------------------------------------------------------------------------------------------------------------//

// Check --
func (ah *AuthHandler) Check(id uint64, prefix string, path string, w http.ResponseWriter, r *http.Request) (identity *auth.Identity, tryNext bool, err error) {
	if ah.cfg == nil || !ah.cfg.Enabled {
		return nil, true, nil
	}

	u, p, ok := r.BasicAuth()
	if !ok {
		return nil, true, nil
	}

	identity, _, err = auth.StdCheckUser(u, p, ah.options.HashedPassword)
	if err != nil {
		auth.Log.Message(log.INFO, `[%d] Basic login error: %s`, id, err)
		return nil, false, err
	}

	if identity == nil {
		auth.Log.Message(log.INFO, `[%d] Basic login error: user "%s" not found or illegal password`, id, u)
		return nil, false, fmt.Errorf(`user "%s" not found or illegal password`, u)
	}

	identity.Method = module
	return identity, false, nil
}

//----------------------------------------------------------------------------------------------------------------------------//
