package reposerver

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	mw "github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/tus/tusd/pkg/filestore"
	tusd "github.com/tus/tusd/pkg/handler"

	"go.amplifyedge.org/booty-v2/internal/osutil"
	"go.amplifyedge.org/booty-v2/internal/pass"
)

var tokenAuth *jwtauth.JWTAuth
var uploadUser string
var uploadPassword string

func init() {
	tokenAuth = jwtauth.New("HS512", []byte(os.Getenv("BOOTY_SESS_SECRET")), nil)
	uploadUser = os.Getenv("BOOTY_UPLOAD_USER")
	uploadPassword = os.Getenv("BOOTY_UPLOAD_PASSWORD")
}

type AuthRequest struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type AuthResponse struct {
}

type tusHandler struct {
	handler *tusd.Handler
	logger  *log.Logger
}

func NewServer() http.Handler {
	// create new router
	r := chi.NewRouter()
	r.Use(mw.RealIP)
	r.Use(mw.Logger)
	r.Use(mw.Recoverer)
	r.Use(mw.RequestID)

	logger := log.New(os.Stdout, "booty-repo", log.LstdFlags)

	th := createTusHandler(logger)

	// restricted route (upload)
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(mwAuth)
		r.Method(
			"POST",
			"/upload/*",
			th.handleUpload(),
		)
	})
	r.Post("/auth", handleAuth)
	r.Get("/dl/*", th.handleDownload)

	return r
}

func handleAuth(w http.ResponseWriter, r *http.Request) {
	var ar AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&ar); err != nil {
		http.Error(w, "request invalid", http.StatusBadRequest)
		return
	}

	if ar.User != uploadUser {
		http.Error(w, "invalid user", http.StatusForbidden)
		return
	}

	passOk, err := pass.VerifyHash(ar.Password, uploadPassword)
	if err != nil || !passOk {
		http.Error(w, "invalid password", http.StatusForbidden)
		return
	}
	_, tok, err := tokenAuth.Encode(map[string]interface{}{"user": ar.User})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	j, err := json.Marshal(&map[string]interface{}{"token": tok})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	w.Write(j)
}

func createTusHandler(l *log.Logger) *tusHandler {
	store := filestore.FileStore{
		Path: osutil.GetUploadDir(),
	}
	composer := tusd.NewStoreComposer()
	store.UseIn(composer)
	handler, err := tusd.NewHandler(tusd.Config{
		BasePath:                "/uploads/",
		NotifyCompleteUploads:   true,
		NotifyTerminatedUploads: true,
		StoreComposer:           composer,
		Logger:                  l,
		//PreUploadCreateCallback:   nil,
		//PreFinishResponseCallback: nil,
	})
	if err != nil {
		l.Fatalf("error creating upload handler: %v", err)
	}
	return &tusHandler{handler, l}
}

func (t *tusHandler) handleUpload() *tusd.Handler {
	eventHandler(t.logger, t.handler)
	return t.handler
}

func (t *tusHandler) handleDownload(w http.ResponseWriter, r *http.Request) {
	t.handler.GetFile(w, r)
}

func eventHandler(l *log.Logger, handler *tusd.Handler) {
	go func() {
		for {
			select {
			case info := <-handler.CompleteUploads:
				l.Printf("addr: %s, uri: %s,\n created: %v\n", info.HTTPRequest.RemoteAddr, info.HTTPRequest.URI, info.Upload)
			case info := <-handler.TerminatedUploads:
				l.Printf("addr: %s, uri: %s,\n terminated: %v\n", info.HTTPRequest.RemoteAddr, info.HTTPRequest.URI, info.Upload)
			}
		}
	}()
}

func mwAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _, err := tokenFromContext(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if token == nil || jwt.Validate(token) != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
	})
}

func tokenFromContext(ctx context.Context) (jwt.Token, map[string]interface{}, error) {
	token, _ := ctx.Value("Token").(jwt.Token)

	var err error
	var claims map[string]interface{}

	if token != nil {
		claims, err = token.AsMap(context.Background())
		if err != nil {
			return token, nil, err
		}
	} else {
		claims = map[string]interface{}{}
	}

	err, _ = ctx.Value("Err").(error)

	return token, claims, err
}
