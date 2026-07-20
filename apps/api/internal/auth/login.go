package auth

import (
	"log"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

func (h *Handler) Login(
	w http.ResponseWriter,
	r *http.Request,
) {

	state, err := GenerateRandomValue(32)
	if err != nil {
		log.Printf(
			"generate OIDC state: %v",
			err,
		)

		http.Error(
			w,
			"could not start authentication",
			http.StatusInternalServerError,
		)
		return
	}

	nonce, err := GenerateRandomValue(32)
	if err != nil {
		log.Printf(
			"generate OIDC nonce: %v",
			err,
		)

		http.Error(
			w,
			"could not start authentication",
			http.StatusInternalServerError,
		)
		return
	}

	codeVerifier := oauth2.GenerateVerifier()

	now := h.now().UTC()

	transaction := LoginTransaction{
		State:        state,
		Nonce:        nonce,
		CodeVerifier: codeVerifier,
		ReturnTo:     h.resolveReturnTo(r),
		CreatedAt:    now,
		ExpiresAt:    now.Add(h.loginTransactionTTL),
	}

	if err := h.transactionStore.Save(
		r.Context(),
		transaction,
	); err != nil {
		log.Printf(
			"save OIDC login transaction: %v",
			err,
		)

		http.Error(
			w,
			"could not start authentication",
			http.StatusInternalServerError,
		)
		return
	}

	authorizationURL :=
		h.oidcClient.OAuth2Config.AuthCodeURL(
			state,
			oidc.Nonce(nonce),
			oauth2.S256ChallengeOption(codeVerifier),
		)

	http.Redirect(
		w,
		r,
		authorizationURL,
		http.StatusFound,
	)
}

func (h *Handler) resolveReturnTo(
	r *http.Request,
) string {
	requested := strings.TrimSpace(
		r.URL.Query().Get("return_to"),
	)

	if requested == "" {
		return h.defaultLoginDestination
	}

	if !isSafeLocalPath(requested) {
		return h.defaultLoginDestination
	}

	return requested
}
