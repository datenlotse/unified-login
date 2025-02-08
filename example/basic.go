package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	unified_login "github.com/datenlotse/unified-login-go"
)

var handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	userInfo, ok := r.Context().Value(unified_login.UserKey).(*unified_login.UserInformation)
	if !ok {
		// Only possible when using m.CheckJWT alone
		return
	}

	// Use the user data here
	js, _ := json.Marshal(userInfo)
	w.Write(js)
})

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scopes := []unified_login.Scope{
		{Scope: "scope_1", Description: "scope 1"},
		{Scope: "scope_2", Description: "scope 2"},
	}
	err := unified_login.SyncScopes(
		ctx,
		"https://login-microservice.local",
		"ba82da2be2ffb80ea0f12986f454e53acd6a81b891802307b48d129fb5cf4613",
		"951F48DE-CF08-EC11-945D-005056A2678B",
		scopes,
	)
	if err != nil {
		panic(err)
	}
	log.Println("Synced scopes")

	m := unified_login.NewMiddleware("<secret>")
	http.ListenAndServe("0.0.0.0:3000", m.CheckJWT(m.MustBeAuthenticated(handler)))
}
